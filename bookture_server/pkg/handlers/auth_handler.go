package handlers

import (
	"net/http"

	"github.com/Mahaveer86619/bookture/pkg/errz"
	"github.com/Mahaveer86619/bookture/pkg/services"
	"github.com/Mahaveer86619/bookture/pkg/views"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Service services.AuthService
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{
		Service: service,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req views.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errz.New(errz.BadRequest, "Invalid request body", err)
	}

	user, err := h.Service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusCreated,
		Message:    "User registered successfully",
		Data:       views.ToUserResponse(user),
	}
	return resp.JSON(c)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req views.LoginRequest
	if err := c.Bind(&req); err != nil {
		return errz.New(errz.BadRequest, "Invalid request body", err)
	}

	accessToken, refreshToken, user, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Login successful",
		Data:       views.ToTokenResponse(accessToken, refreshToken, user),
	}
	return resp.JSON(c)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req views.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return errz.New(errz.BadRequest, "Invalid request body", err)
	}

	accessToken, refreshToken, err := h.Service.RefreshTokens(req.RefreshToken)
	if err != nil {
		return err
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Tokens refreshed successfully",
		Data:       views.ToTokenOnlyResponse(accessToken, refreshToken),
	}
	return resp.JSON(c)
}