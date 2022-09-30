package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	m     *metrics
	lock  = &sync.Mutex{}
	mTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hvac_temperature_celcius",
		Help: "HVAC observed temperature in celcius",
	})
	mSetPoint = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hvac_setpoint_celcius",
		Help: "HVAC desired temperature in celcius",
	})
	mPower = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hvac_power_state",
		Help: "HVAC power state 0: off 1: on",
	})
	mMode = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hvac_mode_state",
		Help: "HVAC mode state 0: auto 1: heat 2: dry 3: fan 4: cool",
	})
)

// NOTE: the http handler still needs to be registered elsewhere
// TODO: write some tests for this
func New() *metrics {
	lock.Lock()
	defer lock.Unlock()
	if m == nil {
		m = &metrics{}
		prometheus.MustRegister(mTemp)
		prometheus.MustRegister(mSetPoint)
		prometheus.MustRegister(mPower)
		prometheus.MustRegister(mMode)
		mTemp.Set(20)
		mSetPoint.Set(20)
		mPower.Set(0)
		mMode.Set(0)
	}
	return m
}

func (m *metrics) Temperature(t float64) {
	mTemp.Set(t)
}

func (m *metrics) SetPoint(t float64) {
	mSetPoint.Set(t)
}

func (m *metrics) Power(t float64) {
	mPower.Set(t)
}

func (m *metrics) Mode(t float64) {
	mMode.Set(t)
}
