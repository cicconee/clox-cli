package api

import "encoding/json"

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

// parseErrorResponse will unmarshal an API error response and return it as a
// *APIError. The JSON in the []byte is unmarshalled into an ErrorResponse. The
// ErrorResponse is then used to construct and return a *APIError.
//
// If unmarshalling the []byte fails, it will still return a *APIError, but the
// Err field will specify that parsing the API error response failed. If this ever
// happens, most likely the server is responding with invalid data and something is
// wrong.
func ParseErrorResponse(b []byte, statusCode int) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(b, &errResp); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Err:        "Failed to parse API error response"}
	}

	return &APIError{Err: errResp.Err, StatusCode: errResp.StatusCode}
}
