package cmd

import (
	"fmt"
	"os"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/prompt"
	"github.com/spf13/cobra"
)

// The 'init' command.
//
// initCmd will initialize and set up the Clox CLI configuration.
//
// TODO:
//   - Encrypt the api token with password.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up the Clox CLI",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		exists, err := store.DirExists()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if !exists {
			err := store.WriteDir()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		fileExists, err := store.FileExists()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if fileExists {
			fmt.Println("Clox CLI already initialized")
			os.Exit(0)
		}

		user, err = config.NewUser(keys, prompt.Passowrd(), prompt.APIToken())
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if err := store.WriteConfigFile(user); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Success")
		os.Exit(0)
	},
}
