package errz

import (
	"errors"
	"log"
	"net/http"

	"github.com/Mahaveer86619/bookture/server/pkg/views"
)

type ErrzType string

const (
	BadRequest          ErrzType = "bad_request"
	NotFound            ErrzType = "not_found"
	Conflict            ErrzType = "conflict"
	Unauthorized        ErrzType = "unauthorized"
	Forbidden           ErrzType = "forbidden"
	InternalServerError ErrzType = "internal_server_error"
)

var statusMap = map[ErrzType]int{
	BadRequest:          http.StatusBadRequest,
	NotFound:            http.StatusNotFound,
	Conflict:            http.StatusConflict,
	Unauthorized:        http.StatusUnauthorized,
	Forbidden:           http.StatusForbidden,
	InternalServerError: http.StatusInternalServerError,
}

type BooktureError struct {
	Type    ErrzType
	Message string
	Err     error
}

func (e *BooktureError) Error() string {
	return e.Message
}

func New(errType ErrzType, msg string, err error) error {
	return &BooktureError{
		Type:    errType,
		Message: msg,
		Err:     err,
	}
}

func HandleErrors(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var bkErr *BooktureError
	if errors.As(err, &bkErr) {
		statusCode := statusMap[bkErr.Type]
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}

		if statusCode == http.StatusInternalServerError {
			log.Printf("Internal Error: %v | Source: %v", bkErr.Message, bkErr.Err)
		}

		resp := &views.Failure{
			StatusCode: statusCode,
			Message:    bkErr.Message,
		}
		_ = resp.JSON(w)
		return
	}

	log.Printf("Unknown Error: %v", err)
	resp := &views.Failure{
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal Server Error",
	}
	_ = resp.JSON(w)
}
