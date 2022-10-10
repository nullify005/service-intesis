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
	_username string
	_password string

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
	rootCmd.PersistentFlags().StringVarP(&_username, "username", "u", "", "Intesis Cloud Username")
	rootCmd.PersistentFlags().StringVarP(&_password, "password", "p", "", "Intesis Cloud Password")
	rootCmd.MarkPersistentFlagRequired("username")
	rootCmd.MarkPersistentFlagRequired("password")
}

func toInt64(s string) int64 {
	r, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("unable to coerce %s to int: %v", s, err.Error())
		os.Exit(1)
	}
	return int64(r)
}
