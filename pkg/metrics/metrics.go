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
)

// NOTE: the http handler still needs to be registered elsewhere
func New() *metrics {
	lock.Lock()
	defer lock.Unlock()
	if m == nil {
		m = &metrics{}
		prometheus.MustRegister(mTemp)
		prometheus.MustRegister(mSetPoint)
		prometheus.MustRegister(mPower)
		mTemp.Set(20)
		mSetPoint.Set(20)
		mPower.Set(0)
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
