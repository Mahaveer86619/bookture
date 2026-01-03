package handlers

import (
	"net/http"

	"github.com/Mahaveer86619/bookture/pkg/services"
	"github.com/Mahaveer86619/bookture/pkg/views"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	Service *services.HealthService
}

func NewHealthHandler(service *services.HealthService) *HealthHandler {
	return &HealthHandler{
		Service: service,
	}
}

func (h *HealthHandler) Check(c echo.Context) error {
	dbOk, err := h.Service.Check()

	status := "running"
	dbStatus := "connected"

	if err != nil || !dbOk {
		dbStatus = "disconnected"
	}

	resp := &views.Success{
		StatusCode: http.StatusOK,
		Message:    "Server is healthy",
		Data: views.HealthStatus{
			Status:   status,
			Database: dbStatus,
		},
	}
	return resp.JSON(c)
}
