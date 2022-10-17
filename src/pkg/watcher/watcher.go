package watcher

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/nullify005/service-intesis/pkg/metrics"
	"github.com/nullify005/service-intesis/pkg/secrets"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultListen      string        = "127.0.0.1:2112"
	DefaultInterval    time.Duration = 30 * time.Second
	DefaultSecretsPath string        = "/.secrets/creds.yaml"
	DefaultHealthPath  string        = "/health"
	DefaultMetricsPath string        = "/metrics"
)

// the watcher polls the Intesis Home cloud API for changes in state
// and exposes for Prometheus scraping
type Watcher struct {
	interval    time.Duration
	listen      string
	hostname    string
	username    string
	password    string
	device      int64
	healthPath  string
	metricsPath string
	verbose     bool
	secrets     string
}

type Option func(w *Watcher)

// sets an alternate hostname for the Intesis Home API (testing / debugging)
func WithHostname(host string) Option {
	return func(w *Watcher) {
		w.hostname = host
	}
}

// which host:port to listen on for metrics
func WithListen(listen string) Option {
	return func(w *Watcher) {
		w.listen = listen
	}
}

// the polling interval
func WithDuration(interval time.Duration) Option {
	return func(w *Watcher) {
		w.interval = interval
	}
}

// the context path which the metrics will be exposed on
func WithMetricsPath(path string) Option {
	return func(w *Watcher) {
		w.metricsPath = path
	}
}

// the context path the health endpoint will be on for k8s liveliness checks
func WithHealthPath(path string) Option {
	return func(w *Watcher) {
		w.healthPath = path
	}
}

// whether debug logging should be enabled
func WithVerbose(v bool) Option {
	return func(w *Watcher) {
		w.verbose = v
	}
}

// path to the Intesis Cloud credentials
func WithSecrets(s string) Option {
	return func(w *Watcher) {
		w.secrets = s
	}
}

func New(user, pass string, device int64, opts ...Option) Watcher {
	w := Watcher{
		interval:    DefaultInterval,
		listen:      DefaultListen,
		username:    user,
		password:    pass,
		device:      device,
		healthPath:  DefaultHealthPath,
		metricsPath: DefaultMetricsPath,
		verbose:     false,
		secrets:     DefaultSecretsPath,
		hostname:    intesishome.DefaultHostname,
	}
	for _, opt := range opts {
		opt(&w)
	}
	if w.username == "" || w.password == "" {
		s, err := secrets.Read(w.secrets)
		if err != nil {
			p := fmt.Sprintf("no credentials specified: %s", err.Error())
			panic(p)
		}
		w.username = s.Username
		w.password = s.Password
	}
	return w
}

func (w *Watcher) Watch() {
	log.SetPrefix("service-intesis: ")
	log.SetFlags(log.LstdFlags)
	log.Printf("starting watcher")
	log.Printf("device: %v", w.device)
	log.Printf("interval: %v", w.interval)
	log.Printf("listen: %s", w.listen)
	watch(w)
	router := gin.Default()
	router.GET(w.metricsPath, promHandler())
	router.GET(w.healthPath, healthHandler)
	router.GET("/hvac/:device", hvacHandlerRead)
	log.Fatal(router.Run(w.listen))
}

func watch(w *Watcher) {
	m := metrics.New()
	ih := intesishome.New(
		w.username, w.password,
		intesishome.WithVerbose(w.verbose),
		intesishome.WithHostname(w.hostname),
	)
	if ok, err := ih.HasDevice(w.device); !ok {
		if err != nil {
			fmt.Println(err.Error())
		}
		p := fmt.Sprintf("%v was not a valid device", w.device)
		panic(p)
	}
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

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func promHandler() gin.HandlerFunc {
	p := promhttp.Handler()
	return func(c *gin.Context) {
		p.ServeHTTP(c.Writer, c.Request)
	}
}

func hvacHandlerRead(c *gin.Context) {
	device, err := toInt64(c.Param("device"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid device id")
		return
	}
	c.String(http.StatusOK, "%v", device)
}

func toInt64(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
