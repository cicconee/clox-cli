package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// The root command of Clox CLI.
var rootCmd = &cobra.Command{
	Use:   "clox",
	Short: "The official client of the Clox API",
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

// The 'init' command.
//
// initCmd will initialize and set up the Clox CLI configuration.
//
// TODO:
//   - Create hidden .clox directory in users home directory.
//   - Create ./clox/config.json file that will store the password hash, encrypted api token,
//     encrypted private key, and the public key.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up the Clox CLI",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var token string
		var pass string
		var confirmPass string

		fmt.Println("Configuring Clox CLI")
		fmt.Print("API Token: ")
		fmt.Scanln(&token)
		for {
			fmt.Print("Password: ")
			fmt.Scanln(&pass)
			fmt.Print("Confirm Password: ")
			fmt.Scanln(&confirmPass)

			if pass == confirmPass {
				fmt.Printf("Pass: %q\n", pass)
				fmt.Printf("Conf: %q\n", confirmPass)
				break
			}

			fmt.Println("Passwords do not match")
			pass = ""
			confirmPass = ""
		}

		fmt.Println("API Token:", token)
		fmt.Println("Initialized the Clox CLI")
	},
}

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

var mkdirPath string

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mkdirCmd)
	mkdirCmd.Flags().StringVarP(&mkdirPath, "path", "p", "", "Path to create directory in")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("\n[ERROR] %v\n", err)
	}
}

func main() {
	Execute()
}
