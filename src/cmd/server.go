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
	_serverTCPListen  *string
	_serverHTTPListen *string
	_serverTimeout    *time.Duration
	serverCmd         = &cobra.Command{
		Use:   "server",
		Short: "runs a test tcp server",
		Run: func(cmd *cobra.Command, args []string) {
			t := mock.NewTCPServer(
				mock.WithTCPListen(*_serverTCPListen),
				mock.WithTCPReadTimeout(*_serverTimeout),
			)
			go t.Run()
			h := mock.NewHTTPServer(mock.WithHTTPListen(*_serverHTTPListen))
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
	_serverTCPListen = serverCmd.Flags().String("tcpserver", mock.DefaultTCPListen, "the TCPServer host:port to listen on")
	_serverHTTPListen = serverCmd.Flags().String("httpserver", mock.DefaultHTTPListen, "the HTTPServer host:port to listen on")
	_serverTimeout = serverCmd.Flags().DurationP("timeout", "t", mock.DefaultReadTimeout, "read timeout duration")
}
