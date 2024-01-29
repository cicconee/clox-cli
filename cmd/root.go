package cmd

import (
	"fmt"
	"os"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/security"
	"github.com/spf13/cobra"
)

// Command is the interface that wraps the Command function.
type Command interface {
	// Command returns the cobra.Command.
	Command() *cobra.Command
}

// CommandSetter is the interface that wraps the Command and SetUser functions.
type CommandSetter interface {
	Command

	// SetUser sets the config.User for a command.
	SetUser(*config.User)
}

// The root command of Clox CLI.
type RootCommand struct {
	store   *config.Store
	cmd     *cobra.Command
	subCmds map[string]CommandSetter
}

// NewRootCommand creates and returns a RootCommand.
func NewRootCommand(store *config.Store) *RootCommand {
	rootCmd := &RootCommand{
		store:   store,
		subCmds: map[string]CommandSetter{},
	}

	rootCmd.cmd = &cobra.Command{
		Use:              "clox",
		Short:            "The official client of the Clox API",
		SilenceErrors:    true,
		PersistentPreRun: rootCmd.PersistentPreRun,
	}

	return rootCmd
}

// AddCommand adds a *cobra.Command to this RootCommand.
func (c *RootCommand) AddCommand(cmd Command) {
	c.cmd.AddCommand(cmd.Command())
}

// AddCommand adds a *cobra.Command to this RootCommand and sets
// the CommandSetter in the subCmds map.
//
// The PersistentPreRun will initialize variables that are used through out all the
// sub commands of this RootCommand. Only commands set with this method will be
// passed these variables. This method is what enables global-free variables.
func (c *RootCommand) AddCommandSetter(cs CommandSetter) {
	cmd := cs.Command()
	c.cmd.AddCommand(cmd)
	c.subCmds[cmd.Name()] = cs
}

// PersistentPreRun is the PersistentPreRun of the cobra.Command in this
// RootCommand.
func (c *RootCommand) PersistentPreRun(cmd *cobra.Command, args []string) {
	if cmd.Name() != "init" {
		user := &config.User{}
		err := c.store.ReadConfigFile(user)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		c.subCmds[cmd.Name()].SetUser(user)
	}
}

// Execute creates the Clox CLI commands and executes the root command.
func Execute() {
	s, err := config.NewStore()
	if err != nil {
		fmt.Printf("Error: Failed initializing the configuration: %v\n", err)
		os.Exit(1)
	}

	keys := &security.Keys{AES: &crypto.AES{}}

	root := NewRootCommand(s)
	root.AddCommand(NewInitCommand(s, keys))

	if err := root.cmd.Execute(); err != nil {
		fmt.Printf("\n[ERROR] %v\n", err)
	}
}
