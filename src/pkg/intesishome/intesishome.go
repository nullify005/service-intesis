package intesishome

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//go:embed "assets/StateMapping.json"
var mappingBody []byte

//go:embed "assets/ControlResponse.json"
var mockBody []byte

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
	err := json.Unmarshal(mappingBody, &ret)
	if err != nil {
		fmt.Printf("problem unmarshalling the statemapping when inspecting capabilities")
		os.Exit(1)
	}
	return
}

func capabilities(widgets []int) (caps []string) {
	var stateMapping map[string]interface{}
	err := json.Unmarshal(mappingBody, &stateMapping)
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
