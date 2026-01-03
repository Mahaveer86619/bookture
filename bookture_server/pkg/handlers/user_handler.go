package handlers

import (
	"net/http"

	"github.com/Mahaveer86619/bookture/pkg/errz"
	"github.com/Mahaveer86619/bookture/pkg/middleware"
	"github.com/Mahaveer86619/bookture/pkg/services"
	"github.com/Mahaveer86619/bookture/pkg/views"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID := middleware.GetUserID(c.Request())
	if userID == uuid.Nil {
		return errz.New(errz.Unauthorized, "User not authenticated", nil)
	}

	user, err := h.Service.GetProfile(userID)
	if err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Profile fetched",
		Data:       views.ToUserResponse(user),
	}
	return resp.JSON(c)
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	userID := middleware.GetUserID(c.Request())
	var req views.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return errz.New(errz.BadRequest, "Invalid request", err)
	}

	user, err := h.Service.UpdateProfile(userID, req)
	if err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Profile updated",
		Data:       views.ToUserResponse(user),
	}
	return resp.JSON(c)
}

func (h *UserHandler) GetProfile(c echo.Context) error {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return errz.New(errz.BadRequest, "Invalid User UUID", err)
	}

	user, err := h.Service.GetProfile(userID)
	if err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "User found",
		Data:       views.ToUserResponse(user),
	}
	return resp.JSON(c)
}

func (h *UserHandler) Follow(c echo.Context) error {
	followerID := middleware.GetUserID(c.Request())
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errz.New(errz.BadRequest, "Invalid Target UUID", err)
	}

	if err := h.Service.FollowUser(followerID, targetID); err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Followed successfully",
	}
	return resp.JSON(c)
}

func (h *UserHandler) Unfollow(c echo.Context) error {
	followerID := middleware.GetUserID(c.Request())
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errz.New(errz.BadRequest, "Invalid Target UUID", err)
	}

	if err := h.Service.UnfollowUser(followerID, targetID); err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Unfollowed successfully",
	}
	return resp.JSON(c)
}

func (h *UserHandler) GetFollowers(c echo.Context) error {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errz.New(errz.BadRequest, "Invalid UUID", err)
	}

	users, err := h.Service.GetFollowers(targetID)
	if err != nil {
		return err
	}

	// Map to views
	var userViews []views.UserResponse
	for _, u := range users {
		userViews = append(userViews, views.ToUserResponse(&u))
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Data:       userViews,
	}
	return resp.JSON(c)
}

func (h *UserHandler) GetFollowing(c echo.Context) error {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errz.New(errz.BadRequest, "Invalid UUID", err)
	}

	users, err := h.Service.GetFollowing(targetID)
	if err != nil {
		return err
	}

	var userViews []views.UserResponse
	for _, u := range users {
		userViews = append(userViews, views.ToUserResponse(&u))
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Data:       userViews,
	}
	return resp.JSON(c)
}
