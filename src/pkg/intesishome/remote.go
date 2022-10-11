package intesishome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
)

const (
	_controlHostname string = "https://user.intesishome.com"
	controlEndpoint  string = "/api.php/get/control"
	statusCommand    string = `{"status":{"hash":"x"},"config":{"hash":"x"}}`
	statusVersion    string = "1.8.5"
)

var (
	socket net.Conn // holder for the TCP session
)

type ControlResponse struct {
	Config struct {
		Token          int     `json:"token"`
		PushToken      string  `json:"pushToken"`
		LastAppVersion string  `json:"lastAppVersion"`
		ForceUpdate    int     `json:"forceUpdate"`
		SetDelay       float64 `json:"setDelay"`
		ServerIP       string  `json:"serverIP"`
		ServerPort     int     `json:"serverPort"`
		Hash           string  `json:"hash"`
		Inst           []struct {
			ID      int      `json:"id"`
			Order   int      `json:"order"`
			Name    string   `json:"name"`
			Devices []Device `json:"devices"`
		} `json:"inst"`
	} `json:"config"`
	Status struct {
		Hash   string `json:"hash"`
		Status []struct {
			DeviceID int64 `json:"deviceId"`
			UID      int   `json:"uid"`
			Value    int   `json:"value"`
			Name     string
		} `json:"status"`
	} `json:"status"`
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type CommandResponse struct {
	Command string `json:"command"`
	Data    struct {
		DeviceID int64  `json:"devceId"`
		SeqNo    int    `json:"seqNo"`
		Rssi     int    `json:"rssi"`
		Status   string `json:"status"`
	} `json:"data"`
}

type CommandRequest struct {
	Command string `json:"command"`
	Data    struct {
		DeviceID int64 `json:"deviceId,omitempty"`
		Uid      int   `json:"uid,omitempty"`
		Value    int   `json:"value,omitempty"`
		SeqNo    int   `json:"seqNo,omitempty"`
		Token    int   `json:"token,omitempty"`
	} `json:"data"`
}

func controlRequest(ih *IntesisHome) (r ControlResponse, err error) {
	// controlResponse := &ControlResponse{}
	// if self.mock {
	// 	return mockResponse(self)
	// }
	form := statusForm(ih.username, ih.password)
	uri := ih.hostname + controlEndpoint
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
	ret.Add("version", statusVersion)
	ret.Add("cmd", statusCommand)
	return
}
