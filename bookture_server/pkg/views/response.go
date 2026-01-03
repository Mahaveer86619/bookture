package views

import "github.com/labstack/echo/v4"

type Response interface {
	SetStatusCode(int)
	SetMessage(string)
	SetData(any)
	JSON(c echo.Context) error
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

// JSON writes the Success response using Echo Context
func (s *Success) JSON(c echo.Context) error {
	return c.JSON(s.StatusCode, s)
}

// Setters for Failure
func (f *Failure) SetStatusCode(statusCode int) {
	f.StatusCode = statusCode
}

func (f *Failure) SetMessage(message string) {
	f.Message = message
}

// JSON writes the Failure response using Echo Context
func (f *Failure) JSON(c echo.Context) error {
	return c.JSON(f.StatusCode, f)
}
