// Copyright © 2018 Will Rowe <will.rowe@stfc.ac.uk>
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
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/hulk/src/stream"
	"github.com/will-rowe/hulk/src/version"
)

// the command line arguments
var (
	fastq      *[]string // list of FASTQ files to sketch
	epsilon    *float64  // relative accuracy for countmin sketching
	delta      *float64  // relative probability for countmin sketching
	kSize      *int      // size of k-mer
	interval   *int      // size of read sampling interval (0 == no interval)
	sketchSize *uint     // size of sketch
	decayRatio *float64  // the decay ratio used for concept drift (1.00 = concept drift disabled)
	streaming  *bool // writes the sketches to STDOUT (as well as to disk)
)

// the sketchCmd
var sketchCmd = &cobra.Command{
	Use:   "sketch",
	Short: "Create a histosketch from a set of reads",
	Long:  `Create a histosketch from a set of reads.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSketch()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	fastq = sketchCmd.Flags().StringSliceP("fastq", "f", []string{}, "FASTQ file(s) to sketch (can also pipe in STDIN)")
	epsilon = sketchCmd.Flags().Float64P("epsilon", "e", 0.0001, "relative accuracy factor for countmin sketching")
	delta = sketchCmd.Flags().Float64P("delta", "d", 0.99, "relative accuracy probability for countmin sketching")
	kSize = sketchCmd.Flags().IntP("kmerSize", "k", 11, "size of k-mer")
	interval = sketchCmd.Flags().IntP("interval", "i", 0, "size of read sampling interval (default 0 (= no interval))")
	sketchSize = sketchCmd.Flags().UintP("sketchSize", "s", 256, "size of sketch")
	decayRatio = sketchCmd.Flags().Float64P("decayRatio", "x", 1.0, "decay ratio used for concept drift (1.0 = concept drift disabled)")
	streaming = sketchCmd.Flags().Bool("stream", false, "prints the sketches to STDOUT after every interval is reached (sketches also written to disk)")
	RootCmd.AddCommand(sketchCmd)
}

//  a function to check user supplied parameters
func sketchParamCheck() error {
	// check the supplied FASTQ file(s)
	if len(*fastq) == 0 {
		stat, err := os.Stdin.Stat()
		if err != nil {
			fmt.Println("error with STDIN")
			return fmt.Errorf("error with STDIN")
		}
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			fmt.Println("no STDIN found")
			return fmt.Errorf("no STDIN found")
		}
		log.Printf("\tinput file: using STDIN")
	} else {
		for _, fastqFile := range *fastq {
			if _, err := os.Stat(fastqFile); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("FASTQ file does not exist: %v", fastqFile)
				} else {
					return fmt.Errorf("can't access FASTQ file (check permissions): %v", fastqFile)
				}
			}
			splitFilename := strings.Split(fastqFile, ".")
			if splitFilename[len(splitFilename)-1] == "gz" {
				if splitFilename[len(splitFilename)-2] == "fastq" || splitFilename[len(splitFilename)-2] == "fq" {
					continue
				}
			} else {
				if splitFilename[len(splitFilename)-1] == "fastq" || splitFilename[len(splitFilename)-1] == "fq" {
					continue
				}
			}
			return fmt.Errorf("does not look like a FASTQ file: %v", fastqFile)
		}
	}
	// setup the outFile
	filePath := filepath.Dir(*outFile)
	if filePath != "." {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				return fmt.Errorf("can't create specified output directory: ", err)
			}
		}
	}
	// set number of processors to use
	if *proc <= 0 || *proc > runtime.NumCPU() {
		*proc = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*proc)
	return nil
}

/*
  The main function for the sketch command
*/
func runSketch() {
	// set up profiling
	if *profiling == true {
		//defer profile.Start(profile.MemProfile, profile.ProfilePath("./")).Stop()
		defer profile.Start(profile.ProfilePath("./")).Stop()
	}
	// start logging
	logFH := misc.StartLogging((*outFile + ".log"))
	defer logFH.Close()
	log.SetOutput(logFH)
	log.Printf("hulk (version %s)", version.VERSION)
	log.Printf("starting the sketch subcommand")
	// check the supplied files and then log some stuff
	log.Printf("checking parameters...")
	misc.ErrorCheck(sketchParamCheck())
	log.Printf("\tno. processors: %d", *proc)
	log.Printf("\tk-mer size: %d", *kSize)
	log.Printf("\tepsilon value: %.4f", *epsilon)
	log.Printf("\tdelta value: %.2f", *delta)
	log.Printf("\tsketch size: %d", *sketchSize)
	if *decayRatio == 1 {
		log.Printf("\tconcept drift: disabled")
	} else {
		log.Printf("\tconcept drift: enabled")
		log.Printf("\tdecay ratio: %.2f", *decayRatio)
	}
	if *streaming {
		log.Printf("\tstreaming: enabled")
	} else {
		log.Printf("\tstreaming: disabled")
	}
	// create the base countmin sketch for recording the k-mer spectrum
	log.Printf("creating the base countmin sketch for kmer counting...")
	// TODO: epsilon and delta values need some checking
	spectrum := histosketch.NewCountMinSketch(*epsilon, *delta, 1.0)
	log.Printf("\tnumber of tables: %d", spectrum.Tables())
	log.Printf("\tnumber of counters per table: %d", spectrum.Counters())
	// create the pipeline
	pipeline := stream.NewPipeline()
	// initialise processes
	log.Printf("initialising the data streams...")
	dataStream := stream.NewDataStreamer()
	fastqHandler := stream.NewFastqHandler()
	fastqChecker := stream.NewFastqChecker()
	counter := stream.NewCounter()
	sketcher := stream.NewSketcher()
	// add in the process parameters TODO: consolidate and remove some of these
	dataStream.InputFile = *fastq
	fastqChecker.Ksize, counter.Ksize = *kSize, *kSize
	counter.Interval = *interval / *proc
	counter.Spectrum, sketcher.Spectrum = spectrum.Copy(), spectrum.Copy()
	counter.NumCPU, sketcher.NumCPU = *proc, *proc
	counter.SketchSize, sketcher.SketchSize = *sketchSize, *sketchSize
	sketcher.DecayRatio = *decayRatio
	sketcher.OutFile = *outFile
	sketcher.Stream = *streaming

	// arrange pipeline processes
	fastqHandler.Input = dataStream.Output
	fastqChecker.Input = fastqHandler.Output
	counter.Input = fastqChecker.Output
	sketcher.Input = counter.TheCollector

	// submit each process to the pipeline to be run
	pipeline.AddProcesses(dataStream, fastqHandler, fastqChecker, counter, sketcher)
	log.Printf("\tnumber of stream processes: %d\n", len(pipeline.Processes))
	pipeline.Run()
	log.Printf("finished")
}
