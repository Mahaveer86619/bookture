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

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req views.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	if err := req.Valid(); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
	}

	resp, err := h.svc.Register(req.Email, req.Password, req.DisplayName)
	if err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	success := views.Success{StatusCode: http.StatusCreated, Data: resp, Message: "User created"}
	_ = success.JSON(w)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req views.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	if err := req.Valid(); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
	}

	resp, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		errz.HandleErrors(w, http.StatusUnauthorized, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "Login successful"}
	_ = success.JSON(w)
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req views.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	if err := req.Valid(); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	resp, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		errz.HandleErrors(w, http.StatusUnauthorized, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "Token refreshed"}
	_ = success.JSON(w)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	user, err := h.svc.GetUser(userID)
	if err != nil {
		errz.HandleErrors(w, http.StatusNotFound, err)
		return
	}
	success := views.Success{StatusCode: http.StatusOK, Data: user, Message: "User profile fetched"}
	_ = success.JSON(w)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctxUserID := middleware.GetUserID(r)

	var req views.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	reqId, err := utils.UnmaskID(req.ID)
	if err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	if ctxUserID != reqId {
		errz.HandleErrors(w, http.StatusForbidden, err)
		return
	}

	resp, err := h.svc.UpdateUser(reqId, req.DisplayName)
	if err != nil {
		errz.HandleErrors(w, http.StatusInternalServerError, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: resp, Message: "User profile updated"}
	_ = success.JSON(w)
}

func (h *UserHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctxUserID := middleware.GetUserID(r)

	queryID := r.URL.Query().Get("id")

	reqId, err := utils.UnmaskID(queryID)
	if err != nil {
		errz.HandleErrors(w, http.StatusBadRequest, err)
		return
	}

	if ctxUserID != reqId {
		errz.HandleErrors(w, http.StatusForbidden, err)
		return
	}

	err = h.svc.DeleteUser(ctxUserID)
	if err != nil {
		errz.HandleErrors(w, http.StatusInternalServerError, err)
		return
	}

	success := views.Success{StatusCode: http.StatusOK, Data: nil, Message: "User profile deleted"}
	_ = success.JSON(w)
}
