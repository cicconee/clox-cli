package cmd

import (
	"fmt"
	"net/http"

	"github.com/cicconee/clox-cli/internal/api"
	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/spf13/cobra"
)

// The 'mkdir' command.
//
// MkdirCommand will create a new directory on the Clox server. The path flag is
// optional. If not provided directory will default to the users root.
type MkdirCommand struct {
	cmd      *cobra.Command
	user     *config.User
	password string
	aes      *crypto.AES
	path     string
	id       string
}

// NewInitCommand creates and returns a InitCommand.
//
// A force flag '-f', is set for the InitCommand. This flag allows users to overwrite
// their current configuration if already set.
func NewMkdirCommand(aes *crypto.AES) *MkdirCommand {
	mkdirCmd := &MkdirCommand{aes: aes}

	mkdirCmd.cmd = &cobra.Command{
		Use:   "mkdir <name>",
		Short: "Create a new directory",
		Args:  cobra.ExactArgs(1),
		Run:   mkdirCmd.Run,
	}

	mkdirCmd.cmd.Flags().StringVarP(&mkdirCmd.path, "path", "p", "", "The path where the directory will be created")
	mkdirCmd.cmd.Flags().StringVarP(&mkdirCmd.id, "id", "i", "", "The ID of the parent directory")

	return mkdirCmd
}

// Command returns the cobra.Command of this InitCommand.
func (c *MkdirCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *MkdirCommand) SetUser(user *config.User) {
	c.user = user
}

func (c *MkdirCommand) SetPassword(password string) {
	c.password = password
}

// Run is the Run function of the cobra.Command in this MkdirCommand.
//
// Run will create a new directory on the Clox server. The password is used to
// decrypt the API token, and then calls the API endpoint to create a directory.
//
// If the path flag (-p, --path) is set it will create a directory by specifying
// the path to the new directory. If the id flag (-i, --id) is set, it will create
// a directory by specifying the ID of the parent. If no flag is set, it will create
// the directory using an empty path. This will default to the users root directory.
func (c *MkdirCommand) Run(cmd *cobra.Command, args []string) {
	if c.path != "" && c.id != "" {
		fmt.Println("Only one flag can be set: path (-p, --path) or id (-i, --id)")
		return
	}

	token, err := c.user.APIToken(c.aes, c.password)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create the HTTP client and do the request.
	client := &http.Client{}
	dirParams := api.NewDirParams{
		BaseURL: "http://localhost:8081",
		DirName: args[0],
		Token:   token,
	}
	var res *api.NewDirResponse
	var rErr error
	if c.path != "" || (c.path == "" && c.id == "") {
		res, rErr = api.NewDirWithPath(client, c.path, dirParams)
	} else {
		res, rErr = api.NewDirWithID(client, c.id, dirParams)
	}
	if rErr != nil {
		switch e := rErr.(type) {
		case *api.APIError:
			fmt.Printf("API Error [%d]: %s\n", e.StatusCode, e.Err)
			fmt.Printf("-> [ARG] Name: %s\n", args[0])
			fmt.Printf("-> [FLAG] Path: %s\n", c.path)
			fmt.Printf("-> [FLAG] Parent ID: %s\n", c.id)
		default:
			fmt.Printf("Error: %v\n", rErr)
		}
		return
	}

	fmt.Printf("API [%d]: Directory Created\n", 200)
	fmt.Printf("-> Name: %s\n", res.DirName)
	fmt.Printf("-> Path: %s\n", res.DirPath)
	fmt.Printf("-> ID: %s\n", res.ID)
	return
}
