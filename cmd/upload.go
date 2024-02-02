package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

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

// UploadFileResponse is the result of a successful file upload. Each
// UploadFileResponse corresponds to a single file. This is a single entry within
// UploadResponse.Uploads.
type UploadFileResponse struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	DirectoryID string    `json:"directory_id"`
	Name        string    `json:"file_name"`
	Path        string    `json:"file_path"`
	Size        int64     `json:"file_size"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// UploadErrorResponse is the result of a failed file upload. Each
// UploadErrorResponse corresponds to a single file failure. This is a single
// entry within UploadResponse.Errors.
type UploadErrorResponse struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"file_size"`
	Error    string `json:"error"`
}

// UploadResponse is the response body of the POST request when uploading files.
type UploadResponse struct {
	Uploads []UploadFileResponse  `json:"uploads"`
	Errors  []UploadErrorResponse `json:"errors"`
}

// UploadInput is the input for uploading a file. A UploadInput corresponds to a
// single file to be uploaded with the 'upload' command.
type UploadInput struct {
	Index    int
	Path     string
	Filename string
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
	uploads := []UploadInput{}
	for i, a := range args {
		parts := strings.Split(a, ":")
		if len(parts) != 2 {
			fmt.Printf("Invalid syntax [Index: %d, Input: %s]: ", i, a)
			fmt.Println("Must be in format <file>:<name>")
			return
		}
		uploads = append(uploads, UploadInput{
			Index:    i,
			Path:     parts[0],
			Filename: parts[1],
		})
	}

	// Build the request body by reading each file on the file system into a
	// form file.
	var reqBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&reqBody)
	for _, u := range uploads {
		file, err := os.Open(u.Path)
		if err != nil {
			fmt.Printf("Error: Opening file [Path: %s, Index: %d]: %v", u.Path, u.Index, err)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			fmt.Printf("Error: Reading File [Path: %s, Index: %d]: %v\n",
				u.Path, u.Index, err)
			return
		}

		encData, err := c.aes.Encrypt(data, encryptKey)
		if err != nil {
			fmt.Printf("Error: Encrypting File [Path: %s, Index: %d]: %v\n",
				u.Path, u.Index, err)
			return
		}

		formFile, err := multipartWriter.CreateFormFile("file_uploads", u.Filename)
		if err != nil {
			fmt.Printf("Error: Creating form file [Filename: %s, Index: %d]: %v", u.Filename, u.Index, err)
			return
		}

		if _, err := io.Copy(formFile, bytes.NewReader(encData)); err != nil {
			fmt.Printf("Error: Copying file [Filename: %s, Index: %d]: %v", u.Filename, u.Index, err)
			return
		}
	}
	multipartWriter.Close()

	// Build the request with the request body.
	req, err := NewAuthRequest(AuthRequestParams{
		Method: "POST",
		URL:    "http://localhost:8081/api/upload",
		Body:   &reqBody,
		Token:  token,
	})
	if err != nil {
		fmt.Println("Error: Request:", err)
		return
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Do the request and parse the response.
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: Sending Request:", err)
		return
	}
	defer res.Body.Close()

	respData := UploadResponse{}
	err = HandleResponse(res, &respData)
	if err != nil {
		switch e := err.(type) {
		case *APIError:
			fmt.Printf("API Error [%d]: %s\n", e.StatusCode, e.Err)
			fmt.Printf("-> [ARG] Name: %s\n", args[0])
			fmt.Printf("-> [FLAG] Path: %s\n", c.path)
			fmt.Printf("-> [FLAG] Directory ID: %s\n", c.id)
		default:
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	uploadCount := len(respData.Uploads)
	fmt.Printf("\nUploaded: %d\n", uploadCount)
	for _, u := range respData.Uploads {
		fmt.Printf("%s -> %s\n", u.ID, u.Path)
	}

	errorCount := len(respData.Errors)
	fmt.Printf("\nErrors: %d\n", errorCount)
	for _, e := range respData.Errors {
		fmt.Printf("%s -> %s\n", e.FileName, e.Error)
	}
}
