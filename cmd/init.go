package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/spf13/cobra"
)

// The 'init' command.
//
// initCmd will initialize and set up the Clox CLI configuration.
//
// TODO:
//   - Create private and public key.
//   - Hash the password
//   - Encrypt the api token, private key, and public key with password
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up the Clox CLI",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		store, err := config.NewStore()
		if err != nil {
			fmt.Printf("Error: Failed initializing the configuration: %v\n", err)
			os.Exit(1)
		}

		exists, err := store.DirExists()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if exists {
			fmt.Println("Clox CLI already initialized")
			os.Exit(0)
		}

		err = store.Write(config.WriteFileParams{
			Password: Passowrd(),
			APIToken: APIToken(),
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Success")
		os.Exit(0)
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
