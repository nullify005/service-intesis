/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status device",
	Short: "fetch the current status of an AC Unit",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ih := intesishome.New(_username, _password)
		device := toInt64(args[0])
		state, err := ih.Status(int64(device))
		if err != nil {
			fmt.Printf("encountered error fetching status: %v\n", err.Error())
			os.Exit(1)
		}
		keys := make([]string, 0)
		for k := range state {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			mappedV := intesishome.DecodeState(k, state[k].(int))
			fmt.Printf("%v: %v\n", k, mappedV)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
