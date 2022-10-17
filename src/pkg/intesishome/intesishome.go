package intesishome

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
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
}

type Option func(c *IntesisHome)

func New(user, pass string, opts ...Option) IntesisHome {
	c := IntesisHome{
		username: user,
		password: pass,
		hostname: DefaultHostname,
		verbose:  false,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
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
func (ih *IntesisHome) Set(device int64, uid, value int) (err error) {
	var cmdResp CommandResponse

	// setup a new token if there is none
	if ih.token == 0 {
		if _, err = controlRequest(ih); err != nil {
			return
		}
	}

	// authenticate
	cmd := &CommandRequest{
		Command: commandReqCon,
		Data: CommandRequestData{
			Token: ih.token,
		},
	}
	bytes, err := json.Marshal(cmd)
	if err != nil {
		return
	}
	resp, err := socketWrite(ih, bytes)
	if err != nil {
		return
	}
	ih.token = 0 // consume the token
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		return
	}
	if cmdResp.Command != commandRspCon {
		err = fmt.Errorf("unexpected reply, expected: %s got: %s", commandRspCon, cmdResp.Command)
		return
	}
	if cmdResp.Data.Status != commandAuthOk {
		err = fmt.Errorf("unexpected reply, expected: %s got: %s", commandAuthOk, cmdResp.Data.Status)
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
		return
	}
	resp, err = socketWrite(ih, bytes)
	if err != nil {
		return
	}
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		return
	}
	if cmdResp.Command != commandRspSet {
		err = fmt.Errorf("set failed, expected: %s got: %s", commandRspSet, cmdResp.Command)
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
