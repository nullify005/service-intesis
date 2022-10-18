package watcher

// TODO: split the package so that the web API is in another file

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
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

// holds the internal state
var (
	watcher state
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

// internal state
type state struct {
	ih        *intesishome.IntesisHome
	devices   []intesishome.Device
	status    map[string]interface{}
	statusRaw map[string]interface{}
	metrics   metrics.Metrics
	mu        sync.Mutex
}

// HVAC POST request
type HVACRequest struct {
	Device int64       `json:"device"`
	Param  string      `json:"param" binding:"required"`
	Value  interface{} `json:"value" binding:"required"`
}

// HVAC GET response
type HVACResponse struct {
	Device intesishome.Device     `json:"device"`
	Status map[string]interface{} `json:"status"`
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
	watcher.ih = intesishome.New(
		w.username, w.password,
		intesishome.WithVerbose(w.verbose),
		intesishome.WithHostname(w.hostname),
	)
	watcher.metrics = metrics.New()
	if ok, err := watcher.ih.HasDevice(w.device); !ok {
		p := "device not found"
		if err != nil {
			p = p + "\nerror: " + err.Error()
		}
		panic(p)
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET(w.metricsPath, promHandler())
	router.GET(w.healthPath, healthHandler)
	router.GET("/hvac/:device", hvacReadHandler)
	router.POST("/hvac/:device", hvacWriteHandler)
	router.GET("/shutdown", shutdownHandler)
	log.Fatal(router.Run(w.listen))
}

func watch(w *Watcher) {
	var err error
	// collect the startup info 1st before entering the loop
	// if we can't bootstrap at ths point then we should panic
	watcher.devices, err = watcher.ih.Devices()
	if err != nil {
		panic(err)
	}
	if err = refreshState(w.device); err != nil {
		panic(err)
	}
	go func() {
		for {
			var err error
			time.Sleep(w.interval)
			if err = refreshState(w.device); err != nil {
				log.Printf("error refreshing state: %v", err.Error())
				continue
			}
			log.Printf("(%v) power: %v mode: %v temp: %v setpoint: %v",
				w.device, watcher.status["power"], watcher.status["mode"],
				watcher.status["temperature"], watcher.status["setpoint"],
			)
			watcher.metrics.SetPoint(float64(watcher.statusRaw["setpoint"].(int) / 10))
			watcher.metrics.Temperature(float64(watcher.statusRaw["temperature"].(int) / 10))
			watcher.metrics.Power(float64(watcher.statusRaw["power"].(int)))
			watcher.metrics.Mode(float64(watcher.statusRaw["mode"].(int)))
		}
	}()
}

func refreshState(device int64) (err error) {
	watcher.mu.Lock()
	defer watcher.mu.Unlock()
	state, err := watcher.ih.Status(device)
	if err != nil {
		return
	}
	mapped := make(map[string]interface{})
	for k, v := range state {
		mV := intesishome.DecodeState(k, v.(int))
		mapped[k] = mV
	}
	watcher.statusRaw = state
	watcher.status = mapped
	return
}

func healthHandler(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func promHandler() gin.HandlerFunc {
	p := promhttp.Handler()
	return func(c *gin.Context) {
		p.ServeHTTP(c.Writer, c.Request)
	}
}

// TODO: add in the device status in here too
// should we actually append this to the Device struct?
func hvacReadHandler(c *gin.Context) {
	resp := HVACResponse{}
	for _, d := range watcher.devices {
		if c.Param("device") == d.ID {
			resp.Device = d
			resp.Status = watcher.status
			c.JSON(http.StatusOK, resp)
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "no such device"})
}

// handles param set requests
// try to conduct the set via the underlying API & then do a state refresh immediately after
func hvacWriteHandler(c *gin.Context) {
	var (
		uid   int
		value int
		err   error
	)
	request := HVACRequest{}
	if err = c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.Device, err = strconv.ParseInt(c.Param("device"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, value, err = intesishome.MapCommand(request.Param, request.Value)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = watcher.ih.Set(request.Device, uid, value); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, request)
	_ = refreshState(request.Device)
}

// signals the watcher that is should shutdown the observation loop & quit
func shutdownHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// func toInt64(s string) (int64, error) {
// 	i, err := strconv.ParseInt(s, 10, 64)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return i, nil
// }
