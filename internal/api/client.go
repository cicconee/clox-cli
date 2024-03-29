package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client makes requests to the Clox API. Client should be created using the
// NewClient function.
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

// NewClient creates a *Client.
func NewClient(http *http.Client, baseURL string, token string) *Client {
	return &Client{http: http, baseURL: baseURL, token: token}
}

// RequestParams is the parameters when creating a new request. The Query and Header
// field is optional.
type RequestParams struct {
	Method string
	URL    string
	Body   *bytes.Buffer
	Token  string
	Query  map[string]string
	Header map[string]string
}

// NewRequest creates a new *http.Request that is configured with RequestParams.
func NewRequest(p RequestParams) (*http.Request, error) {
	r, err := http.NewRequest(p.Method, p.URL, p.Body)
	if err != nil {
		return nil, err
	}
	authHeader := fmt.Sprintf("Bearer %s", p.Token)
	r.Header.Set("Authorization", authHeader)

	if p.Query != nil && len(p.Query) > 0 {
		q := r.URL.Query()
		for k, v := range p.Query {
			q.Set(k, v)
		}
		r.URL.RawQuery = q.Encode()
	}

	for k, v := range p.Header {
		r.Header.Set(k, v)
	}

	return r, nil
}

// DoRequest creates and executes a *http.Request that is configured with
// RequestParams. The response is parsed into dst.
//
// If the API responds with an error (non-200 status code), it will return an
// *APIError.
func DoRequest(client *http.Client, dst any, p RequestParams) error {
	req, err := NewRequest(p)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	defer req.Body.Close()

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	err = ParseResponse(res, dst)
	if err != nil {
		return err
	}

	return nil
}

// HandleResponse handles *http.Response from the Clox API. A successful request will
// parse JSON body into dst.
//
// If the API responds with an error (non-200 status code), it will return an
// *APIError.
func ParseResponse(r *http.Response, dst any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	if r.StatusCode != 200 {
		return ParseErrorResponse(body, r.StatusCode)
	}

	err = json.Unmarshal(body, dst)
	if err != nil {
		return fmt.Errorf("unmarshalling body: %w", err)
	}

	return nil
}
