/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/nullify005/service-intesis/pkg/intesishome"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var (
	setCmd = &cobra.Command{
		Use:   "set key value",
		Short: "set parameter key to value on an AC Unit",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			device := toInt64(args[0])
			key := args[1]
			val := args[2]
			ih := intesishome.New(flagUsername, flagPassword)
			// what happens when the uid isn't within the map?
			uid, value, err := intesishome.MapCommand(key, val)
			if err != nil {
				fmt.Printf("encoutered error during mapping: %s\n", err.Error())
				os.Exit(1)
			}
			if err = ih.Set(int64(device), uid, value); err != nil {
				fmt.Printf("encountered error during set: %s\n", err.Error())
				os.Exit(1)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(setCmd)
}
