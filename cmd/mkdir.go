package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cicconee/clox-cli/internal/config"
	"github.com/cicconee/clox-cli/internal/crypto"
	"github.com/cicconee/clox-cli/internal/security"
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
	keys     *security.Keys
	aes      *crypto.AES
	path     string
	id       string
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
		os.Exit(0)
	}

	if c.path != "" {
		c.runPath(cmd, args)
	} else if c.id != "" {
		c.runID(cmd, args)
	} else {
		c.runPath(cmd, args)
	}
}

// runID creates a directory using the id (-i, --id) flag.
func (c *MkdirCommand) runID(cmd *cobra.Command, args []string) {
	fmt.Println("Create directory using parent ID")
}

// runPath creates a directory using the path (-p, --path) flag.
func (c *MkdirCommand) runPath(cmd *cobra.Command, args []string) {
	token, err := c.user.APIToken(c.aes, c.password)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	data := &NewDirRequest{Name: args[0]}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error: Marshal:", err)
		os.Exit(1)
	}

	// Create the request, add the token to the authorization header, and set the
	// "path" query parameter.
	url := "http://localhost:8081/api/dir"
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error: Creating Request:", err)
		os.Exit(1)
	}
	authHeader := fmt.Sprintf("Bearer %s", token)
	r.Header.Set("Authorization", authHeader)
	q := r.URL.Query()
	q.Set("path", c.path)
	r.URL.RawQuery = q.Encode()

	// Create the HTTP client and do the request.
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Println("Error: Sending Request:", err)
		os.Exit(1)
	}

	respData := &NewDirResponse{}
	err = HandleResponse(resp, respData)
	if err != nil {
		switch e := err.(type) {
		case *APIError:
			fmt.Printf("API Error [%d]: %s\n", e.StatusCode, e.Err)
			fmt.Printf("-> Name: %s\n", args[0])
			fmt.Printf("-> Path: %s\n", c.path)
			os.Exit(0)
		default:
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("API [%d]: Directory Created\n", resp.StatusCode)
	fmt.Printf("-> Name: %s\n", respData.DirName)
	fmt.Printf("-> Path: %s\n", respData.DirPath)
	fmt.Printf("-> ID: %s\n", respData.ID)
	return
}

// HandleResponse handles http.Response from the Clox API. A successful request will
// parse JSON body into dst.
//
// If the API responds with an error (non-200 status code), it will return an
// *APIError.
func HandleResponse(r *http.Response, dst any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return parseErrorResponse(body, r.StatusCode)
	}

	return parseResponse(body, dst)
}

// parseResponse will unmarshal the JSON in the []byte into dst.
//
// The []byte is expected to hold a valid JSON structure, if it does not, an error
// is returned.
func parseResponse(b []byte, dst any) error {
	err := json.Unmarshal(b, dst)
	if err != nil {
		return fmt.Errorf("unmarshalling body: %w", err)
	}

	return nil
}

// parseErrorResponse will unmarshal an API error response and return it as a
// *APIError. The JSON in the []byte is unmarshalled into an ErrorResponse. The
// ErrorResponse is then used to construct and return a *APIError.
//
// If unmarshalling the []byte fails, it will still return a *APIError, but the
// Err field will specify that parsing the API error response failed. If this ever
// happens, most likely the server is responding with invalid data and something is
// wrong.
func parseErrorResponse(b []byte, statusCode int) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(b, &errResp); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Err:        "Failed to parse API error response"}
	}

	return &APIError{Err: errResp.Err, StatusCode: errResp.StatusCode}
}
