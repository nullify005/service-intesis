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

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get key",
	Short: "get the value of parameter key from the AC Unit",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ih := intesishome.New(_username, _password)
		device := toInt64(args[0])
		state, err := ih.Status(int64(device))
		if err != nil {
			fmt.Printf("error getting status: %v", err.Error())
			os.Exit(1)
		}
		if _, ok := state[args[1]]; !ok {
			fmt.Printf("unable to locate status: %s", args[1])
			os.Exit(1)
		}
		mappedV := intesishome.DecodeState(args[1], state[args[1]].(int))
		fmt.Printf("(%s) %s: %v\n", args[0], args[1], mappedV)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
