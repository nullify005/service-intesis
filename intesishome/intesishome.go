package intesishome

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func (c *Connection) Status(device int64) (state map[string]interface{}) {
	state = make(map[string]interface{})
	response := callControl(c)
	var stateMapping map[string]interface{}
	d, _ := os.Getwd()
	body, _ := os.ReadFile(d + "/intesishome/" + mappingPayload)
	_ = json.Unmarshal(body, &stateMapping)
	for _, s := range response.Status.Status {
		if s.DeviceID != device {
			continue
		}
		uid := fmt.Sprint(s.UID)
		_, ok := stateMapping[uid]
		if !ok {
			// key doesn't exist
			continue
		}
		name := stateMapping[uid].(map[string]interface{})["name"].(string)
		_, ok = stateMapping[uid].(map[string]interface{})["values"]
		if !ok {
			// there's no human mapping for the value
			state[name] = s.Value
			continue
		}
		key := fmt.Sprint(s.Value)
		values := stateMapping[uid].(map[string]interface{})["values"].(map[string]interface{})
		state[name] = values[key]
		// fmt.Printf("%v: %v\n", name, s.Value)
	}
	return
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

func (d *DeviceState) String() (s string) {
	var m map[string]interface{}
	data, _ := json.Marshal(d)
	json.Unmarshal(data, &m)
	for k, v := range m {
		s += fmt.Sprintf("%s: %.1f ", k, v)
	}
	return
}

func capabilities(widgets []int) (caps []string) {
	for _, v := range widgets {
		caps = append(caps, Capability[v])
	}
	return
}

func callControl(self *Connection) ControlResponse {
	controlResponse := &ControlResponse{}
	if self.Mock != "" {
		return mockResponse(self)
	}
	form := statusForm(self.Username, self.Password)
	resp, err := http.PostForm(ApiEndpoint, form)
	if err != nil {
		fmt.Printf("problem contacting endpoint: %v with: %v\n", ApiEndpoint, err)
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
	return *controlResponse
}

func mockResponse(self *Connection) ControlResponse {
	r := &ControlResponse{}
	body, err := os.ReadFile(self.Mock)
	if err != nil {
		fmt.Printf("unable to read in mock response: %v with: %v\n", self.Mock, err)
		os.Exit(1)
	}
	err = json.Unmarshal(body, &r)
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
