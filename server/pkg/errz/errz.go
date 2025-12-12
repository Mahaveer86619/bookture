package errz

import (
	"log"
	"net/http"

	"github.com/Mahaveer86619/bookture/server/pkg/views"
)

func HandleErrors(w http.ResponseWriter, statusCode int, err error) {
	msg := "Unknown error"
	if err != nil {
		msg = err.Error()
	}

	if statusCode == http.StatusInternalServerError {
		log.Printf("Internal Server Error: %v", err)
		msg = "Internal Server Error"
	}

	resp := &views.Failure{
		StatusCode: statusCode,
		Message:    msg,
	}

	_ = resp.JSON(w)
}
