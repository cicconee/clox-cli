package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mkdirPath string

// The 'mkdir' command.
//
// mkdirCmd will create a new directory on the Clox server. The path flag is optional.
// If not provided directory will default to the users root.
//
// TODO:
//   - Call Clox API endpoint to create directory.
var mkdirCmd = &cobra.Command{
	Use:   "mkdir name",
	Short: "Create a new directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Name: %s\n", args[0])
		fmt.Printf("Path: %s\n", mkdirPath)
	},
}
