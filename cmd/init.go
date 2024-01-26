package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		config, err := NewConfig()
		if err != nil {
			fmt.Printf("Error: Failed initializing the configuration: %v\n", err)
			os.Exit(1)
		}

		exists, err := config.Exists()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if exists {
			fmt.Println("Clox CLI already initialized")
			os.Exit(0)
		}

		err = config.Write(WriteFileParams{
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

// Config manages the configuration for the Clox CLI app. It ensures all the proper directories
// and files are initialized. User configuration settings can be read with Config.
//
// Config should only be initialized by calling NewConfig.
type Config struct {
	// The path to the .clox directory. Path will always be the path to the users directory
	// with /.clox appended at the end.
	Path string
}

// NewConfig creates a Config and sets to the Path to the users home directory with /.clox
// appended to the end. If it cannot get the users home directory an error is returned.
func NewConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed getting home directory: %w", err)
	}

	return &Config{
		Path: filepath.Join(homeDir, ".clox"),
	}, nil
}

// Exists checks if the Path of this Config exists on the operating system as a directory.
func (c *Config) Exists() (bool, error) {
	fi, err := os.Stat(c.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	if fi.IsDir() {
		return true, nil
	}

	return false, errors.New(".clox already exists as a file in home directory")
}

// The parameters when writing the config.json file.
type WriteFileParams struct {
	Password string `json:"password"`
	APIToken string `json:"api_token"`
}

// Write will write the parameters to the config.json file. The config.json file will be
// stored within the Path of this Config.
func (c *Config) Write(p WriteFileParams) error {
	if err := os.Mkdir(c.Path, 0700); err != nil {
		return fmt.Errorf("failed creating directory %s: %w", c.Path, err)
	}

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed marshalling data to json: %w", err)
	}

	filePath := filepath.Join(c.Path, "config.json")
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed writing file %s: %w", filePath, err)
	}

	return nil
}
