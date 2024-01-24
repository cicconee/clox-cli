package cmd

import (
	"fmt"
	"strings"

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
		fmt.Println("Configuring Clox CLI")
		token := APIToken()
		password := Passowrd()
		fmt.Println("API Token:", token)
		fmt.Println("Password:", password)
		fmt.Println("Initialized the Clox CLI")
	},
}

// APIToken will prompt the user to enter an API token. If an empty value is entered, it will
// loop until user enters a value. Once a valid API token is entered, it will return it.
func APIToken() string {
	var token string

	for {
		InString("API Token", &token)
		token = strings.TrimSpace(token)
		if token != "" {
			break
		}

		fmt.Println("Token cannot be empty")
	}

	return token
}

// Password will prompt the user to enter and confirm a password. If passwords do not match,
// it will loop until user confirms a valid password. Once a password is confirmed, it will
// be returned.
func Passowrd() string {
	var pass string
	var confirmPass string

	for {
		InString("Password", &pass)
		InString("Confirm Password", &confirmPass)

		if pass == confirmPass {
			break
		}

		fmt.Println("Passwords do not match")
		pass = ""
		confirmPass = ""
	}

	return pass
}

// InString prints msg and takes a string input from the user. The input value will be stored
// in dst. The prompt is formatted as "msg: ".
func InString(msg string, dst *string) {
	fmt.Printf("%s: ", msg)
	fmt.Scanln(dst)
}
