// Copyright © 2018 Will Rowe
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
	"github.com/spf13/cobra"
	"os"
	"time"
)

// the command line arguments
var (
	proc           *int                                                      // number of processors to use
	outFile        *string                                                   // basename for the outfile(s)
	defaultOutFile = "./hulk-" + string(time.Now().Format("20060102150405")) // a default output filename
	profiling      *bool                                                     // create profile for go pprof
	defaultLogFile = "./hulk.log"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hulk",
	Short: "Histosketching Using Little Kmers",
	Long: `
HULK is a tool that creates small, fixed-size sketches from streaming microbiome sequencing data,
enabling rapid metagenomic dissimilarity analysis. HULK generates a k-mer spectrum from a FASTQ data stream,
incrementally sketches it and makes similarity search queries against other microbiome sketches.

It works by using count-min sketching to create a k-mer spectrum from a data stream.
After some reads have been added to a k-mer spectrum, HULK begins to process the counter frequencies and
populates a histosketch. Similarly to MinHash sketches, histosketches can be used to estimate similarity
between microbiome samples.`,
}

/*
  A function to add all child commands to the root command and sets flags appropriately
*/
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// launch subcommand
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/*
  A function to initialise the command line arguments
*/
func init() {
	proc = RootCmd.PersistentFlags().IntP("processors", "p", 1, "number of processors to use")
	outFile = RootCmd.PersistentFlags().StringP("outFile", "o", defaultOutFile, "directory and basename for saving the outfile(s)")
	profiling = RootCmd.PersistentFlags().Bool("profiling", false, "create the files needed to profile HULK using the go tool pprof")
}
