// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
)

// the command line arguments
var (
	sketchFile *string // the first sketch to compare
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print will print a hulk sketch to screen in csv format",
	Long:  `Print will print a hulk sketch to screen in csv format.`,
	Run: func(cmd *cobra.Command, args []string) {
		runPrint()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	sketchFile = printCmd.Flags().StringP("sketchFile", "f", "", "the sketch to print")
	printCmd.MarkFlagRequired("sketchFile")
	RootCmd.AddCommand(printCmd)
}

func runPrint() {
	// load sketch
	s1, err := histosketch.LoadSketch(*sketchFile)
	misc.ErrorCheck(err)
	// print the sketch
	//fmt.Printf("sketch file:\n%v\nsketch minimums:\n", s1.File)
	for _, val := range s1.Sketch {
		fmt.Printf("%d,", val)
	}
	fmt.Printf("%v\n", *sketchFile)
}
