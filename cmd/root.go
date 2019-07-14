// Copyright © 2019 Will Rowe
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// the global command line arguments
var (
	kmerSize       *uint                                                     // minimizer k-mer length
	outFile        *string                                                   // basename for the outfile(s)
	defaultOutFile = "./hulk-" + string(time.Now().Format("20060102150405")) // a default output file basename
	logFile        *string                                                   // name to use for log file
	proc           *int                                                      // number of processors to use
	profiling      *bool                                                     // create profile for go pprof
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hulk",
	Short: "Histosketching Using Little Kmers",
	Long: `
	HULK is a tool that creates small, fixed-size sketches from streaming microbiome sequencing data,
	enabling rapid metagenomic dissimilarity analysis. HULK generates an approximate k-mer spectrum from
	a FASTQ data stream, incrementally sketches it and makes similarity search queries against other microbiome sketches.
`,
}

// Execute is a function to launch the requested subcommand
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init is a function to initialise the default command line arguments
func init() {
	kmerSize = RootCmd.PersistentFlags().UintP("kmerSize", "k", 21, "minimizer k-mer length")
	outFile = RootCmd.PersistentFlags().StringP("outFile", "o", defaultOutFile, "directory and basename for saving the outfile(s)")
	logFile = RootCmd.PersistentFlags().String("log", "", "filename for log file, if omitted then STDOUT used by default")
	proc = RootCmd.PersistentFlags().IntP("processors", "p", 1, "number of processors to use")
	profiling = RootCmd.PersistentFlags().Bool("profiling", false, "create the files needed to profile HULK using the go tool pprof")
}
