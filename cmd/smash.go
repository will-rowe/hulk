package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/helpers"
	"github.com/will-rowe/hulk/src/sketchio"
	"github.com/will-rowe/hulk/src/version"
)

// the command line arguments
var (
	sketchDir    *string // the directory containing the sketches
	recursive    *bool   // recursively search the supplied directory
	algo         *string // which sketching algorithm to use (histosketch, KMV, khf)
	metric       *string // the distance metric to use
	bannerMatrix *bool   // also write a bannerMatrix
)

// the available distance metrics
var availMetrics = []string{"jaccard", "weightedjaccard"}

// the sketches
var hSketches map[string]*sketchio.HULKdata

// smashCmd is used by cobra
var smashCmd = &cobra.Command{
	Use:   "smash",
	Short: "Smash a bunch of sketches and return a distance matrix",
	Long: `
		Smash a bunch of sketches and return a distance matrix.

		This subcommand performs pairwise comparisons of sketches and then writes a distance matrix.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSmash()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return helpers.CheckRequiredFlags(cmd.Flags())
	},
}

// init the command line arguments
func init() {
	sketchDir = smashCmd.Flags().StringP("sketchDir", "d", "./", "the directory containing the sketches to smash (compare)...")
	recursive = smashCmd.Flags().Bool("recursive", false, "recursively search the supplied sketch directory (-d)")
	algo = smashCmd.PersistentFlags().StringP("algorithm", "a", "histosketch", fmt.Sprintf("tells HULK which sketching algorithm to use %v", sketchio.AvailAlgorithms))
	metric = smashCmd.Flags().StringP("metric", "m", "jaccard", fmt.Sprintf("tells HULK which distance metric to use %v", availMetrics))
	bannerMatrix = smashCmd.Flags().Bool("bannerMatrix", false, "write a matrix file for banner")
	RootCmd.AddCommand(smashCmd)
}

// runSmash is the main function for this subcommand
func runSmash() {

	// set up cpu profiling
	if *profiling == true {
		// you can swap to memory profiling by uncommenting the line below and then commenting out the line after
		//defer profile.Start(profile.MemProfile, profile.ProfilePath("./")).Stop()
		defer profile.Start(profile.ProfilePath("./")).Stop()
	}

	// set up the log
	if *logFile != "" {
		logFH := helpers.StartLogging(*logFile)
		defer logFH.Close()
		log.SetOutput(logFH)
	} else {
		// normal behaviour is to print the log to STDOUT
		log.SetOutput(os.Stdout)
	}

	// start the sketch subcommand
	log.Printf("this is hulk (version %s)\n", version.VERSION)
	log.Printf("starting the smash subcommand\n")

	// create the empty sketch pile
	hSketches = make(map[string]*sketchio.HULKdata)

	// check the parameters and load the sketches
	helpers.ErrorCheck(smashParamCheck())
	log.Printf("checking parameters and collecting sketches...\n")
	log.Printf("\talgorithm: %v\n", *algo)
	log.Printf("\tk-mer size: %d\n", *kmerSize)
	log.Printf("\tcreate matrix for banner: %v\n", *bannerMatrix)
	log.Printf("\tnumber of sketch objects: %d\n", len(hSketches))
	log.Print("HULK SMASH!\n")

	// hulk smash
	helpers.ErrorCheck(makeMatrix())
	log.Printf("\twritten similarity matrix to disk: %v\n", *outFile+".hulk-matrix.csv")

	// create banner matrix if requested
	if *bannerMatrix {
		helpers.ErrorCheck(makeBannerMatrix())
		log.Printf("\twritten banner matrix to disk: %v\n", *outFile+".banner-matrix.csv")
	}

	log.Printf("finished")
}

// smashParamCheck is a function to check user supplied parameters
func smashParamCheck() error {
	// check the metric choice
	ok := false
	for _, availMetric := range availMetrics {
		if *metric == availMetric {
			ok = true
		}
	}
	if ok == false {
		return fmt.Errorf("supplied distance metric is not available: %v\nplease select one of the following: %v", *metric, availMetrics)
	}

	// check the algorithm choice
	ok = false
	for _, availAlgo := range sketchio.AvailAlgorithms {
		if *algo == availAlgo {
			ok = true
		}
	}
	if ok == false {
		return fmt.Errorf("supplied algorithm not available: %v\nplease select one of the following: %v", *algo, sketchio.AvailAlgorithms)
	}

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

	// check the supplied sketch directory
	helpers.ErrorCheck(helpers.CheckDir(*sketchDir))

	// add a slash if not already present in dir param
	sDir := []byte(*sketchDir)
	if sDir[len(sDir)-1] != 47 {
		sDir = append(sDir, 47)
	}

	// find all the json files under the supplied dir
	jsonFiles, err := helpers.CollectJSONs(string(sDir), *recursive)
	helpers.ErrorCheck(err)

	// load the json files, check they are hulk sketches and get an array of sketch objects
	for _, jsonFile := range jsonFiles {

		// load the json
		loadedSketch, err := sketchio.LoadHULKdata(jsonFile)
		helpers.ErrorCheck(err)

		// add the sketch to the pile
		hSketches[jsonFile] = loadedSketch
	}

	// make sure there are at least 2 sketches to smash
	if len(hSketches) < 2 {
		return fmt.Errorf("%d sketches found in the supplied directory, HULK needs at least 2 to smash!\n", len(hSketches))
	}

	return nil
}

// makeMatrix perform pairwise comparisons of sketches, populates a distance matrix and then writes to csv
func makeMatrix() error {

	// create the matrix csv outfile
	matrixFile, err := os.Create((*outFile + ".hulk-matrix.csv"))
	defer matrixFile.Close()
	if err != nil {
		return err
	}
	matrixWriter := csv.NewWriter(matrixFile)
	defer matrixWriter.Flush()

	// sort the sketches
	ordering := make([]string, len(hSketches))
	count := 0
	for fileName := range hSketches {
		ordering[count] = fileName
		count++
	}
	sort.Strings(ordering)

	// write the header
	if matrixWriter.Write(ordering) != nil {
		return err
	}

	// hulk smash
	for _, fileName := range ordering {
		distances := make([]string, len(ordering))
		for i, fileName2 := range ordering {

			// the GetDistance method will call the sketch check, which will make sure the sketches are compatible (in terms of length etc)
			distanceVal, err := hSketches[fileName].GetDistance(hSketches[fileName2], *metric, *kmerSize, *algo)
			helpers.ErrorCheck(err)

			// convert the distance to a similarity, then to string so it can be written with the csv library
			similarityVal := 100 - (distanceVal * 100)
			distances[i] = strconv.FormatFloat(similarityVal, 'f', 2, 64)
		}
		if matrixWriter.Write(distances) != nil {
			return err
		}
	}
	return nil
}

// makeBannerMatrix checks sketches, creates a matrix for Banner, assigns a banner label and writes to csv
func makeBannerMatrix() error {
	// create the Banner matrix csv outfile
	bannerFile, err := os.Create((*outFile + ".banner-matrix.csv"))
	defer bannerFile.Close()
	if err != nil {
		return err
	}
	bannerWriter := csv.NewWriter(bannerFile)
	defer bannerWriter.Flush()
	// range over each sketch and create the line for the csv writer
	for _, HULKdata := range hSketches {

		// get the correct k size
		sketchObj, err := HULKdata.FindSketch(*kmerSize, *algo)
		if err != nil {
			return err
		}
		sketch := sketchObj.GetSketch()

		// convert sketch to string
		printString := make([]string, len(sketch))
		for i, element := range sketch {
			printString[i] = fmt.Sprintf("%d", element)
		}

		// append the banner label to the line and then write it
		printString = append(printString, HULKdata.Banner)
		if bannerWriter.Write(printString) != nil {
			return err
		}
	}
	return nil
}
