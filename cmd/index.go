// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/lshForest"
	"github.com/will-rowe/hulk/src/misc"
	"github.com/will-rowe/hulk/src/version"
)

// the command line arguments
var (
	indexName      *string  // the name of the index that is to be created/augmented/searched
	indexFunc      *string  // the indexing function to run
	indexSketchDir *string  // the directory containing the sketches to create an index from/add to an index/search against and index
	indexRecursive *bool    // recursively search the supplied directory
	jsThresh       *float64 // the jaccard similarity theshold for lsh indexing
)

// the available indexing functions
var indexFuncs = []string{"create", "add", "search"}

// the sketches
var hulkSketches map[string]*histosketch.SketchStore
var sketchLength int

// indexCmd represents the distance command
var indexCmd = &cobra.Command{
	Use:   "index -r <create | add | search>",
	Short: "Index will create, add to, or search an LSH Forest index of HULK sketches",
	Long:  `Index will create, add to, or search an LSH Forest index of HULK sketches.`,
	Run: func(cmd *cobra.Command, args []string) {
		runIndex()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	indexFunc = indexCmd.Flags().StringP("run", "r", "create", "the indexing function to run (create/add/search)")
	indexName = indexCmd.Flags().StringP("indexName", "n", "", "the name of the index that is to be created/augmented/searched")
	indexSketchDir = indexCmd.Flags().StringP("sketchDir", "d", "./", "the directory containing the sketches to use")
	indexRecursive = indexCmd.Flags().Bool("recursive", false, "recursively search the supplied sketch directory (-d)")
	jsThresh = indexCmd.Flags().Float64P("jsThresh", "j", 0.97, "the jaccard similarity threshold for the LSH Forest index")
	indexCmd.MarkFlagRequired("indexName")
	RootCmd.AddCommand(indexCmd)
}

//  a function to check user supplied parameters
func indexParamCheck() error {
	var err error
	// check the index function
	funCheck := false
	for _, fun := range indexFuncs {
		if fun == *indexFunc {
			funCheck = true
		}
	}
	if funCheck == false {
		return fmt.Errorf("unrecognised index function: %v\n", *indexFunc)
	}
	// create the sketch pile
	hulkSketches, sketchLength, err = histosketch.CreateSketchCollection(*indexSketchDir, *indexRecursive)
	if err != nil {
		return err
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
  The main function for the index subcommand
*/
func runIndex() {
	// start logging
	logFH := misc.StartLogging((*outFile + ".log"))
	defer logFH.Close()
	log.SetOutput(logFH)
	log.Printf("hulk (version %s)", version.VERSION)
	log.Printf("starting the index subcommand")
	// check the supplied files and then log some stuff
	log.Printf("checking parameters...")
	misc.ErrorCheck(indexParamCheck())
	log.Printf("\tno. sketches: %d", len(hulkSketches))
	log.Printf("\tsketch length: %d", sketchLength)
	log.Printf("\tno. processors: %d", *proc)

	// run the specified index function
	log.Printf("\tindexing function: %v", *indexFunc)
	switch *indexFunc {
	case "create":
		log.Printf("creating index...")
		log.Printf("\tjaccard similarity: %.2f", *jsThresh)
		// create the forest
		hulkForest := lshForest.NewLSHforest(sketchLength, *jsThresh)
		numHF, numB := hulkForest.Settings()
		log.Printf("\tno. hashes per bucket: %d", numHF)
		log.Printf("\tno. buckets: %d", numB)
		// add entries
		counter := 0
		for filename, sketchStore := range hulkSketches {
			hulkForest.Add(filename, sketchStore.Sketch)
			counter++
		}
		log.Printf("added %d sketches to the index", counter)
		// save the index to disk
		misc.ErrorCheck(hulkForest.Dump(*indexName))
		log.Printf("written to disk (%v)", *indexName)
	case "add":
		// load the index, check it and index it // TODO: need to add a check to make sure the length and jsThresh are the same as the original parameters
		log.Printf("loading index...")
		hulkForest := lshForest.NewLSHforest(sketchLength, *jsThresh)
		misc.ErrorCheck(hulkForest.Load(*indexName))
		numHF, numB := hulkForest.Settings()
		log.Printf("\tno. hashes per bucket: %d", numHF)
		log.Printf("\tno. buckets: %d", numB)
		// now add to the index
		log.Printf("adding to index...")
		counter := 0
		for filename, sketchStore := range hulkSketches {
			hulkForest.Add(filename, sketchStore.Sketch)
			counter++
		}
		log.Printf("\t no. sketches added: %d", counter)
		// rewrite it to disk
		misc.ErrorCheck(hulkForest.Dump(*indexName))
		log.Printf("written to disk (%v)", *indexName)
	case "search":
		log.Printf("loading index...")
		hulkForest := lshForest.NewLSHforest(sketchLength, *jsThresh)
		misc.ErrorCheck(hulkForest.Load(*indexName))
		hulkForest.Index()
		numHF, numB := hulkForest.Settings()
		log.Printf("\tno. hashes per bucket: %d", numHF)
		log.Printf("\tno. buckets: %d", numB)
		// now search the index
		log.Printf("searching index...")
		counter := 0
		for filename, sketchStore := range hulkSketches {
			counter++
			results := hulkForest.Query(sketchStore.Sketch)
			fmt.Printf("query:\t%v\n", filename)
			for i, hit := range results {
				fmt.Printf("hit %d:\t%v\n", i, hit)
			}
			fmt.Println("----")
		}
		log.Printf("\t no. sketches queried: %d", counter)
	default:
		log.Fatal("unsupported indexing function")
	}

}
