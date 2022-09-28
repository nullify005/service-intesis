package intesishome

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const mockContext string = "/api.php/get/control"
const mockDeviceId int64 = 127934703953

func TestDevices(t *testing.T) {
	server := mockServer()
	defer server.Close()

	ih := &Connection{Username: "user", Password: "pass"}
	devices := ih.Devices()
	for _, d := range devices {
		if d.ID != "127934703953" {
			t.Errorf("wanted %s got %s", d.ID, "127934703953")
		}
	}
}

func TestState(t *testing.T) {
	server := mockServer()
	defer server.Close()

	ih := &Connection{Username: "user", Password: "pass"}
	state := ih.Status(mockDeviceId)

	var m map[string]interface{}
	data, _ := json.Marshal(state)
	json.Unmarshal(data, &m)
	s := string("")
	for k, v := range m {
		s += fmt.Sprintf("%s: %s ", k, v)
	}
	// TODO: actually barf if the results aren't what we expect
}

func mockServer() (s *httptest.Server) {
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != mockContext {
			e := fmt.Errorf("expected to request: %s got: %s", mockContext, r.URL.Path)
			error.Error(e)
		}
		w.WriteHeader(http.StatusOK)
		p, err := os.ReadFile(mockPayload)
		if err != nil {
			e := fmt.Errorf("unable to read in payload: %s with: %s", mockPayload, err)
			error.Error(e)
		}
		w.Write(p)
	}))
	return
}
