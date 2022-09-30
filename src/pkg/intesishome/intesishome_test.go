package intesishome

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const mockContext string = "/api.php/get/control"
const mockDeviceId int64 = 127934703953

//go:embed "assets/ControlResponse.json"
var responseBody []byte
var (
	testServer        *httptest.Server
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
)

func TestMain(m *testing.M) {
	testServer = mockServer()
	defer testServer.Close()
	code := m.Run()
	os.Exit(code)
}

func TestDevices(t *testing.T) {
	ih := viaTestServer()
	devices := ih.Devices()
	if len(devices) < 1 {
		t.Errorf("devices was empty!")
	}
	for _, d := range devices {
		if d.ID != fmt.Sprint(mockDeviceId) {
			t.Errorf("wanted: %v got: %v", d.ID, mockDeviceId)
		}
	}
}

func TestMapping(t *testing.T) {
	ih := viaTestServer()
	for k, v := range stateVerifyUnmapped {
		mappedV := ih.MapValue(k, v.(int))
		if mappedV != stateVerifyMapped[k] {
			t.Errorf("wanted: %v got %v", mappedV, stateVerifyMapped[k])
		}
	}
}

func TestState(t *testing.T) {
	ih := viaTestServer()
	state := ih.Status(mockDeviceId)
	for k, v := range stateVerifyUnmapped {
		_, ok := state[k]
		if !ok {
			t.Errorf("missing key: %v in state", k)
			continue
		}
		if state[k] != v {
			t.Errorf("unexpected value for: %v wanted: %v got %v", k, v, state[k])
		}
	}
}

func viaTestServer() (c *Connection) {
	c = connection(testServer.URL)
	return
}

func connection(endpoint string) (c *Connection) {
	c = &Connection{Username: "a", Password: "b", Endpoint: endpoint + mockContext}
	return
}

func mockServer() (s *httptest.Server) {
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != mockContext {
			fmt.Printf("expected to request: %s got: %s", mockContext, r.URL.Path)
			os.Exit(1)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}))
	return
}
