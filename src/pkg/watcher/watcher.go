package watcher

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/nullify005/service-intesis/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Watcher struct {
	interval    time.Duration
	listen      string
	hostname    string
	username    string
	password    string
	device      int64
	healthPath  string
	metricsPath string
}

type Option func(w *Watcher)

func WithHostname(host string) Option {
	return func(w *Watcher) {
		w.hostname = host
	}
}

func WithListen(listen string) Option {
	return func(w *Watcher) {
		w.listen = listen
	}
}

func WithDuration(interval time.Duration) Option {
	return func(w *Watcher) {
		w.interval = interval
	}
}

func WithMetricsPath(path string) Option {
	return func(w *Watcher) {
		w.metricsPath = path
	}
}

func WithHealsthPath(path string) Option {
	return func(w *Watcher) {
		w.healthPath = path
	}
}

func New(user, pass string, device int64, opts ...Option) Watcher {
	w := Watcher{
		interval:    30 * time.Second,
		listen:      "127.0.0.1:2112",
		username:    user,
		password:    pass,
		device:      device,
		healthPath:  "/health",
		metricsPath: "/metrics",
	}
	for _, opt := range opts {
		opt(&w)
	}
	return w
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "ok")
}

func (w *Watcher) Watch() {
	log.Printf("starting watcher")
	log.Printf("device: %v", w.device)
	log.Printf("interval: %v", w.interval)
	log.Printf("listen: %s", w.listen)
	watch(w)
	http.HandleFunc(w.healthPath, healthHandler)
	http.Handle(w.metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(w.listen, nil))
}

func watch(w *Watcher) {
	log.SetPrefix("service-intesis: ")
	log.SetFlags(log.LstdFlags)
	m := metrics.New()
	ih := intesishome.New(w.username, w.password)
	go func() {
		for {
			time.Sleep(w.interval)
			state, err := ih.Status(w.device)
			if err != nil {
				log.Printf("error getting status: %v", err.Error())
				continue
			}
			mapped := make(map[string]interface{})
			for k, v := range state {
				mV := intesishome.DecodeState(k, v.(int))
				mapped[k] = mV
			}
			log.Printf("(%v) power: %v mode: %v temp: %v setpoint: %v",
				w.device, mapped["power"], mapped["mode"],
				mapped["temperature"], mapped["setpoint"],
			)
			m.SetPoint(float64(state["setpoint"].(int) / 10))
			m.Temperature(float64(state["temperature"].(int) / 10))
			m.Power(float64(state["power"].(int)))
			m.Mode(float64(state["mode"].(int)))
		}
	}()
}
