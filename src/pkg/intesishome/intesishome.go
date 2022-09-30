package intesishome

import (
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

//go:embed "assets/CommandMapping.json"
var commandMap []byte

//go:embed "assets/StateMapping.json"
var statusMap []byte

//go:embed "assets/ControlResponse.json"
var mockBody []byte

var socket net.Conn // holder for the TCP session

func (c *Connection) Set(device int64, uid, value int) {
	_ = callControl(c) // ensure that the connection is initialised for a new token
	cmd := []byte(fmt.Sprintf(WriteAuth, c.Token))
	socketWrite(c, cmd)
	c.Token = -1 // consume the token
	// TODO: validate the response
	// {"command":"connect_rsp","data":{"status":"ok"}}

	// now write the command
	cmd = []byte(fmt.Sprintf(WriteCommand, device, uid, value))
	socketWrite(c, cmd)
	// TODO: validate the response
	// {"command":"set_ack","data":{"deviceId":127934703953,"seqNo":0,"rssi":195}}
}

func (c *Connection) Status(device int64) (state map[string]interface{}) {
	state = make(map[string]interface{})
	response := callControl(c)
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

func (c *Connection) MapValue(name string, value int) interface{} {
	mapping := mappings()
	for k := range mapping {
		if mapping[k].(map[string]interface{})["name"].(string) == name {
			values, ok := mapping[k].(map[string]interface{})["values"].(map[string]interface{})
			if !ok {
				// there's no human mapping for the value
				return value
			}
			return values[fmt.Sprint(value)]
		}
	}
	// couldn't find it give back the value
	return value
}

func (c *Connection) Devices() []Device {
	devices := []Device{}
	response := callControl(c)
	for _, inst := range response.Config.Inst {
		devices = append(devices, inst.Devices...)
	}
	return devices
}

func (d *Device) String() (s string) {
	s = fmt.Sprintf("device id: %v name: %v family: %v model: %v capabilities [%v]", d.ID, d.Name, d.FamilyID, d.ModelID, strings.Join(capabilities(d.Widgets), ","))
	return
}

func mappings() (ret map[string]interface{}) {
	err := json.Unmarshal(statusMap, &ret)
	if err != nil {
		fmt.Printf("problem unmarshalling the statemapping when inspecting capabilities")
		os.Exit(1)
	}
	return
}

func capabilities(widgets []int) (caps []string) {
	var stateMapping map[string]interface{}
	err := json.Unmarshal(statusMap, &stateMapping)
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

func callControl(self *Connection) ControlResponse {
	if self.Endpoint == "" {
		self.Endpoint = ApiEndpoint
	}
	controlResponse := &ControlResponse{}
	if self.Mock {
		return mockResponse(self)
	}
	form := statusForm(self.Username, self.Password)
	resp, err := http.PostForm(self.Endpoint, form)
	if err != nil {
		fmt.Printf("problem contacting endpoint: %v with: %v\n", self.Endpoint, err)
		os.Exit(1)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("problem reading response body: %s with: %v\n", resp.Body, err)
		os.Exit(1)
	}
	if string(body) == "[]" {
		fmt.Printf("response body was empty?")
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("received status code: %v with reason: %s\n", resp.StatusCode, body)
		os.Exit(1)
	}
	err = json.Unmarshal(body, &controlResponse)
	if err != nil {
		fmt.Printf("unable to marshal response body into ControlResponse: %v body: %s\n", err, body)
		os.Exit(1)
	}
	self.Token = controlResponse.Config.Token
	self.ServerIP = controlResponse.Config.ServerIP
	self.ServerPort = controlResponse.Config.ServerPort
	return *controlResponse
}

func mockResponse(self *Connection) ControlResponse {
	r := &ControlResponse{}
	err := json.Unmarshal(mockBody, &r)
	if err != nil {
		fmt.Printf("unable to marshal mock response body into ControlResponse: %v\n", err)
		os.Exit(1)
	}
	return *r
}

func statusForm(user, pass string) (ret url.Values) {
	ret = url.Values{}
	ret.Set("username", user)
	ret.Add("password", pass)
	ret.Add("version", StatusVersion)
	ret.Add("cmd", StatusCommand)
	return
}

func socketWrite(c *Connection, b []byte) {
	var err error
	if socket == nil {
		socket, err = net.Dial("tcp", fmt.Sprintf("%s:%v", c.ServerIP, c.ServerPort))
		if err != nil {
			fmt.Printf("unable to establish connection to: %v:%v with: %v\n", c.ServerIP, c.ServerPort, err.Error())
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
	response := make([]byte, 1024)
	_, err = socket.Read(response)
	if err != nil {
		fmt.Printf("failure to read from socket with: %v\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("received response: %s\n", response)
}

func (c *Connection) MapCommand(key string, value interface{}) (uid, mValue int) {
	var commands map[string]interface{}
	err := json.Unmarshal(commandMap, &commands)
	if err != nil {
		fmt.Printf("unable to unmarshal command map: %v\n", err.Error())
		os.Exit(1)
	}
	_, ok := commands[key]
	if !ok {
		fmt.Printf("no such key: %s exists in the command map", key)
		os.Exit(1)
	}
	if i, err := strconv.Atoi(key); err == nil {
		// it's an int already
		uid = i
	} else {
		// map the key to the uid
		uid = int(commands[key].(map[string]interface{})["uid"].(float64))
	}
	if i, err := strconv.Atoi(value.(string)); err == nil {
		// it's an int so pass it back
		mValue = i
		return
	}
	// otherwise we have to map it
	values := commands[key].(map[string]interface{})["values"].(map[string]interface{})
	if _, ok := values[value.(string)]; !ok {
		fmt.Printf("no such value: %v exists for command: %v wanted: %v", value, key, values)
		os.Exit(1)
	}
	mValue = int(values[value.(string)].(float64))
	return
}
