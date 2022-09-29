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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		flagMonitor  = flag.Bool("monitor", false, "continuously monitor the state of device")
		flagInterval = flag.Int("interval", 120, "the interval between state collections")
		flagDevice   = flag.Int64("device", 0, "get status from the device")
		flagMock     = flag.String("mock", "", "pass in a test file for mock responses")
		flagUsername = flag.String("username", "", "intesis cloud username")
		flagPassword = flag.String("password", "", "intesis cloud password")
		flagListen   = flag.String("listen", "127.0.0.1:2112", "the addr:port to listen on for HTTP requests")
	)
	const (
		metricsPath string = "/metrics"
	)

	flag.Parse()
	if *flagUsername == "" || *flagPassword == "" {
		fmt.Println("intesis home cloud controller.\nusage:")
		flag.PrintDefaults()
		return
	}
	// by default we list the devices
	ih := intesishome.Connection{Username: *flagUsername, Password: *flagPassword}
	if *flagMock != "" {
		ih.Mock = *flagMock
	}
	if *flagMonitor {
		if *flagDevice == 0 {
			fmt.Println("no device specified, you must specify a device to monitor")
			os.Exit(1)
		}
		m := metrics.New()
		go func() {
			for {
				log.Printf("starting state collection loop at interval: %v", *flagInterval)
				// TODO: pull the metrics & update the instance
				m.SetPoint(20)
				m.Temperature(20)
				time.Sleep(time.Duration(*flagInterval) * time.Second)
			}
		}()
		http.Handle(metricsPath, promhttp.Handler())
		log.Fatal(http.ListenAndServe(fmt.Sprintf(*flagListen), nil))
		os.Exit(0)
	}
	if *flagDevice != 0 {
		status := ih.Status(*flagDevice)
		keys := make([]string, 0)
		for k := range status {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%v: %v\n", k, status[k])
		}
		os.Exit(0)
	}
	devices := ih.Devices()
	for _, device := range devices {
		fmt.Println(device.String())
	}
}
