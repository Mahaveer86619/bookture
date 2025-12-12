package views

import (
	"encoding/json"
	"net/http"
)

type Response interface {
	SetStatusCode(int)
	SetMessage(string)
	SetData(any)
	JSON(w http.ResponseWriter) error
}

type Success struct {
	StatusCode int    `json:"status_code"`
	Data       any    `json:"data,omitempty"`
	Message    string `json:"message"`
}

type Failure struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// Setters for Success
func (s *Success) SetStatusCode(statusCode int) {
	s.StatusCode = statusCode
}

func (s *Success) SetMessage(message string) {
	s.Message = message
}

func (s *Success) SetData(data any) {
	s.Data = data
}

// JSON writes the Success response to the http.ResponseWriter
func (s *Success) JSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s.StatusCode)
	return json.NewEncoder(w).Encode(s)
}

// Setters for Failure
func (f *Failure) SetStatusCode(statusCode int) {
	f.StatusCode = statusCode
}

func (f *Failure) SetMessage(message string) {
	f.Message = message
}

// JSON writes the Failure response to the http.ResponseWriter
func (f *Failure) JSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(f.StatusCode)
	return json.NewEncoder(w).Encode(f)
}
