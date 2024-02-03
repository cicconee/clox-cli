package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

// NewDirRequest is the body of the POST request when creating a new directory.
type NewDirRequest struct {
	// The name of the directory being created.
	Name string `json:"name"`
}

// NewDirResponse is the response body of the POST request when creating a new
// directory.
type NewDirResponse struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	ParentID  string    `json:"parent_id"`
	DirName   string    `json:"directory_name"`
	DirPath   string    `json:"directory_path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastWrite time.Time `json:"last_write"`
}

// ErrorResponse is the response body when the API server responds with an error.
// Every error response from the server conforms to this structure.
type ErrorResponse struct {
	Err        string `json:"error"`
	StatusCode int    `json:"status_code"`
}

// APIError is a custom error type that represents an HTTP error response from the
// API.
//
// APIError satisfies the error interface.
type APIError struct {
	Err        string
	StatusCode int
}

// The function that satisfies the error interface.
func (e *APIError) Error() string {
	return e.Err
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

	data := &NewDirRequest{Name: args[0]}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error: Marshal:", err)
		return
	}

	var req *http.Request
	var rErr error
	if c.path != "" || (c.path == "" && c.id == "") {
		req, rErr = c.newPathRequest(bytes.NewBuffer(jsonData), token)
	} else {
		req, rErr = c.newIDRequest(bytes.NewBuffer(jsonData), token)
	}
	if rErr != nil {
		fmt.Println("Error: Request:", rErr)
		return
	}
	defer req.Body.Close()

	// Create the HTTP client and do the request.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: Sending Request:", err)
		return
	}
	defer res.Body.Close()

	respData := &NewDirResponse{}
	err = api.ParseResponse(res, respData)
	if err != nil {
		switch e := err.(type) {
		case *APIError:
			fmt.Printf("API Error [%d]: %s\n", e.StatusCode, e.Err)
			fmt.Printf("-> [ARG] Name: %s\n", args[0])
			fmt.Printf("-> [FLAG] Path: %s\n", c.path)
			fmt.Printf("-> [FLAG] Parent ID: %s\n", c.id)
		default:
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	fmt.Printf("API [%d]: Directory Created\n", res.StatusCode)
	fmt.Printf("-> Name: %s\n", respData.DirName)
	fmt.Printf("-> Path: %s\n", respData.DirPath)
	fmt.Printf("-> ID: %s\n", respData.ID)
	return
}

func (c *MkdirCommand) newPathRequest(body *bytes.Buffer, token string) (*http.Request, error) {
	return api.NewRequest(api.RequestParams{
		Method: "POST",
		URL:    "http://localhost:8081/api/dir",
		Body:   body,
		Token:  token,
		Query:  map[string]string{"path": c.path},
	})
}

func (c *MkdirCommand) newIDRequest(body *bytes.Buffer, token string) (*http.Request, error) {
	return api.NewRequest(api.RequestParams{
		Method: "POST",
		URL:    fmt.Sprintf("http://localhost:8081/api/dir/%s", c.id),
		Body:   body,
		Token:  token,
	})
}
