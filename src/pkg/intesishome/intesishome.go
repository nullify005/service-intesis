package intesishome

import (
	"encoding/json"
	"fmt"
)

const (
	writeCommand  string = `{"command":"set","data":{"deviceId":%v,"uid":%v,"value":%v,"seqNo":0}}`
	writeAuth     string = `{"command":"connect_req","data":{"token":%v}}`
	commandConnOk string = "connect_rsp"
	commandSetAck string = "set_ack"
	commandAuthOk string = "ok"
)

type IntesisHome struct {
	username   string
	password   string
	hostname   string
	serverIP   string
	serverPort int
	token      int
}

type Option func(c *IntesisHome)

func New(user, pass string, opts ...Option) IntesisHome {
	c := IntesisHome{
		username: user,
		password: pass,
		hostname: _controlHostname,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// set an alternate hostname for API calls, useful for testing
func WithHostname(host string) Option {
	return func(c *IntesisHome) {
		c.hostname = host
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

	// setup a new token
	if _, err = controlRequest(ih); err != nil {
		return
	}

	// authenticate
	cmd := &CommandRequest{}
	cmd.Command = "connect_req"
	cmd.Data.Token = ih.token
	bytes, err := json.Marshal(cmd)
	if err != nil {
		return
	}
	resp := socketWrite(ih, bytes)
	ih.token = 0 // consume the token
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		return
	}
	if cmdResp.Command != commandConnOk && cmdResp.Data.Status != commandAuthOk {
		err = fmt.Errorf("expected ok auth response, got: %#v", cmdResp)
		return
	}

	// now write the command
	cmd = &CommandRequest{}
	cmd.Command = "set"
	cmd.Data.DeviceID = device
	cmd.Data.Uid = uid
	cmd.Data.Value = value
	cmd.Data.SeqNo = 0
	bytes, err = json.Marshal(cmd)
	if err != nil {
		return
	}
	resp = socketWrite(ih, bytes)
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		return
	}
	if cmdResp.Command != commandSetAck {
		err = fmt.Errorf("unable to perform requested set: %#v", cmdResp)
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

// func mockResponse(self *IntesisHome) ControlResponse {
// 	r := &ControlResponse{}
// 	err := json.Unmarshal(mockBody, &r)
// 	if err != nil {
// 		fmt.Printf("unable to marshal mock response body into ControlResponse: %v\n", err)
// 		os.Exit(1)
// 	}
// 	return *r
// }
