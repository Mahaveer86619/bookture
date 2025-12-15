package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/middleware"
	"github.com/Mahaveer86619/bookture/server/pkg/services"
	"github.com/Mahaveer86619/bookture/server/pkg/utils"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
)

type BookHandler struct {
	svc *services.BookService
}

func NewBookHandler(svc *services.BookService) *BookHandler {
	return &BookHandler{svc: svc}
}

func (h *BookHandler) CreateDraft(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req views.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, err)
		return
	}

	if err := req.Valid(); err != nil {
		errz.HandleErrors(w, err)
		return
	}

	libID, err := utils.UnmaskID(req.LibraryID)
	if err != nil {
		errz.HandleErrors(w, errz.New(errz.BadRequest, "Invalid Library ID", err))
		return
	}

	if req.Title == "" {
		req.Title = "Untitled draft"
	}
	if req.Author == "" {
		req.Author = "Unknown"
	}
	if req.Description == "" {
		req.Description = "No description provided"
	}

	resp, err := h.svc.CreateDraftBook(userID, libID, req.Title, req.Author, req.Description)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusCreated, Data: resp, Message: "Draft book created"}
	_ = success.JSON(w)
}

func (h *BookHandler) UploadVolume(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	queryID := r.URL.Query().Get("book_id")
	if queryID == "" {
		errz.HandleErrors(w, errz.New(errz.BadRequest, "book_id is required", nil))
		return
	}

	bookID, err := utils.UnmaskID(queryID)
	if err != nil {
		errz.HandleErrors(w, errz.New(errz.BadRequest, "Invalid Book ID", err))
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		errz.HandleErrors(w, errz.New(errz.BadRequest, "File too large or invalid format", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		errz.HandleErrors(w, errz.New(errz.BadRequest, "Failed to retrieve file", err))
		return
	}
	defer file.Close()

	resp, err := h.svc.UploadVolume(userID, bookID, header)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusCreated, Data: resp, Message: "Volume uploaded successfully"}
	_ = success.JSON(w)
}
