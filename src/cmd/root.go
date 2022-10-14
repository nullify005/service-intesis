/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	flagUsername   string // username for intesis cloud
	flagPassword   string // password for intesis cloud
	flagVerbose    bool   // debug logging
	flagTCPServer  string // debug local emulated TCPServer
	flagHTTPServer string // debug local emulated HTTPServer

	rootCmd = &cobra.Command{
		Use:   "service-intesis",
		Short: "An API integration with the Intesis Cloud + Intesis Home services",
	}
)

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagUsername, "username", "u", "", "Intesis Cloud Username")
	rootCmd.PersistentFlags().StringVarP(&flagPassword, "password", "p", "", "Intesis Cloud Password")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbosity")
	rootCmd.PersistentFlags().StringVarP(&flagTCPServer, "tcpserver", "t", "", "use the following TCPServer host:port for HVAC control commands (DEBUG)")
	rootCmd.PersistentFlags().StringVar(&flagHTTPServer, "httpserver", "", "use the following HTTPServer host:port for HVAC status (DEBUG)")
	rootCmd.MarkPersistentFlagRequired("username")
	rootCmd.MarkPersistentFlagRequired("password")
}

func toInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Printf("unable to coerce %s to int64: %v", s, err.Error())
		os.Exit(1)
	}
	return i
}
