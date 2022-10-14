package intesishome

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDeviceId                    string = "127934703953"
	testNilResponsePayload          string = "./assets/tests/nilResponse.json"
	testValidControlResponsePayload string = "./assets/tests/validControlResponse.json"
	testErrorControlResponsePayload string = "./assets/tests/noMethodResponse.json"
	testMalformedResponsePayload    string = "./assets/tests/malformedControlResponse.txt"
)

var (
	_testStateVerifyMapped = map[string]interface{}{
		"alarm_status": 0,
		"mode":         "heat",
		"power":        "on",
		"rssi":         200,
		"setpoint":     200,
		"setpoint_max": 300,
		"setpoint_min": 160,
		"temperature":  210,
	}
	_testStateVerifyUnmapped = map[string]interface{}{
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
			testValidControlResponsePayload,
			func(t *testing.T, d []Device, e error) {
				assert.NoError(t, e)
				assert.Equal(t, len(d), 1)
				assert.Equal(t, d[0].ID, testDeviceId)
			},
		},
		{
			"invalid response",
			http.StatusBadGateway,
			testNilResponsePayload,
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
			testValidControlResponsePayload,
			func(t *testing.T, s map[string]interface{}, e error) {
				assert.NoError(t, e)
				for k, v := range _testStateVerifyUnmapped {
					if _, ok := s[k]; !ok {
						t.Errorf("%s was missing from state map", k)
						continue
					}
					assert.Equal(t, s[k], v)
					mVal := DecodeState(k, v.(int))
					assert.Equal(t, _testStateVerifyMapped[k], mVal)
				}
			},
		},
		{
			"invalid response",
			http.StatusOK,
			testErrorControlResponsePayload,
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
			d, _ := strconv.ParseInt(testDeviceId, 10, 64)
			state, err := ih.Status(int64(d))
			tt.want(t, state, err)
		})
	}
}

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
