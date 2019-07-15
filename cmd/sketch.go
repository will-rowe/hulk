package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/helpers"
	"github.com/will-rowe/hulk/src/pipeline"
	"github.com/will-rowe/hulk/src/version"
)

// the command line arguments
var (
	fastq       *[]string // list of FASTQ files to sketch
	fasta       *bool     // tells HULK that the input file is actually in FASTA format
	windowSize  *uint     // minimizer window size [2/3 of k-mer length]. A minimizer is the smallest k-mer in a window of w consecutive k-mers.
	interval    *uint     // size of k-mer sampling interval (0 == no interval)
	sketchSize  *uint     // size of sketch
	decayRatio  *float64  // the decay ratio used for concept drift (1.00 = concept drift disabled)
	streaming   *bool     // writes the sketches to STDOUT (as well as to disk)
	bannerLabel *string   // adds a label to the saved sketch, for use with banner
	addKHF      *bool     // HULK will also produce a MinHash KHF sketch
	addKMV      *bool     // HULK will also produce a MinHash KMV sketch
)

// sketchCmd is used by cobra
var sketchCmd = &cobra.Command{
	Use:   "sketch",
	Short: "Create a sketch from a set of reads",
	Long: `
		Create a sketch from a set of reads.
		
		The sketch subcommand can be used to create a histosketch, minhash or count min sketch.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSketch()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return helpers.CheckRequiredFlags(cmd.Flags())
	},
}

// init the command line arguments
func init() {
	fastq = sketchCmd.Flags().StringSliceP("fastq", "f", []string{}, "FASTQ file(s) to sketch (can also pipe in STDIN)")
	fasta = sketchCmd.Flags().Bool("fasta", false, "tells HULK that the input file is actually FASTA format (.fna/.fasta/.fa), not FASTQ (experimental feature)")
	windowSize = sketchCmd.Flags().UintP("windowSize", "w", 9, "minimizer window size")
	interval = sketchCmd.Flags().UintP("interval", "i", 0, "size of k-mer sampling interval (default 0 (= no interval))")
	sketchSize = sketchCmd.Flags().UintP("sketchSize", "s", 50, "size of sketch")
	decayRatio = sketchCmd.Flags().Float64P("decayRatio", "x", 1.0, "decay ratio used for concept drift (1.0 = concept drift disabled)")
	streaming = sketchCmd.Flags().Bool("stream", false, "prints the sketches to STDOUT after every interval is reached, whilst still writting them to disk (log file is redirected to disk))")
	bannerLabel = sketchCmd.Flags().StringP("bannerLabel", "b", "blank", "adds a label to the sketch object, for use with BANNER")
	addKHF = sketchCmd.Flags().Bool("khf", false, "also generate a MinHash K-Hash Functions sketch")
	addKMV = sketchCmd.Flags().Bool("kmv", false, "also generate a MinHash K-Minimum Values (bottom-k) sketch")
	sketchCmd.Flags().SortFlags = false
	RootCmd.AddCommand(sketchCmd)
}

// runSketch is the main function for this subcommand
func runSketch() {

	// set up cpu profiling
	if *profiling == true {
		// you can swap to memory profiling by uncommenting the line below and then commenting out the line after
		//defer profile.Start(profile.MemProfile, profile.ProfilePath("./")).Stop()
		defer profile.Start(profile.ProfilePath("./")).Stop()
	}

	// if streaming sketches out, write the log file to disk
	if *streaming == true {
		// make sure a filename for the log exists
		if *logFile == "" {
			*logFile = *outFile + ".log"
		}
		logFH := helpers.StartLogging(*logFile)
		defer logFH.Close()
		log.SetOutput(logFH)
	} else {
		// normal behaviour is to print the log to STDOUT
		log.SetOutput(os.Stdout)
	}

	// start the sketch subcommand
	start := time.Now()
	log.Printf("this is hulk (version %s)\n", version.VERSION)
	log.Printf("please cite Rowe et al. 2019, doi: https://doi.org/10.1186/s40168-019-0653-2")
	log.Printf("starting the sketch subcommand\n")

	// check the supplied files and then log some stuff
	log.Printf("checking parameters...\n")
	helpers.ErrorCheck(sketchParamCheck())
	if *fasta {
		log.Printf("\tmode: FASTA\n")
	} else {
		log.Printf("\tmode: FASTQ\n")
	}
	log.Printf("\tno. processors: %d\n", *proc)
	log.Printf("\tminimizer k-mer size: %d\n", *kmerSize)
	log.Printf("\tminimizer window size: %d\n", *windowSize)

	log.Printf("\tsketch size: %d\n", *sketchSize)
	if *streaming {
		log.Printf("\tstreaming: enabled\n")
	} else {
		log.Printf("\tstreaming: disabled\n")
	}
	if *decayRatio == 1 {
		log.Printf("\tconcept drift: disabled\n")
	} else {
		log.Printf("\tconcept drift: enabled\n")
		log.Printf("\tdecay ratio: %.2f\n", *decayRatio)
	}
	spectrumSize := helpers.Pow(*kmerSize, 4)
	log.Printf("\tnumber of bins in k-mer spectrum: %d\n", spectrumSize)
	// adding any additional sketches?
	log.Printf("\tadding KHF sketch: %v\n", *addKHF)
	log.Printf("\tadding KMV sketch: %v\n", *addKMV)

	// create the runtime info struct
	hulkInfo := &pipeline.Info{
		Version: version.VERSION,
	}

	// add the sketch command to the hulk runtime info
	hulkInfo.Sketch = &pipeline.SketchCmd{
		Fasta:        *fasta,
		KmerSize:     *kmerSize,
		WindowSize:   *windowSize,
		SpectrumSize: spectrumSize,
		SketchSize:   *sketchSize,
		DecayRatio:   *decayRatio,
		Stream:       *streaming,
		Interval:     *interval,
		OutFile:      *outFile,
		NumMinions:   *proc * 1, // TODO: can increase minions for faster minimizer generation but big bottleneck happens during flushing
		BannerLabel:  *bannerLabel,
		KHF:          *addKHF,
		KMV:          *addKMV,
	}

	// create the pipeline
	log.Printf("initialising sketching pipeline...\n")
	sketchPipeline := pipeline.NewPipeline()

	// initialise processes
	log.Printf("\tinitialising the processes\n")
	dataStream := pipeline.NewDataStreamer(hulkInfo)
	fastqHandler := pipeline.NewFastqHandler(hulkInfo)
	fastqHasher := pipeline.NewSeqMinimizer(hulkInfo)
	sketcher := pipeline.NewSketcher(hulkInfo)

	// connect the pipeline processes
	log.Printf("\tconnecting data streams\n")
	dataStream.Connect(*fastq)
	fastqHandler.Connect(dataStream)
	fastqHasher.Connect(fastqHandler)
	sketcher.Connect(fastqHasher)

	// submit each process to the pipeline and run it
	sketchPipeline.AddProcesses(dataStream, fastqHandler, fastqHasher, sketcher)
	log.Printf("\tnumber of processes added to the sketching pipeline: %d\n", sketchPipeline.GetNumProcesses())
	log.Printf("\tnumber of minions in the sketching pool: %d\n", hulkInfo.Sketch.NumMinions)
	sketchPipeline.Run()

	log.Printf("finished in %s", time.Since(start))
}

// sketchParamCheck is a function to check user supplied parameters
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

	// check the supplied FASTQ file(s)
	if len(*fastq) == 0 {
		helpers.ErrorCheck(helpers.CheckSTDIN())
		log.Printf("\tinput file: using STDIN")
	} else {
		for _, fastqFile := range *fastq {
			helpers.ErrorCheck(helpers.CheckFile(fastqFile))
			helpers.ErrorCheck(helpers.CheckExt(fastqFile, []string{"fastq", "fq", "fasta", "fna", "fa"}))
		}
	}
	return nil
}
