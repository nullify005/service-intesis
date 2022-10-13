package intesishome

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	_commandReqSet string = "set"
	_commandRspSet string = "set_ack"
	_commandReqCon string = "connect_req"
	_commandRspCon string = "connect_rsp"
	_commandAuthOk string = "ok"
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
}

type Option func(c *IntesisHome)

func New(user, pass string, opts ...Option) IntesisHome {
	c := IntesisHome{
		username: user,
		password: pass,
		hostname: controlHostname,
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
		Command: _commandReqCon,
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
	if cmdResp.Command != _commandRspCon && cmdResp.Data.Status != _commandAuthOk {
		err = fmt.Errorf("unexpected reply, expected: %s/%s got: %s/%s",
			_commandRspCon, _commandAuthOk, cmdResp.Command, cmdResp.Data.Status)
		return
	}

	// now write the command
	cmd = &CommandRequest{
		Command: _commandReqSet,
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
	if cmdResp.Command != _commandRspSet {
		err = fmt.Errorf("set failed, expected: %s got: %s", _commandRspSet, cmdResp.Command)
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
