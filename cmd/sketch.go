// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

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
	epsilon     *float64     // epsilon value for countminsketch generation
	delta     *float64     // delta value for countminsketch generation
	kSize      *int      // size of k-mer
	minCount   *int      // minimum count number for a kmer to be added to the histosketch from this interval
	interval   *int      // size of read sampling interval (0 == no interval)
	sketchSize *uint     // size of sketch
	decayRatio *float64  // the decay ratio used for concept drift (1.00 = concept drift disabled)
	streaming  *bool     // writes the sketches to STDOUT (as well as to disk)
	fasta      *bool     // tells HULK that the input file is in FASTA format
	chunkSize  *int      // splits the FASTA entry to equally sized chunks (if FASTA length not exactly divisible by chunkSize, last chunk will be smaller)
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
	epsilon = sketchCmd.Flags().Float64P("epsilon", "e", 0.0001, "epsilon value for countminsketch generation")
	delta = sketchCmd.Flags().Float64P("delta", "d", 0.90, "delta value for countminsketch generation")
	kSize = sketchCmd.Flags().IntP("kmerSize", "k", 31, "size of k-mer")
	minCount = sketchCmd.Flags().IntP("minCount", "m", 1, "minimum k-mer count for it to be histosketched for a given interval")
	interval = sketchCmd.Flags().IntP("interval", "i", 0, "size of read sampling interval (default 0 (= no interval))")
	sketchSize = sketchCmd.Flags().UintP("sketchSize", "s", 100, "size of sketch")
	decayRatio = sketchCmd.Flags().Float64P("decayRatio", "x", 1.0, "decay ratio used for concept drift (1.0 = concept drift disabled)")
	streaming = sketchCmd.Flags().Bool("stream", false, "prints the sketches to STDOUT after every interval is reached (sketches also written to disk)")
	fasta = sketchCmd.Flags().Bool("fasta", false, "tells HULK that the input file is actually FASTA format (.fna/.fasta/.fa), not FASTQ (experimental feature)")
	chunkSize = sketchCmd.Flags().IntP("chunkSize", "z", -1, "the chunk size for shredding FASTA sequences (requires --fasta) (use -1 to deactivate)")
	sketchCmd.Flags().SortFlags = false
	RootCmd.AddCommand(sketchCmd)
}

//  a function to check user supplied parameters
func sketchParamCheck() error {
	// setup the outFile
	filePath := filepath.Dir(*outFile)
	if filePath != "." {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				return fmt.Errorf("can't create specified output directory: %v", err)
			}
		}
	}
	// set number of processors to use
	if *proc <= 0 || *proc > runtime.NumCPU() {
		*proc = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*proc)
	// check if using STDIN or file(s)
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
		return nil
	}
	// check the supplied file(s)
	return checkInputFiles()
}

// if files are being read, check they exist and are FASTQ/FASTA
func checkInputFiles() error {
	for _, fastqFile := range *fastq {
		if _, err := os.Stat(fastqFile); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %v", fastqFile)
			} else {
				return fmt.Errorf("can't access file (check permissions): %v", fastqFile)
			}
		}
		suffix1, suffix2, suffix3 := "fastq", "fq", "fq"
		if *fasta == true {
			suffix1, suffix2, suffix3 = "fasta", "fna", "fa"
		}
		splitFilename := strings.Split(fastqFile, ".")
		var ext string
		if splitFilename[len(splitFilename)-1] == "gz" {
			ext = splitFilename[len(splitFilename)-2]
		} else {
			ext = splitFilename[len(splitFilename)-1]
		}
		switch ext {
		case suffix1:
			continue
		case suffix2:
			continue
		case suffix3:
			continue
		case "":
			return fmt.Errorf("could not parse filename")
		default:
			return fmt.Errorf("does not look like a %v file: %v", suffix1, fastqFile)
		}
	}
	return nil
}

/*
  The main function for the sketch subcommand
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
	if *fasta {
		log.Printf("\tmode: FASTA\n")
		log.Printf("\tchunk size: %d", *chunkSize)
	} else {
		log.Printf("\tmode: FASTQ\n")
	}
	log.Printf("\tno. processors: %d", *proc)
	log.Printf("\tk-mer size: %d", *kSize)
	log.Printf("\tmin. k-mer count: %d", *minCount)
	log.Printf("\tepsilon: %.2f", *epsilon)
	log.Printf("\tdelta: %.4f", *delta)
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
	fastqHandler.Fasta, counter.Fasta = *fasta, *fasta
	fastqChecker.Ksize, counter.Ksize = *kSize, *kSize
	counter.Interval = *interval / *proc
	counter.Spectrum, sketcher.Spectrum = spectrum.Copy(), spectrum.Copy()
	counter.NumCPU, sketcher.NumCPU = *proc, *proc
	counter.SketchSize, sketcher.SketchSize = *sketchSize, *sketchSize
	counter.ChunkSize = *chunkSize
	sketcher.MinCount = float64(*minCount)
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
