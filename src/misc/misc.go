// contains some misc helper functions
package misc

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// a function to throw error to the log and exit the program
func ErrorCheck(msg error) {
	if msg != nil {
		log.Fatal("encountered error: ", msg)
	}
}

// a function to check for required flags
func CheckRequiredFlags(flags *pflag.FlagSet) error {
	requiredError := false
	flagName := ""

	flags.VisitAll(func(flag *pflag.Flag) {
		requiredAnnotation := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if len(requiredAnnotation) == 0 {
			return
		}

		flagRequired := requiredAnnotation[0] == "true"

		if flagRequired && !flag.Changed {
			requiredError = true
			flagName = flag.Name
		}
	})

	if requiredError {
		return errors.New("Required flag `" + flagName + "` has not been set")
	}

	return nil
}

// StartLogging is a function to start the log...
func StartLogging(logFile string) *os.File {
	logPath := strings.Split(logFile, "/")
	joinedLogPath := strings.Join(logPath[:len(logPath)-1], "/")
	if len(logPath) > 1 {
		if _, err := os.Stat(joinedLogPath); os.IsNotExist(err) {
			if err := os.MkdirAll(joinedLogPath, 0700); err != nil {
				log.Fatal("can't create specified directory for log")
			}
		}
	}
	logFH, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return logFH
}
