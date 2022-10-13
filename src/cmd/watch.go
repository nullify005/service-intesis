/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"github.com/nullify005/service-intesis/pkg/watcher"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var (
	_interval *time.Duration
	_listen   *string
	watchCmd  = &cobra.Command{
		Use:   "watch [-i time.Duration] [-l host:port]",
		Short: "watch an AC Units state and expose it to prometheus scraping",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			device := toInt64(args[0])
			w := watcher.New(
				username, password, int64(device),
				watcher.WithDuration(*_interval),
				watcher.WithListen(*_listen),
			)
			w.Watch()
		},
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)
	_listen = watchCmd.Flags().StringP("listen", "l", "127.0.0.1:2112", "hostname:port to listen on to expose metrics")
	_interval = watchCmd.Flags().DurationP("interval", "i", 30*time.Second, "time.Duration polling interval")
}
