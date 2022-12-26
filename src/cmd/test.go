/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>
*/
package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var testCmd = &cobra.Command{
	Use:   "test <device>",
	Short: "testing function for async socket handling",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "" /* no prefix */, log.Ldate|log.Ltime|log.Lshortfile)
		logger.Print("not implemented")

		// ih := intesishome.New(flagUsername, flagPassword)
		// device := toInt64(args[0])
		// token, err := ih.Token()
		// if err != nil {
		// 	logger.Fatalf("problem fetching token. cause: %v", err)
		// }
		// logger.Printf("received auth token: %d", token)
		// a := command.New(device, token, command.WithLogger(logger), command.WithKeepalive(10*time.Second))
		// err = a.Connect(ih.Controller())
		// if err != nil {
		// 	logger.Fatalf("problem connecting to: %s cause: %v", ih.Controller(), err)
		// }
		// logger.Printf("controller is at: %s", ih.Controller())
		// // err = a.Auth(token)
		// // if err != nil {
		// // 	logger.Fatalf("problem authenticating with controller. cause: %v", err)
		// // }
		// // logger.Print("authenticated")
		// go a.Listen()
		// time.Sleep(20 * time.Second)
		// if err := a.Set(1, 1); err != nil {
		// 	logger.Printf("error conducting set. cause: %v", err)
		// }
		// time.Sleep(20 * time.Second)
		// if err := a.Set(1, 0); err != nil {
		// 	logger.Printf("error conducting set. cause: %v", err)
		// }
		// time.Sleep(20 * time.Second)
		// a.Close()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
