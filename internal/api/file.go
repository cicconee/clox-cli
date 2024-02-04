package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/cicconee/clox-cli/internal/crypto"
)

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

// FileUpload represents a file to be read, encrypted, and written to the server.
type FileUpload struct {
	// The local path to the file. This is the path to the file that will be
	// uploaded.
	Path string
	// The file name for the encrypted file on the server. The contents of this
	// file will be the encrypted contents of the file defined in Path.
	Filename string
}

// UploadParams is the parameters needed when uploading files.
type UploadParams struct {
	// The base URL for the API.
	BaseURL string
	// The users API token.
	Token string
	// The file(s) metadata.
	Uploads []FileUpload
	// The encryption key for encrypting the files.
	Key []byte
	// The encryption algorithm used to encrypt.
	Alg *crypto.AES
}

// UploadWithPath calls the API to upload files using a path. The path parameter is
// the path that the files will be written to. This parameter is optional and if
// empty will upload the files to the users root directory on the server.
//
// Every file that is uploaded will be encrypted with the encryption key
// (UploadParams.Key) using the encryption algorithm (UploadParams.Alg).
//
// The files to be uploaded are defined in UploadParams.Uploads. Each FileUpload
// represents a file that will be read, encrypted, and uploaded. The Path is the
// location on the local machine, and Filename is the name of the encrypted file
// to be written to the server.
//
// If the API responds with an error (non-200 status code), it will return nil and
// an *APIError.
func UploadWithPath(client *http.Client, path string, p UploadParams) (*UploadResponse, error) {
	return upload(client, uploadConfig{
		UploadParams: p,
		URLPath:      "api/upload",
		Query:        map[string]string{"path": path},
	})
}

// UploadWithID calls the API to upload files using a directory ID. The id parameter
// is the ID of the directory that the files will be written to.
//
// Every file that is uploaded will be encrypted with the encryption key
// (UploadParams.Key) using the encryption algorithm (UploadParams.Alg).
//
// The files to be uploaded are defined in UploadParams.Uploads. Each FileUpload
// represents a file that will be read, encrypted, and uploaded. The Path is the
// location on the local machine, and Filename is the name of the encrypted file
// to be written to the server.
//
// If the API responds with an error (non-200 status code), it will return nil and
// an *APIError.
func UploadWithID(client *http.Client, id string, p UploadParams) (*UploadResponse, error) {
	return upload(client, uploadConfig{
		UploadParams: p,
		URLPath:      fmt.Sprintf("api/upload/%s", id),
	})
}

// uploadConfig is the configuration for calling the Clox API to upload files.
type uploadConfig struct {
	UploadParams
	URLPath string
	Query   map[string]string
}

// upload uploads files by calling the Clox API.
func upload(client *http.Client, c uploadConfig) (*UploadResponse, error) {
	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	for i, u := range c.Uploads {
		path := u.Path
		filename := u.Filename

		// Build the request body by reading each file on the file system,
		// encrypt the data, and write to the form file.
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening '%s' [index: %d]: %w", path, i, err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("reading '%s' [index: %d]: %w", path, i, err)
		}

		encData, err := c.Alg.Encrypt(data, c.Key)
		if err != nil {
			return nil, fmt.Errorf("encrypting '%s' [index: %d]: %w", path, i, err)
		}

		formFile, err := writer.CreateFormFile("file_uploads", filename)
		if err != nil {
			return nil, fmt.Errorf("creating form file '%s' [index: %d, name: %s]: %w",
				path, i, filename, err)
		}

		if _, err := io.Copy(formFile, bytes.NewReader(encData)); err != nil {
			return nil, fmt.Errorf("copying file '%s' [index: %d, name: %s]: %w",
				path, i, filename, err)
		}
	}
	writer.Close()

	req, err := NewRequest(RequestParams{
		Method: "POST",
		URL:    fmt.Sprintf("%s/%s", c.BaseURL, c.URLPath),
		Body:   &reqBody,
		Token:  c.Token,
		Query:  c.Query,
		Header: map[string]string{"Content-Type": writer.FormDataContentType()},
	})
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	defer req.Body.Close()

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	respData := &UploadResponse{}
	err = ParseResponse(res, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}
