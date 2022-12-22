/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>
*/
package cmd

import (
	"log"
	"os"
	"time"

	"github.com/nullify005/service-intesis/pkg/async"
	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var testCmd = &cobra.Command{
	Use:   "test <device>",
	Short: "testing function for async socket handling",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "" /* no prefix */, log.Ldate|log.Ltime|log.Lshortfile)

		ih := intesishome.New(
			flagUsername, flagPassword,
			intesishome.WithVerbose(flagVerbose), intesishome.WithHostname(flagHTTPServer),
		)
		device := toInt64(args[0])
		a := async.New(async.WithDevice(device), async.WithLogger(logger), async.WithKeepalive(10*time.Second))
		token, err := ih.Token()
		if err != nil {
			logger.Fatalf("problem fetching token. cause: %v", err)
		}
		logger.Printf("received auth token: %d", token)
		err = a.Connect(ih.Controller())
		if err != nil {
			logger.Fatalf("problem connecting to: %s cause: %v", ih.Controller(), err)
		}
		logger.Printf("controller is at: %s", ih.Controller())
		// err = a.Auth(token)
		// if err != nil {
		// 	logger.Fatalf("problem authenticating with controller. cause: %v", err)
		// }
		// logger.Print("authenticated")
		go a.EventListener(12345)
		time.Sleep(20 * time.Second)
		if err := a.Set(1, 1); err != nil {
			logger.Printf("error conducting set. cause: %v", err)
		}
		time.Sleep(20 * time.Second)
		if err := a.Set(1, 0); err != nil {
			logger.Printf("error conducting set. cause: %v", err)
		}
		time.Sleep(20 * time.Second)
		a.Close()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
