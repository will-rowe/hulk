// Copyright © 2018 Science and Technology Facilities Council (UK) <will.rowe@stfc.ac.uk>

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/histosketch"
	"github.com/will-rowe/hulk/src/misc"
)

// the command line arguments
var (
	sketch1 *string // the first sketch to compare
	sketch2 *string // the second sketch to compare
	metric  *string // the distance metric to use
)

// distanceCmd represents the distance command
var distanceCmd = &cobra.Command{
	Use:   "distance",
	Short: "Distance will compare two sketches and a distance metric",
	Long:  `Distance will compare two sketches and a distance metric (braycurtis/canberra/euclidean/jaccard/weightedjaccard).`,
	Run: func(cmd *cobra.Command, args []string) {
		runDistance()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return misc.CheckRequiredFlags(cmd.Flags())
	},
}

// a function to initialise the command line arguments
func init() {
	sketch1 = distanceCmd.Flags().StringP("sketch1", "1", "", "the first sketch to compare")
	sketch2 = distanceCmd.Flags().StringP("sketch2", "2", "", "the second sketch to compare")
	metric = distanceCmd.Flags().StringP("metric", "m", "jaccard", "the distance metric to use (braycurtis/canberra/euclidean/jaccard/weightedjaccard)")
	distanceCmd.MarkFlagRequired("sketch1")
	distanceCmd.MarkFlagRequired("sketch2")
	RootCmd.AddCommand(distanceCmd)
}

/*
  The main function for the distance subcommand
*/
func runDistance() {
	// load sketches
	s1, err := histosketch.LoadSketch(*sketch1)
	misc.ErrorCheck(err)
	s2, err := histosketch.LoadSketch(*sketch2)
	misc.ErrorCheck(err)
	// check them, calculate the specified distance and print the value
	dist, err := s1.GetDistance(s2, *metric)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(dist)
}
