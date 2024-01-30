package cmd

import (
	"fmt"
	"os"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/prompt"
	"github.com/cicconee/clox-cli/internal/security"
	"github.com/spf13/cobra"
)

// InitCommand is the 'init' command.
//
// InitCommand will create the user configuration and write it to the config file.
type InitCommand struct {
	cmd   *cobra.Command
	store *config.Store
	keys  *security.Keys
	aes   *crypto.AES
	user  *config.User
	force bool
}

// NewInitCommand creates and returns a InitCommand.
//
// A force flag '-f', is set for the InitCommand. This flag allows users to overwrite
// their current configuration if already set.
func NewInitCommand(store *config.Store, keys *security.Keys, aes *crypto.AES) *InitCommand {
	initCmd := &InitCommand{store: store, keys: keys, aes: aes}

	initCmd.cmd = &cobra.Command{
		Use:   "init",
		Short: "Set up the Clox CLI",
		Args:  cobra.ExactArgs(0),
		Run:   initCmd.Run,
	}

	initCmd.cmd.Flags().BoolVarP(&initCmd.force, "force", "f", false, "Overwrites current configuration")

	return initCmd
}

// Command returns the cobra.Command of this InitCommand.
func (c *InitCommand) Command() *cobra.Command {
	return c.cmd
}

// Run is the Run function of the cobra.Command in this InitCommand.
//
// Run will create a user and write it to the configuration file. If the
// configuration directory does not exist it will create it. If the user is already
// configured, it will print a message stating Clox CLI is already set up.
func (c *InitCommand) Run(cmd *cobra.Command, args []string) {
	dirExists, err := c.store.DirExists()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !dirExists {
		err := c.store.WriteDir()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	user := &config.User{}
	err = c.store.ReadConfigFile(user)
	if err == nil && !c.force {
		fmt.Println("Clox CLI already configured")
		fmt.Println("Run 'clox init -f' to force initialize")
		os.Exit(0)
	}

	user, err = config.NewUser(
		c.keys,
		c.aes,
		prompt.ConfigurePassowrd(),
		prompt.ConfigureAPIToken())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if err := c.store.WriteConfigFile(user); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success")
	os.Exit(0)
}
