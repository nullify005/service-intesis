/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/spf13/cobra"
)

// devicesCmd represents the devices command
var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "list all of the confgured devices",
	Run: func(cmd *cobra.Command, args []string) {
		ih := intesishome.New(username, password)
		devices, err := ih.Devices()
		if err != nil {
			fmt.Printf("error getting devices: %v\n", err.Error())
			os.Exit(1)
		}
		for _, device := range devices {
			fmt.Println(device.String())
		}
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
}
