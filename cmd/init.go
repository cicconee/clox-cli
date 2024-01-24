package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
