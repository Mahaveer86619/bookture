package handlers

import (
	"errors"
	"net/http"

	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/services"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
)

type HealthHandler struct {
	healthService *services.HealthService
}

func NewHealthHandler(healthService *services.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

func (hh *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	healthView, err := hh.healthService.CheckHealth()
	if err != nil {
		errz.HandleErrors(w, http.StatusInternalServerError, errors.New("service is unavailable"))
		return
	}

	successResp := views.Success{}
	successResp.SetStatusCode(http.StatusOK)
	successResp.SetMessage("Health check successful")
	successResp.SetData(healthView)

	if err := successResp.JSON(w); err != nil {
		errz.HandleErrors(w, http.StatusInternalServerError, err)
		return
	}
}
