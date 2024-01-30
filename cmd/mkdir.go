package cmd

import (
	"fmt"
	"os"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/security"
	"github.com/spf13/cobra"
)

// The 'mkdir' command.
//
// mkdirCmd will create a new directory on the Clox server. The path flag is optional.
// If not provided directory will default to the users root.
//
// TODO:
//   - Call Clox API endpoint to create directory.
type MkdirCommand struct {
	cmd      *cobra.Command
	user     *config.User
	password string
	keys     *security.Keys
	aes      *crypto.AES
	path     string
}

// NewInitCommand creates and returns a InitCommand.
//
// A force flag '-f', is set for the InitCommand. This flag allows users to overwrite
// their current configuration if already set.
func NewMkdirCommand(keys *security.Keys, aes *crypto.AES) *MkdirCommand {
	mkdirCmd := &MkdirCommand{keys: keys, aes: aes}

	mkdirCmd.cmd = &cobra.Command{
		Use:   "mkdir <name>",
		Short: "Create a new directory",
		Args:  cobra.ExactArgs(1),
		Run:   mkdirCmd.Run,
	}

	mkdirCmd.cmd.Flags().StringVarP(&mkdirCmd.path, "path", "p", "", "The path where the directory will be created")

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

// Run is the Run function of the cobra.Command in this InitCommand.
//
// Run will create a user and write it to the configuration file. If the
// configuration directory does not exist it will create it. If the user is already
// configured, it will print a message stating Clox CLI is already set up.
func (c *MkdirCommand) Run(cmd *cobra.Command, args []string) {
	fmt.Println("User:", c.user)
	fmt.Println("Dir Name:", args[0])
	fmt.Println("Path:", c.path)
	fmt.Println("Password:", c.password)

	token, err := c.user.APIToken(c.aes, c.password)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Token:", token)

	os.Exit(0)
}
