package types

import (
	"encoding/json"
)

// ErrorResponse provides the structure for the error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewErrorResponse returns data for an error response
func NewErrorResponse(e error) []byte {
	errResp := ErrorResponse{e.Error()}
	response, err := json.Marshal(errResp)

	if err != nil {
		response = []byte("Unable to create error response")
	}

	return response
}
