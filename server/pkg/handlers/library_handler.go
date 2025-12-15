package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/middleware"
	"github.com/Mahaveer86619/bookture/server/pkg/services"
	"github.com/Mahaveer86619/bookture/server/pkg/utils"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
)

type LibraryHandler struct {
	libService *services.LibraryService
}

func NewLibraryHander(libSvs *services.LibraryService) *LibraryHandler {
	return &LibraryHandler{
		libService: libSvs,
	}
}

func (h *LibraryHandler) CreateNewLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req views.CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, err)
		return
	}

	if err := req.Valid(); err != nil {
		errz.HandleErrors(w, err)
	}

	resp, err := h.libService.CreateNewLibrary(userID, req.Name)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "Created New Library"}
	_ = success.JSON(w)
}

func (h *LibraryHandler) GetLibraries(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	resp, err := h.libService.GetUserLibraries(userID)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "Libraries fetched"}
	_ = success.JSON(w)
}

func (h *LibraryHandler) GetLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	queryID := r.URL.Query().Get("id")
	if queryID == "" {
		errz.HandleErrors(w, errors.New("id is required"))
		return
	}

	libID, err := utils.UnmaskID(queryID)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	resp, err := h.libService.GetLibrary(libID, userID)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "Library fetched"}
	_ = success.JSON(w)
}

func (h *LibraryHandler) UpdateLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req views.UpdateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, err)
		return
	}

	if err := req.Valid(); err != nil {
		errz.HandleErrors(w, err)
		return
	}

	libID, err := utils.UnmaskID(req.ID)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	resp, err := h.libService.UpdateLibrary(libID, userID, req.Name)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "Library updated"}
	_ = success.JSON(w)
}

func (h *LibraryHandler) DeleteLibrary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	queryID := r.URL.Query().Get("id")
	if queryID == "" {
		errz.HandleErrors(w, errors.New("id is required"))
		return
	}

	libID, err := utils.UnmaskID(queryID)
	if err != nil {
		errz.HandleErrors(w, err)
		return
	}

	if err = h.libService.DeleteLibrary(libID, userID); err != nil {
		errz.HandleErrors(w, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: nil, Message: "Library deleted"}
	_ = success.JSON(w)
}
