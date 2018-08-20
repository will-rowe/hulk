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
	Long:  `Distance will compare two sketches and a distance metric (braycurtis/canberra/euclidean/jaccard).`,
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
	metric = distanceCmd.Flags().StringP("metric", "m", "jaccard", "the distance metric to use (braycurtis/canberra/euclidean/jaccard)")
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
