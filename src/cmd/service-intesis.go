package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/nullify005/service-intesis/pkg/metrics"
	"github.com/nullify005/service-intesis/pkg/secrets"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	flagMock            = flag.Bool("mock", false, "pass in a test file for mock responses")
	flagMonitor         = flag.Bool("monitor", false, "continuously monitor the state of device")
	flagInterval        = flag.Int("interval", 120, "the interval between state collections")
	flagDevice          = flag.Int64("device", 0, "get status from the device")
	flagUsername        = flag.String("username", "", "intesis cloud username")
	flagPassword        = flag.String("password", "", "intesis cloud password")
	flagListen          = flag.String("listen", "127.0.0.1:2112", "the addr:port to listen on for HTTP requests")
	flagSecrets         = flag.String("secrets", "/.secrets/creds.yaml", "path to the credentials yaml")
	username     string = ""
	password     string = ""
)

const (
	metricsPath string = "/metrics"
	healthPath  string = "/health"
)

func connection() (i *intesishome.Connection) {
	i = &intesishome.Connection{Username: username, Password: password, Mock: *flagMock}
	return
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func monitor(device int64) {
	if device == 0 {
		fmt.Println("no device specified, you must specify a device to monitor")
		os.Exit(1)
	}
	m := metrics.New()
	conn := connection()
	go func() {
		for {
			log.Printf("starting state collection loop for device: %v at interval: %vs", *flagDevice, *flagInterval)
			state := conn.Status(device)
			m.SetPoint(float64(state["setpoint"].(int) / 10))
			m.Temperature(float64(state["temperature"].(int) / 10))
			m.Power(float64(state["power"].(int)))
			m.Mode(float64(state["mode"].(int)))
			time.Sleep(time.Duration(*flagInterval) * time.Second)
		}
	}()
	http.HandleFunc(healthPath, healthHandler)
	http.Handle(metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(*flagListen), nil))
}

func status(device int64) {
	conn := connection()
	state := conn.Status(device)
	keys := make([]string, 0)
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		mappedV := conn.MapValue(k, state[k].(int))
		fmt.Printf("%v: %v\n", k, mappedV)
	}
	os.Exit(0)
}

func devices() {
	conn := connection()
	devices := conn.Devices()
	for _, device := range devices {
		fmt.Println(device.String())
	}
	os.Exit(0)
}

func creds() {
	if *flagUsername != "" {
		username = *flagUsername
		password = *flagPassword
		return
	}
	s, err := secrets.Read(*flagSecrets)
	if err != nil {
		fmt.Printf("unable to load secrets: %v\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	username = s.Username
	password = s.Password
}

func main() {
	flag.Parse()
	creds()
	if *flagMonitor {
		monitor(*flagDevice)
	}
	if *flagDevice != 0 {
		status(*flagDevice)
	}
	// by default we list the devices
	devices()
}
