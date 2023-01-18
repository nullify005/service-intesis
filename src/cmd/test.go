/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>
*/
package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var testCmd = &cobra.Command{
	Use:   "test <device>",
	Short: "testing function for async socket handling",
	// Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		flags := cmd.InheritedFlags()
		// disable parent required flags since we're hoping to use a secrets file
		flags.SetAnnotation("username", cobra.BashCompOneRequiredFlag, []string{"false"})
		flags.SetAnnotation("password", cobra.BashCompOneRequiredFlag, []string{"false"})
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "" /* no prefix */, log.Ldate|log.Ltime|log.Lshortfile)
		logger.Print("not implemented")
		// jsonEvt := make(chan []byte)
		// go func() {
		// 	logger.Print("now reading from stdin")
		// 	var buf []byte
		// 	for {
		// 		os.Stdin.Read()
		// 		read, err := io.ReadAll(os.Stdin)
		// 		buf = append(buf, read...)
		// 		if err != nil {
		// 			logger.Printf("read error. cause: %v", err)
		// 			return
		// 		}
		// 		logger.Printf("the buf is: %s", buf)
		// 		start := -1
		// 		opens := 0
		// 		closes := 0
		// 		for i := 0; i < len(buf); i++ {
		// 			logger.Printf("looking at: %c", buf[i])
		// 			if buf[i] == '{' {
		// 				logger.Print("found an open")
		// 				if start == -1 {
		// 					start = i
		// 					logger.Printf("start is at: %d", start)
		// 				}
		// 				opens++
		// 				logger.Printf("opens: %d", opens)
		// 			}
		// 			if buf[i] == '}' {
		// 				logger.Print("found a close")
		// 				closes++
		// 				logger.Printf("closes: %d", closes)
		// 				if opens == closes {
		// 					logger.Printf("end is at: %d", i)
		// 					logger.Printf("json payload is: %s", buf[start:i+1])
		// 					jsonEvt <- buf[start : i+1]
		// 					buf = slices.Delete(buf, start, i+1)
		// 				}
		// 			}
		// 		}
		// 	}
		// }()
		// for {
		// 	json := <-jsonEvt
		// 	logger.Printf("received json payload: %s", json)
		// }
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
