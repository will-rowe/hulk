// Package helpers contains some helper functions which the HULK CL program needs
package helpers

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Pow is a function to raise a to the power of b
func Pow(a, b uint) uint {
	p := uint(1)
	for b > 0 {
		if b&1 != 0 {
			p *= a
		}
		b >>= 1
		a *= a
	}
	return p
}

// a function to throw error to the log and exit the program
func ErrorCheck(msg error) {
	if msg != nil {
		log.Fatalf("ERROR---> %v\n", msg)
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
		return fmt.Errorf("Required flag `" + flagName + "` has not been set")
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

// CheckSTDIN is a function to check that STDIN can be read
func CheckSTDIN() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("error with STDIN")
	}
	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		return fmt.Errorf("no STDIN found")
	}
	return nil
}

// CheckDir is a function to check that a directory exists
func CheckDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("no directory specified")
	}
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %v", dir)
		}
		return fmt.Errorf("can't access adirectory (check permissions): %v", dir)
	}
	return nil
}

// CheckFile is a function to check that a file can be read
func CheckFile(file string) error {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %v", file)
		}
		return fmt.Errorf("can't access file (check permissions): %v", file)
	}
	return nil
}

// CheckExt is a function to check the extensions of a file
func CheckExt(file string, exts []string) error {
	splitFilename := strings.Split(file, ".")
	finalIdx := len(splitFilename) - 1
	if splitFilename[finalIdx] == "gz" {
		finalIdx--
	}
	err := fmt.Errorf("file does not have recognised extension: %v", file)
	for _, ext := range exts {
		if splitFilename[finalIdx] == ext {
			err = nil
			break
		}
	}
	return err
}

// FanOutUnbuffered is a helper function to set up a fanout concurrency pattern
func FanOutUnbuffered(inputChannel chan uint64, numOutputChannels int) []chan uint64 {
	cs := make([]chan uint64, numOutputChannels)
	for i := 0; i < numOutputChannels; i++ {
		cs[i] = make(chan uint64)
	}
	go func() {
		for val := range inputChannel {
			for _, c := range cs {
				c <- val
			}
		}
		for _, c := range cs {
			// close all our fanOut channels when the input channel is exhausted.
			close(c)
		}
	}()
	return cs
}

// MD5sum will return the MD5sum of a []uint64
func MD5sum(data []uint64) [16]byte {
	sketchBytes := make([]byte, 8*len(data))
	for i := 0; i < len(data); i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, data[i])
		for j, val := range b {
			sketchBytes[(i*8)+j] = val
		}
	}
	return md5.Sum(sketchBytes)
}

// CollectJSONs is a function to find all JSON files in a directory and return a list of file names
func CollectJSONs(inputDir string, recursive bool) ([]string, error) {
	filePaths := []string{}

	// if recursive, find all JSON files in the input directory and its subdirectories
	if recursive == true {
		recursiveSketchGrabber := func(fp string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			matched, err := filepath.Match("*.json", fi.Name())
			if err != nil {
				return err
			}
			// if a json is found, add the file path the list
			if matched {
				filePaths = append(filePaths, fp)
			}
			return nil
		}
		filepath.Walk(inputDir, recursiveSketchGrabber)
	} else {

		// otherwise, just find all JSON files in the supplied directory
		searchTerm := inputDir + "*.json"
		var err error
		filePaths, err = filepath.Glob(searchTerm)
		if err != nil {
			return nil, err
		}
	}

	// check we got some files
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no JSON files found in supplied directory: %v\n", inputDir)
	}
	return filePaths, nil
}

// FloatSlice2string converts []float64 to a string, with a custom delimiter to separate slice values
func FloatSlice2string(floatSlice []float64, sep string) string {
	if len(floatSlice) == 0 {
		return ""
	}
	holder := make([]string, len(floatSlice))
	for i, v := range floatSlice {
		holder[i] = fmt.Sprintf("%.0f", v)
	}
	return strings.Join(holder, sep)
}
