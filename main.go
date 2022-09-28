package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/nullify005/service-intesis/intesishome"
)

// var flagList = flag.Bool("list", false, "list the devices")
var flagDevice = flag.Int64("device", 0, "get status from the device")
var flagMock = flag.String("mock", "", "pass in a test file for mock responses")
var flagUsername = flag.String("username", "", "intesis cloud username")
var flagPassword = flag.String("password", "", "intesis cloud password")

func main() {
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
