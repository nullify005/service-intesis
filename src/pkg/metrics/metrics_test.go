package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var requiredMetrics = []string{
	"hvac_mode_state 2",
	"hvac_power_state 1",
	"hvac_setpoint_celcius 10",
	"hvac_temperature_celcius 10",
}

const metricsPath string = "/metrics"

func TestMetrics(t *testing.T) {
	m := New()
	m.SetPoint(10)
	m.Temperature(10)
	m.Mode(2)
	m.Power(1)
	request := httptest.NewRequest(http.MethodGet, metricsPath, nil)
	recorder := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("received non 200 response code: %v", recorder.Code)
	}
	body := recorder.Body.String()
	for _, s := range requiredMetrics {
		if !strings.Contains(body, s) {
			t.Errorf("metrics response didn't contain: %s", s)
		}
	}
}
