package intesishome

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	mockDeviceId                string = "127934703953"
	nilResponsePayload          string = "./assets/tests/nilResponse.json"
	validControlResponsePayload string = "./assets/tests/validControlResponse.json"
	errorControlResponsePayload string = "./assets/tests/noMethodResponse.json"
	malformedResponsePayload    string = "./assets/tests/malformedControlResponse.txt"
)

var (
	stateVerifyMapped = map[string]interface{}{
		"alarm_status": 0,
		"mode":         "heat",
		"power":        "on",
		"rssi":         200,
		"setpoint":     200,
		"setpoint_max": 300,
		"setpoint_min": 160,
		"temperature":  210,
	}
	stateVerifyUnmapped = map[string]interface{}{
		"alarm_status": 0,
		"mode":         1,
		"power":        1,
		"rssi":         200,
		"setpoint":     200,
		"setpoint_max": 300,
		"setpoint_min": 160,
		"temperature":  210,
	}
	cmdAuthOk = []byte(`{"command":"connect_rsp","data":{"status":"ok"}}`)
	cmdSetOk  = []byte(`{"command":"set_ack","data":{"deviceId":127934703953,"seqNo":0,"rssi":192}}`)
	shutdown  = make(chan bool, 1)
)

type inlineControlCheck func(t *testing.T, r *ControlResponse, e error)

func TestAPICalls(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		payload string
		want    inlineControlCheck
	}{
		{
			"non 200 response",
			http.StatusBadGateway,
			nilResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "unexpected response code")
			},
		},
		{
			"nil payload",
			http.StatusOK,
			nilResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "unexpected response body")
			},
		},
		{
			"error payload",
			http.StatusOK,
			errorControlResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "unexpected response error")
			},
		},
		{
			"malformed payload",
			http.StatusOK,
			malformedResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "malformed payload")
			},
		},
		{
			"success payload",
			http.StatusOK,
			validControlResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.NoError(t, e)
				assert.Equal(t, r.ErrorCode, 0)
				assert.Equal(t, r.Config.Inst[0].Devices[0].ID, mockDeviceId)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := testCallControl(tt.code, tt.payload)
			tt.want(t, &r, err)
		})
	}
}

type inlineDevicesCheck func(t *testing.T, d []Device, e error)

func TestDevices(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		payload string
		want    inlineDevicesCheck
	}{
		{
			"valid response",
			http.StatusOK,
			validControlResponsePayload,
			func(t *testing.T, d []Device, e error) {
				assert.NoError(t, e)
				assert.Equal(t, len(d), 1)
				assert.Equal(t, d[0].ID, mockDeviceId)
			},
		},
		{
			"invalid response",
			http.StatusBadGateway,
			nilResponsePayload,
			func(t *testing.T, d []Device, e error) {
				assert.Error(t, e)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := mockHTTPServer(tt.code, tt.payload)
			if err != nil {
				t.Errorf("mock http server problem: %v", err.Error())
				return
			}
			ih := New("u", "p", WithHostname(s.URL))
			devices, err := ih.Devices()
			tt.want(t, devices, err)
		})
	}
}

type inlineStatusCheck func(t *testing.T, s map[string]interface{}, e error)

func TestStatus(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		payload string
		want    inlineStatusCheck
	}{
		{
			"valid response",
			http.StatusOK,
			validControlResponsePayload,
			func(t *testing.T, s map[string]interface{}, e error) {
				assert.NoError(t, e)
				for k, v := range stateVerifyUnmapped {
					if _, ok := s[k]; !ok {
						t.Errorf("%s was missing from state map", k)
						continue
					}
					assert.Equal(t, s[k], v)
					mVal := MapState(k, v.(int))
					assert.Equal(t, stateVerifyMapped[k], mVal)
				}
			},
		},
		{
			"invalid response",
			http.StatusOK,
			errorControlResponsePayload,
			func(t *testing.T, s map[string]interface{}, e error) {
				assert.Error(t, e)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := mockHTTPServer(tt.code, tt.payload)
			if err != nil {
				t.Errorf("mock http server problem: %v", err.Error())
				return
			}
			ih := New("u", "p", WithHostname(s.URL))
			d, _ := strconv.Atoi(mockDeviceId)
			state, err := ih.Status(int64(d))
			tt.want(t, state, err)
		})
	}
}

func TestStateMapping(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value int
		want  interface{}
	}{
		{
			"valid response",
			"power",
			1,
			"on",
		},
		{
			"invalid key",
			"unknown",
			65535,
			65535,
		},
		{
			"invalid value",
			"power",
			-1,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mVal := MapState(tt.key, tt.value)
			assert.Equal(t, tt.want, mVal)
		})
	}
}

// TODO: write tests for the command push & command mappings

// func viaTestServer() (c *IntesisHome) {
// 	c = IntesisHome(testServer.URL)
// 	return
// }

// func IntesisHome(endpoint string) (c *IntesisHome) {
// 	c = &IntesisHome{Username: "a", Password: "b", Endpoint: endpoint + mockContext}
// 	return
// }

func mockHTTPServer(responseCode int, payloadFile string) (s *httptest.Server, err error) {
	body, err := os.ReadFile(payloadFile)
	if err != nil {
		return
	}
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseCode)
		w.Write(body)
	}))
	return
}

func testCallControl(responseCode int, payload string) (r ControlResponse, err error) {
	s, err := mockHTTPServer(responseCode, payload)
	if err != nil {
		fmt.Printf("mock http server problem: %v", err.Error())
		return
	}
	ih := New("u", "p", WithHostname(s.URL))
	r, err = callControl(&ih)
	return
}

// func mockTCPServer() {
// 	listener, err := net.Listen("tcp", listenAddr+":"+listenPort)
// 	if err != nil {
// 		fmt.Printf("unable to open socket %s:%s with: %v\n", listenAddr, listenPort, err.Error())
// 		os.Exit(1)
// 	}
// 	defer socket.Close()
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			fmt.Printf("error accepting IntesisHome: %v\n", err.Error())
// 			os.Exit(1)
// 		}
// 		select {
// 		case <-shutdown:
// 			return
// 		default:
// 			// continue reading accepting IntesisHomes
// 		}
// 		go handleTCPRequest(conn)
// 	}
// }

// func handleTCPRequest(conn net.Conn) {
// 	request := make([]byte, 1024)
// 	_, err := conn.Read(request)
// 	if err != nil {
// 		fmt.Printf("error reading request: %v\n", err.Error())
// 		os.Exit(1)
// 	}
// }
