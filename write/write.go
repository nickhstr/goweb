// Package write provides http-related functions, types, utilities, etc.
// The package name was deliberately chosen so as to not conflict with net/http,
// as these two packages are bound to be used together.
package write

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse provides a standard error response format.
type ErrorResponse struct {
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
	Error      string `json:"error"`
}

// Error writes an HTTP JSON error response.
func Error(w http.ResponseWriter, err string, code int) {
	errResponse := ErrorResponse{
		code,
		http.StatusText(code),
		err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	e := json.NewEncoder(w)
	_ = e.Encode(errResponse)
}

type OKResponse struct {
	Status     int    `json:"status"`
	StatusText string `json:"statusText"`
}

// OK writes an HTTP JSON ok response.
func OK(w http.ResponseWriter) {
	okResponse := OKResponse{
		http.StatusOK,
		http.StatusText(http.StatusOK),
	}

	EncodeJSON(w, okResponse)
}

// EncodeJSON conveniently writes JSON data responses.
func EncodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	err := e.Encode(v)

	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
	}
}
