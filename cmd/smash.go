// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
)

// the command line arguments
var (
	sketchDir    *string // the directory containing the sketches
	recursive    *bool   // recursively search the supplied directory
	jsMatrix     *bool   // create a pairwise Jaccard Similarity matrix
	wjsMatrix    *bool   // create a pairwise weighted Jaccard Similarity matrix
	bannerMatrix *bool   // create a matrix to train banner on
	label        *string // used in the bannerMatrix - assigns all sketches to a single label
)

// the sketches
var hSketches map[string]*histosketch.SketchStore

// smashCmd represents the smash command
var smashCmd = &cobra.Command{
	Use:   "smash",
	Short: "Smash a bunch of sketches and return a similarity matrix",
	Long: `
The smash subcommand takes 2 or more hulk sketches and
smashes them together...

You can use smash to:
* perform pairwise comparisons between each sketch, storing these in
a matrix so that you can plot them. You can then use the matrix to
make nice plots in R and see how similar your samples are.

* create a sketch matrix to use as input for Banner, which will train an ML
classifier.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSmash()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	sketchDir = smashCmd.Flags().StringP("sketchDir", "d", "./", "the directory containing the sketches to smash (compare)...")
	recursive = smashCmd.Flags().Bool("recursive", false, "recursively search the supplied sketch directory (-d)")
	jsMatrix = smashCmd.Flags().Bool("jsMatrix", false, "create a pairwise Jaccard Similarity matrix")
	wjsMatrix = smashCmd.Flags().Bool("wjsMatrix", false, "create a pairwise weighted Jaccard Similarity matrix")
	bannerMatrix = smashCmd.Flags().Bool("bannerMatrix", false, "create a matrix to train banner on")
	label = smashCmd.Flags().StringP("label", "l", "0", "assign a class to all the sketches (for bannerMatrix)")
	RootCmd.AddCommand(smashCmd)
}

// makeJSMatrix perform pairwise Jaccard SImilarity estimates, populates a matrix and writes to csv
func makeJSMatrix() error {
	// create the jaccard similarity matrix csv outfile
	jsmFile, err := os.Create((*outFile + ".js-matrix.csv"))
	defer jsmFile.Close()
	if err != nil {
		return err
	}
	jsmWriter := csv.NewWriter(jsmFile)
	defer jsmWriter.Flush()
	// create an ordering
	ordering := make([]string, len(hSketches))
	sort.Strings(ordering)
	count := 0
	for id := range hSketches {
		ordering[count] = id
		count++
	}
	// write the header
	if jsmWriter.Write(ordering) != nil {
		return err
	}
	// hulk smash
	for _, id := range ordering {
		jsVals := make([]string, len(ordering))
		for i, id2 := range ordering {
			// the GetDistance method will call the sketch check, which will make sure the sketches are compatible (in terms of length etc)
			jd, err := hSketches[id].GetDistance(hSketches[id2], "jaccard")
			misc.ErrorCheck(err)
			// convert js to the Jaccard Similarity, then to string so it can be written with the csv library
			js := 100 - (jd * 100)
			jsVals[i] = strconv.FormatFloat(js, 'f', 2, 64)
		}
		if jsmWriter.Write(jsVals) != nil {
			return err
		}
	}
	return nil
}

// makeWJSMatrix perform pairwise weighted Jaccard SImilarity estimates, populates a matrix and writes to csv
func makeWJSMatrix() error {
	// create the jaccard similarity matrix csv outfile
	jsmFile, err := os.Create((*outFile + ".wjs-matrix.csv"))
	defer jsmFile.Close()
	if err != nil {
		return err
	}
	jsmWriter := csv.NewWriter(jsmFile)
	defer jsmWriter.Flush()
	// create an ordering
	ordering := make([]string, len(hSketches))
	count := 0
	for id := range hSketches {
		ordering[count] = id
		count++
	}
	// write the header
	if jsmWriter.Write(ordering) != nil {
		return err
	}
	// hulk smash
	for _, id := range ordering {
		jsVals := make([]string, len(ordering))
		for i, id2 := range ordering {
			// the GetDistance method will call the sketch check, which will make sure the sketches are compatible (in terms of length etc)
			jd, err := hSketches[id].GetDistance(hSketches[id2], "weightedjaccard")
			misc.ErrorCheck(err)
			// convert js to the Jaccard Similarity, then to string so it can be written with the csv library
			js := 100 - (jd * 100)
			jsVals[i] = strconv.FormatFloat(js, 'f', 2, 64)
		}
		if jsmWriter.Write(jsVals) != nil {
			return err
		}
	}
	return nil
}

// makeBannerMatrix checks sketches, creates a matrix for Banner, assigns a class and writes to csv
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
	for _, sketch := range hSketches {
		printString := make([]string, sketch.Length)
		for i, element := range sketch.Sketch {
			printString[i] = fmt.Sprintf("%d", element)
		}
		// append the label to the line and then write it
		printString = append(printString, *label)
		if bannerWriter.Write(printString) != nil {
			return err
		}
	}
	return nil
}

/*
  The main function for the smash subcommand
*/
func runSmash() {
	// exit if no smash option requested
	if *jsMatrix == false && *bannerMatrix == false && *wjsMatrix == false {
		fmt.Println("please rerun with a smash option (--jsMatrix, --wjsMatrix, --bannerMatrix)")
		os.Exit(1)
	}
	// create the sketch pile
	var err error
	hSketches, _, err = histosketch.CreateSketchCollection(*sketchDir, *recursive)
	misc.ErrorCheck(err)
	// check we have at least 2 sketches
	if len(hSketches) < 2 {
		fmt.Println("need at least 2 sketches for hulk smash!")
		os.Exit(1)
	}
	// run the requested smash
	if *jsMatrix {
		misc.ErrorCheck(makeJSMatrix())
	}
	if *wjsMatrix {
		misc.ErrorCheck(makeWJSMatrix())
	}
	if *bannerMatrix {
		misc.ErrorCheck(makeBannerMatrix())
	}
}
