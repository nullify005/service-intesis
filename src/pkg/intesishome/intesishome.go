package intesishome

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	commandReqSet  string = "set"
	commandRspSet  string = "set_ack"
	commandReqCon  string = "connect_req"
	commandRspCon  string = "connect_rsp"
	commandAuthOk  string = "ok"
	commandAuthBad string = "err_token"
)

type IntesisHome struct {
	username   string
	password   string
	hostname   string
	serverIP   string
	serverPort int
	tcpServer  string
	token      int
	verbose    bool
	cmdSocket  net.Conn // holder for the tcpSocket
	mu         sync.Mutex
}

type Option func(c *IntesisHome)

func New(user, pass string, opts ...Option) *IntesisHome {
	c := IntesisHome{
		username: user,
		password: pass,
		hostname: DefaultHostname,
		verbose:  false,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}

// set an alternate hostname for API calls, useful for testing
func WithHostname(host string) Option {
	return func(ih *IntesisHome) {
		// set an appropriate default
		ih.hostname = DefaultHostname
		if host == "" {
			return
		}
		if !strings.Contains(host, "http://") {
			ih.hostname = "http://" + host
			return
		}
		ih.hostname = host
	}
}

// toggle the log verbosity
func WithVerbose(v bool) Option {
	return func(ih *IntesisHome) {
		ih.verbose = v
	}
}

// override the TCPServer for HVAC Control
func WithTCPServer(addr string) Option {
	return func(ih *IntesisHome) {
		ih.tcpServer = addr
	}
}

// lists the devices confgured within Intesis Home
func (ih *IntesisHome) Devices() (devices []Device, err error) {
	response, err := controlRequest(ih)
	if err != nil {
		return
	}
	for _, inst := range response.Config.Inst {
		devices = append(devices, inst.Devices...)
	}
	return
}

// checks to see whether a device is one that Intesis Home knows about
func (ih *IntesisHome) HasDevice(d int64) (found bool, err error) {
	found = false
	devices, err := ih.Devices()
	if err != nil {
		return
	}
	for _, dev := range devices {
		if fmt.Sprint(d) == dev.ID {
			found = true
			break
		}
	}
	return
}

// performs a change on a device using a uid & value
// mappings for parameter names to values should be conducted via MapCommand
// we reset & establish the connect here in order to have a single place
// to catch & reset any connection issues with the socket
// TODO: should there actually be some retries here?
func (ih *IntesisHome) Set(device int64, uid, value int) (err error) {
	// force a refresh of the token
	if _, err = controlRequest(ih); err != nil {
		return
	}

	if ih.cmdSocket != nil {
		ih.cmdSocket.Close()
		ih.cmdSocket = nil
	}
	ih.cmdSocket, err = net.Dial("tcp", fmt.Sprintf("%s:%v", ih.serverIP, ih.serverPort))
	if err != nil {
		return
	}

	ih.cmdSocket.SetDeadline(time.Now().Add(_socketReadTimeout))
	err = setHandler(ih, device, uid, value)
	ih.cmdSocket.Close()
	ih.cmdSocket = nil
	return
}

// the inner handler for Set
// expects that the tcp socket has been reset / cleaned for work
func setHandler(ih *IntesisHome, device int64, uid, value int) (err error) {
	var cmdResp CommandResponse

	// authenticate
	cmd := &CommandRequest{
		Command: commandReqCon,
		Data: CommandRequestData{
			Token: ih.token,
		},
	}
	bytes, err := json.Marshal(cmd)
	if err != nil {
		e := fmt.Errorf("auth command encode error. auth: %v cause: %v", cmd, err)
		return e
	}
	resp, err := socketWriteRead(ih, bytes)
	if err != nil {
		e := fmt.Errorf("auth write error. auth: %v cause: %v", cmd, err)
		return e
	}
	ih.token = 0 // consume the token
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		e := fmt.Errorf("auth response decode error. resp: %s caused: %v", string(resp), err)
		return e
	}
	if cmdResp.Command != commandRspCon {
		err = fmt.Errorf("unexpected auth reply. expected: %s got: %s", commandRspCon, cmdResp.Command)
		return
	}
	if cmdResp.Data.Status != commandAuthOk {
		err = fmt.Errorf("unexpected auth reply. expected: %s got: %s", commandAuthOk, cmdResp.Data.Status)
		return
	}

	// now write the command
	cmd = &CommandRequest{
		Command: commandReqSet,
		Data: CommandRequestData{
			DeviceID: device,
			Uid:      uid,
			Value:    value,
			SeqNo:    0,
		},
	}
	bytes, err = json.Marshal(cmd)
	if err != nil {
		e := fmt.Errorf("set command encode error. cmd: %v cause: %v", cmd, err)
		return e
	}
	resp, err = socketWriteRead(ih, bytes)
	if err != nil {
		e := fmt.Errorf("set command write error. cmd: %v cause: %v", cmd, err)
		return e
	}
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		e := fmt.Errorf("set command response error. cmd: %v response: %s cause: %v", cmd, string(resp), err)
		return e
	}
	if cmdResp.Command != commandRspSet {
		err = fmt.Errorf("set command failed. cmd: %v expected: %s got: %s", cmd, commandRspSet, cmdResp.Command)
		return
	}
	return
}

// contacts the Intesis Home API to obtain the status of a device
func (ih *IntesisHome) Status(device int64) (state map[string]interface{}, err error) {
	state = make(map[string]interface{})
	response, err := controlRequest(ih)
	if err != nil {
		return
	}
	for _, s := range response.Status.Status {
		if s.DeviceID != device {
			continue
		}
		state[DecodeUid(s.UID)] = s.Value
	}
	return
}

func (ih *IntesisHome) Token() (int, error) {
	var token int
	_, err := controlRequest(ih)
	if err != nil {
		return token, err
	}
	return ih.token, err
}

func (ih *IntesisHome) Controller() string {
	return fmt.Sprintf("%s:%d", ih.serverIP, ih.serverPort)
}
