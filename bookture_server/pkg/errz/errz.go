package errz

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mahaveer86619/bookture/pkg/views"
	"github.com/labstack/echo/v4"
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

func HandleErrors(err error, c echo.Context) {
    if err == nil {
        return
    }

	if c.Response().Committed {
		return
	}

    var bkErr *BooktureError
    if errors.As(err, &bkErr) {
        statusCode := statusMap[bkErr.Type]
        if statusCode == 0 {
            statusCode = http.StatusInternalServerError
        }

        resp := &views.Failure{
            StatusCode: statusCode,
            Message:    bkErr.Message,
        }
        _ = resp.JSON(c)
        return
    }

    // Standard Echo Errors (like 404 or 405)
    var echoErr *echo.HTTPError
    if errors.As(err, &echoErr) {
        resp := &views.Failure{
            StatusCode: echoErr.Code,
            Message:    fmt.Sprintf("%v", echoErr.Message),
        }
        _ = resp.JSON(c)
        return
    }

    resp := &views.Failure{
        StatusCode: http.StatusInternalServerError,
        Message:    "Internal Server Error",
    }
    _ = resp.JSON(c)
}
