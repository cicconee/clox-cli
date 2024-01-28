package cmd

import (
	"fmt"
	"os"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/security"
	"github.com/spf13/cobra"
)

// The storage for persisting the Clox CLI configuration. It is initialized before executing
// any commands.
var store *config.Store

// The user of Clox CLI. It is initialized in the rootCmd when running commands that require
// user authentication.
var user *config.User

// Handles the AES encryption logic.
var aes = &crypto.AES{}

// The key manager for the users public-private key pairs.
var keys = &security.Keys{AES: aes}

// The root command of Clox CLI.
var rootCmd = &cobra.Command{
	Use:           "clox",
	Short:         "The official client of the Clox API",
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// TODO: Check if cmd is not 'init' and that configuration is setup.
		if cmd.Name() != "init" { // && validConfiguration
			// TODO: If not init command and configuration is set up, set the configuration
			// values from the config files that are needed for all commands. Values may
			// include password hash, private key, public key, api token.
			//
			// TODO: Data will be encrypted with the users password, so prompt the user for
			// their password. Check if its valid by comparing against password hash.
			//
			// TODO: With password decrypt the api token and private key and set the values
			// to variables.
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mkdirCmd)
	mkdirCmd.Flags().StringVarP(&mkdirPath, "path", "p", "", "Path to create directory in")
}

func Execute() {
	s, err := config.NewStore()
	if err != nil {
		fmt.Printf("Error: Failed initializing the configuration: %v\n", err)
		os.Exit(1)
	}

	store = s

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("\n[ERROR] %v\n", err)
	}
}
