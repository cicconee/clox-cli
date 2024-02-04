package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NewDirParams is the parameters needed when creating a new directory.
type NewDirParams struct {
	// The base URL for the API.
	BaseURL string
	// The name of the directory being created.
	DirName string
	// The users API token.
	Token string
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

// newDirRequestBody is the request body of the POST request when creating a new
// directory.
type newDirRequestBody struct {
	Name string `json:"name"`
}

// NewDirWithPath calls the API to create a new directory by specifying the parent
// directory path. The path parameter is the path that the directory will be created
// within. This parameter is optional and if empty will create the directory in the
// users root directory on the server.
//
// If the API responds with an error (non-200 status code), it will return nil and
// an *APIError.
func NewDirWithPath(client *http.Client, path string, p NewDirParams) (*NewDirResponse, error) {
	return newDir(client, newDirConfig{
		NewDirParams: p,
		URLPath:      "api/dir",
		Query:        map[string]string{"path": path}})
}

// NewDirWithID calls the API to create a new directory by specifying the parent
// directory ID. The id parameter is the ID of the directory that the new directory
// will be created in (the parent directory).
//
// If the API responds with an error (non-200 status code), it will return nil and
// an *APIError.
func NewDirWithID(client *http.Client, id string, p NewDirParams) (*NewDirResponse, error) {
	return newDir(client, newDirConfig{
		NewDirParams: p,
		URLPath:      fmt.Sprintf("api/dir/%s", id),
	})
}

// newDirConfig is the configuration for calling the Clox API to create a directory.
type newDirConfig struct {
	NewDirParams
	URLPath string
	Query   map[string]string
}

// newDir creates a new directory by calling the Clox API.
func newDir(client *http.Client, c newDirConfig) (*NewDirResponse, error) {
	reqBody := newDirRequestBody{Name: c.DirName}
	jsonData, err := json.Marshal(&reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshalling data: %w", err)
	}

	req, err := NewRequest(RequestParams{
		Method: "POST",
		URL:    fmt.Sprintf("%s/%s", c.BaseURL, c.URLPath),
		Body:   bytes.NewBuffer(jsonData),
		Token:  c.Token,
		Query:  c.Query,
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

	respData := &NewDirResponse{}
	err = ParseResponse(res, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}
