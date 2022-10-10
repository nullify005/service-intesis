package intesishome

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

//go:embed "assets/mappingCommand.json"
var mappingCommand []byte

//go:embed "assets/mappingState.json"
var mappingState []byte

var socket net.Conn // holder for the TCP session

type Option func(c *IntesisHome)

func WithHostname(host string) Option {
	return func(c *IntesisHome) {
		c.hostname = host
	}
}

func New(user, pass string, opts ...Option) IntesisHome {
	c := IntesisHome{
		username: user,
		password: pass,
		hostname: _defaultHostname,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func (ih *IntesisHome) Devices() (devices []Device, err error) {
	response, err := callControl(ih)
	if err != nil {
		return
	}
	for _, inst := range response.Config.Inst {
		devices = append(devices, inst.Devices...)
	}
	return
}

func MapCommand(key string, value interface{}) (uid, mValue int, err error) {
	var commands map[string]interface{}
	if err = json.Unmarshal(mappingCommand, &commands); err != nil {
		return
	}
	if _, ok := commands[key]; !ok {
		err = fmt.Errorf("key not present in command map: %s", key)
		return
	}
	if i, err := strconv.Atoi(key); err == nil {
		// it's an int already
		uid = i
	} else {
		// map the key to the uid
		uid = int(commands[key].(map[string]interface{})["uid"].(float64))
	}
	i, err := strconv.Atoi(value.(string))
	if err == nil {
		// it's an int so pass it back
		mValue = i
		return
	}
	// otherwise we have to map it, reset the err
	err = nil
	values := commands[key].(map[string]interface{})["values"].(map[string]interface{})
	if _, ok := values[value.(string)]; !ok {
		err = fmt.Errorf("no such value: %v exists for command: %v wanted: %v", value, key, values)
		return
	}
	mValue = int(values[value.(string)].(float64))
	return
}

// returns a string representation of the value, the original value if it cannot be mapped or nil
func MapState(name string, value int) interface{} {
	mapping := mappings()
	for k := range mapping {
		if mapping[k].(map[string]interface{})["name"].(string) == name {
			values, ok := mapping[k].(map[string]interface{})["values"].(map[string]interface{})
			if !ok {
				// there's no human mapping for the value
				return value
			}
			if _, ok := values[fmt.Sprint(value)]; ok {
				return values[fmt.Sprint(value)]
			}
			// there's no mapping for this value, return nil
			return nil
		}
	}
	// couldn't find it give back the value
	return value
}

func (ih *IntesisHome) Set(device int64, uid, value int) (err error) {
	var cmdResp CommandResponse

	// setup a new token
	if _, err = callControl(ih); err != nil {
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
	if cmdResp.Command != CommandConnOk && cmdResp.Data.Status != CommandAuthOk {
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
	if cmdResp.Command != CommandSetAck {
		err = fmt.Errorf("unable to perform requested set: %#v", cmdResp)
		return
	}
	return
}

func (ih *IntesisHome) Status(device int64) (state map[string]interface{}, err error) {
	state = make(map[string]interface{})
	response, err := callControl(ih)
	if err != nil {
		return
	}
	mapping := mappings()
	for _, s := range response.Status.Status {
		if s.DeviceID != device {
			continue
		}
		uid := fmt.Sprint(s.UID)
		_, ok := mapping[uid]
		if !ok {
			// key doesn't exist
			continue
		}
		name := mapping[uid].(map[string]interface{})["name"].(string)
		state[name] = s.Value
	}
	return
}

func (d *Device) String() (s string) {
	s = fmt.Sprintf("device id: %v name: %v family: %v model: %v capabilities [%v]", d.ID, d.Name, d.FamilyID, d.ModelID, strings.Join(capabilities(d.Widgets), ","))
	return
}

func capabilities(widgets []int) (caps []string) {
	var stateMapping map[string]interface{}
	err := json.Unmarshal(mappingState, &stateMapping)
	if err != nil {
		fmt.Printf("problem unmarshalling the statemapping when inspecting capabilities")
		os.Exit(1)
	}
	for _, v := range widgets {
		uid := fmt.Sprint(v)
		_, ok := stateMapping[uid]
		if !ok {
			// mapping doesn't exist
			fmt.Printf("widget: %v has no mapping?\n", v)
			continue
		}
		name := stateMapping[uid].(map[string]interface{})["name"].(string)
		caps = append(caps, name)
	}
	return
}

func callControl(ih *IntesisHome) (r ControlResponse, err error) {
	// controlResponse := &ControlResponse{}
	// if self.mock {
	// 	return mockResponse(self)
	// }
	form := statusForm(ih.username, ih.password)
	uri := ih.hostname + ControlEndpoint
	resp, err := http.PostForm(uri, form)
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected response code: %v body: %s", resp.StatusCode, body)
		return
	}
	if len(body) < 10 {
		err = fmt.Errorf("unexpected response body: %s", string(body))
		return
	}
	if err = json.Unmarshal(body, &r); err != nil {
		err = fmt.Errorf("malformed payload: %v", err.Error())
		return
	}
	if r.ErrorCode != 0 {
		err = fmt.Errorf("unexpected response error: %v message: %v", r.ErrorCode, r.ErrorMessage)
		return
	}
	ih.token = r.Config.Token
	ih.serverIP = r.Config.ServerIP
	ih.serverPort = r.Config.ServerPort
	return r, err
}

func mappings() (ret map[string]interface{}) {
	err := json.Unmarshal(mappingState, &ret)
	if err != nil {
		fmt.Printf("problem unmarshalling the statemapping when inspecting capabilities")
		os.Exit(1)
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

func socketWrite(c *IntesisHome, b []byte) (response []byte) {
	var err error
	if socket == nil {
		socket, err = net.Dial("tcp", fmt.Sprintf("%s:%v", c.serverIP, c.serverPort))
		if err != nil {
			fmt.Printf("unable to establish IntesisHome to: %v:%v with: %v\n", c.serverIP, c.serverPort, err.Error())
			os.Exit(1)
		}
	}
	res, err := socket.Write(b)
	if err != nil {
		fmt.Printf("unable to write: %s to socket with: %v\n", b, err.Error())
		os.Exit(1)
	}
	if res != len(b) {
		fmt.Printf("not all bytes of: %s were written: %v vs. %v\n", b, res, len(b))
		os.Exit(1)
	}
	response = make([]byte, 1024)
	_, err = socket.Read(response)
	if err != nil {
		fmt.Printf("failure to read from socket with: %v\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("received response: %s\n", response)
	response = bytes.Trim(response, "\x00") // trim the nulls from the end
	return
}

func statusForm(user, pass string) (ret url.Values) {
	ret = url.Values{}
	ret.Set("username", user)
	ret.Add("password", pass)
	ret.Add("version", StatusVersion)
	ret.Add("cmd", StatusCommand)
	return
}
