package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, recorder.Code, http.StatusOK)
	body := recorder.Body.String()
	for _, s := range requiredMetrics {
		assert.Contains(t, body, s)
	}
}
