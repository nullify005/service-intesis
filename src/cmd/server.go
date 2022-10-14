/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>
*/
package cmd

import (
	"time"

	"github.com/nullify005/service-intesis/pkg/mock"
	"github.com/spf13/cobra"
)

// devicesCmd represents the devices command
var (
	_serverTimeout *time.Duration
	serverCmd      = &cobra.Command{
		Use:   "server",
		Short: "runs a test tcp server",
		Run: func(cmd *cobra.Command, args []string) {
			t := mock.NewTCPServer(
				mock.WithTCPListen(flagTCPServer),
				mock.WithTCPReadTimeout(*_serverTimeout),
			)
			go t.Run()
			h := mock.NewHTTPServer(mock.WithHTTPListen(flagHTTPServer))
			h.Run()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			flags := cmd.InheritedFlags()
			// disable parent required flags
			flags.SetAnnotation("username", cobra.BashCompOneRequiredFlag, []string{"false"})
			flags.SetAnnotation("password", cobra.BashCompOneRequiredFlag, []string{"false"})
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	_serverTimeout = serverCmd.Flags().Duration("timeout", mock.DefaultReadTimeout, "read timeout duration")
}
