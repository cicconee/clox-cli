package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cicconee/clox-cli/internal/api"
	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/security"
	"github.com/spf13/cobra"
)

// The 'upload' command.
//
// UploadCommand encrypts and uploads files to the Clox server. Both the path and id
// flag are optional, but they can't be used together. If no path or id flag is
// provided, files will be uploaded to the users root directory.
type UploadCommand struct {
	cmd      *cobra.Command
	user     *config.User
	password string
	keys     *security.Keys
	aes      *crypto.AES
	rsa      *crypto.RSA
	path     string
	id       string
}

// NewUploadCommand creates and returns a UploadCommand.
//
// The path flag (-p, --path) is set for the UploadCommand. This flag allows users
// to specify the directory to upload files.
//
// The id flag (-i, --id) is set for the UploadCommand. This flag allows users to
// specify the directory ID to upload files to.
//
// If neither a path or id flag is set, the files will upload to the users root
// directory by default. The path and id flags cannot be used together.
func NewUploadCommand(keys *security.Keys, aes *crypto.AES, rsa *crypto.RSA) *UploadCommand {
	uploadCmd := &UploadCommand{keys: keys, aes: aes, rsa: rsa}

	uploadCmd.cmd = &cobra.Command{
		Use:   "upload <file1>:<name1> [<file2>:<name2>...]",
		Short: "Upload files to the server",
		Args:  cobra.MinimumNArgs(1),
		Run:   uploadCmd.Run,
	}

	uploadCmd.cmd.Flags().StringVarP(&uploadCmd.path, "path", "p", "", "The path to upload the files")
	uploadCmd.cmd.Flags().StringVarP(&uploadCmd.id, "id", "i", "", "The ID of the directory to upload the files")

	return uploadCmd
}

func (c *UploadCommand) Command() *cobra.Command {
	return c.cmd
}

func (c *UploadCommand) SetUser(user *config.User) {
	c.user = user
}

func (c *UploadCommand) SetPassword(password string) {
	c.password = password
}

// Run is the Run function of the cobra.Command in this UploadCommand.
//
// Run will upload files to the Clox server. Users specify the file to upload and
// the name for the file to be stored on the server. The format is <file>:<name>
// where <file> is the path to the local file and <name> is the name to be used to
// store the file on the server. There is no limit on how many file-name pairs can
// be set. The password is used to decrypt the API token, and then calls the API
// endpoint to upload files.
//
// If the path flag (-p, --path) is set, it will upload files to specified directory.
// If the id flag (-i, --id) is set, it will upload files to the directory with the
// specified ID. If no flag is set, it will upload files using an empty path. This
// will default to the users root directory.
func (c *UploadCommand) Run(cmd *cobra.Command, args []string) {
	if c.path != "" && c.id != "" {
		fmt.Println("Only one flag can be set: path (-p, --path) or id (-i, --id)")
		return
	}

	token, err := c.user.APIToken(c.aes, c.password)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	encryptKey, err := c.user.EncryptKey(c.keys, c.rsa, c.password)
	if err != nil {
		fmt.Println("Error: Getting Encryption Key:", err)
		return
	}

	// Parse the <file>:<name> args.
	uploads := []api.FileUpload{}
	for i, a := range args {
		parts := strings.Split(a, ":")
		if len(parts) != 2 {
			fmt.Printf("Invalid syntax [Index: %d, Input: %s]: ", i, a)
			fmt.Println("Must be in format <file>:<name>")
			return
		}
		uploads = append(uploads, api.FileUpload{Path: parts[0], Filename: parts[1]})
	}

	// Create the HTTP client and do the request.
	client := &http.Client{}
	uploadParams := api.UploadParams{
		BaseURL: "http://localhost:8081",
		Token:   token,
		Uploads: uploads,
		Key:     encryptKey,
		Alg:     c.aes,
	}
	var res *api.UploadResponse
	var rErr error
	if c.path != "" || (c.path == "" && c.id == "") {
		res, rErr = api.UploadWithPath(client, c.path, uploadParams)
	} else {
		res, rErr = api.UploadWithID(client, c.id, uploadParams)
	}
	if rErr != nil {
		switch e := rErr.(type) {
		case *api.APIError:
			fmt.Printf("API Error [%d]: %s\n", e.StatusCode, e.Err)
			fmt.Printf("-> [ARGS] Uploads: %v\n", args)
			fmt.Printf("-> [FLAG] Path: %s\n", c.path)
			fmt.Printf("-> [FLAG] Directory ID: %s\n", c.id)
		default:
			fmt.Printf("Error: %v\n", rErr)
		}
		return
	}

	fmt.Printf("\nUploaded: %d\n", len(res.Uploads))
	for _, u := range res.Uploads {
		fmt.Printf("%s -> %s\n", u.ID, u.Path)
	}

	fmt.Printf("\nErrors: %d\n", len(res.Errors))
	for _, e := range res.Errors {
		fmt.Printf("%s -> %s\n", e.FileName, e.Error)
	}
}
