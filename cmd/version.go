package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/will-rowe/hulk/src/version"
)

// hulkVersion is the command used to print the software version and then exit
var hulkVersion = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version and exits",
	Long:  `Prints the current version and exits`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.VERSION)
	},
}

func init() {
	RootCmd.AddCommand(hulkVersion)
}
