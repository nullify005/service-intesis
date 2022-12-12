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
	_flagInterval *time.Duration
	_flagListen   *string
	_flagSecrets  *string
	watchCmd      = &cobra.Command{
		Use:   "watch [-i time.Duration] [-l host:port] device",
		Short: "watch an AC Units state and expose it to prometheus scraping",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			flags := cmd.InheritedFlags()
			// disable parent required flags since we're hoping to use a secrets file
			flags.SetAnnotation("username", cobra.BashCompOneRequiredFlag, []string{"false"})
			flags.SetAnnotation("password", cobra.BashCompOneRequiredFlag, []string{"false"})
		},
		Run: func(cmd *cobra.Command, args []string) {
			device := toInt64(args[0])
			w := watcher.New(
				flagUsername, flagPassword,
				int64(device),
				watcher.WithSecrets(*_flagSecrets),
				watcher.WithDuration(*_flagInterval),
				watcher.WithListen(*_flagListen),
				watcher.WithVerbose(flagVerbose),
				watcher.WithHostname(flagHTTPServer),
			)
			w.Watch()
		},
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)
	_flagListen = watchCmd.Flags().StringP("listen", "l", watcher.DefaultListen, "hostname:port to listen on to expose metrics")
	_flagInterval = watchCmd.Flags().DurationP("interval", "i", watcher.DefaultInterval, "time.Duration polling interval")
	_flagSecrets = watchCmd.Flags().StringP("secrets", "s", watcher.DefaultSecretsPath, "the location of the Intesis Cloud credentials")
}
