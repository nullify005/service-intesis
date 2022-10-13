package intesishome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	controlHostname    string        = "https://user.intesishome.com"
	ControlEndpoint    string        = "/api.php/get/control"
	_statusCommand     string        = `{"status":{"hash":"x"},"config":{"hash":"x"}}`
	_statusVersion     string        = "1.8.5"
	_readLimitBytes    int           = 1024
	_socketReadTimeout time.Duration = 30 * time.Second
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
		DeviceID int64  `json:"devceId,omitempty"`
		SeqNo    int    `json:"seqNo,omitempty"`
		Rssi     int    `json:"rssi,omitempty"`
		Status   string `json:"status,omitempty"`
	} `json:"data"`
}

type CommandRequest struct {
	Command string             `json:"command"`
	Data    CommandRequestData `json:"data"`
}

type CommandRequestData struct {
	DeviceID int64 `json:"deviceId"`
	Uid      int   `json:"uid"`
	Value    int   `json:"value"`
	SeqNo    int   `json:"seqNo"`
	Token    int   `json:"token"`
}

func controlRequest(ih *IntesisHome) (r ControlResponse, err error) {
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
	// override the TCPServer settings
	if ih.tcpServer != "" {
		addr := strings.Split(ih.tcpServer, ":")
		ih.serverIP = addr[0]
		ih.serverPort, err = strconv.Atoi(addr[1])
		if err != nil {
			return
		}
	}
	if ih.verbose {
		fmt.Printf("DEBUG|controlRequest| token: %v server: %s %v\n", ih.token, ih.serverIP, ih.serverPort)
	}
	return r, err
}

func socketWrite(ih *IntesisHome, b []byte) (response []byte, err error) {
	if socket == nil {
		socket, err = net.Dial("tcp", fmt.Sprintf("%s:%v", ih.serverIP, ih.serverPort))
		if err != nil {
			return
		}
		socket.SetDeadline(time.Now().Add(_socketReadTimeout))
	}
	if ih.verbose {
		fmt.Printf("DEBUG|socketWrite| sending request: %s\n", string(b))
	}
	wBytes, err := socket.Write(b)
	if err != nil {
		return
	}
	if wBytes != len(b) {
		err = fmt.Errorf("read byte mismatch, expected: %v actual: %v", wBytes, len(b))
		return
	}
	response = make([]byte, _readLimitBytes)
	_, err = socket.Read(response)
	if err != nil {
		return
	}
	if ih.verbose {
		fmt.Printf("DEBUG|socketWrite| received response: %s\n", response)
	}
	response = bytes.Trim(response, "\x00") // trim any nulls from the end
	return
}

func statusForm(user, pass string) (ret url.Values) {
	ret = url.Values{}
	ret.Set("username", user)
	ret.Add("password", pass)
	ret.Add("version", _statusVersion)
	ret.Add("cmd", _statusCommand)
	return
}