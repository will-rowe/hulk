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
