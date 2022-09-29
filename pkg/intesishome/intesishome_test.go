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
var stateVerification = map[string]interface{}{
	"alarm_status":                0,
	"config_confirm_off":          0,
	"config_mode_map":             31,
	"config_quiet":                0,
	"config_vertical_vanes":       1054,
	"cool_temperature_max":        280,
	"cool_temperature_min":        240,
	"error_address":               65535,
	"error_code":                  0,
	"external_led":                "on",
	"fan_speed":                   0,
	"filter_clean":                0,
	"filter_due_hours":            0,
	"heat_temperature_min":        230,
	"internal_led":                "off",
	"internal_temperature_offset": 0,
	"mainenance_w_reset":          0,
	"mainenance_wo_reset":         0,
	"mode":                        "heat",
	"power":                       "on",
	"quiet_mode":                  "off",
	"rssi":                        200,
	"setpoint":                    200,
	"setpoint_max":                300,
	"setpoint_min":                160,
	"temp_limitation":             "off",
	"temperature":                 210,
}

func TestDevices(t *testing.T) {
	server := mockServer()
	defer server.Close()

	ih := &Connection{Username: "user", Password: "pass", Endpoint: server.URL + mockContext}
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

func TestState(t *testing.T) {
	server := mockServer()
	defer server.Close()

	ih := &Connection{Username: "user", Password: "pass", Endpoint: server.URL + mockContext}
	state := ih.Status(mockDeviceId)
	for k, v := range stateVerification {
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
