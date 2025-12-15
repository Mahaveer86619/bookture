package services

import (
	"github.com/Mahaveer86619/bookture/server/pkg/services/storage"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
)

type HealthService struct{
	ss storage.StorageService
}

func NewHealthService(storage storage.StorageService) *HealthService {
	return &HealthService{
		ss: storage,
	}
}

func (hs *HealthService) CheckHealth() (views.HealthView, error) {
	status := "healthy"
	msg := "Service is running smoothly"

	if err := hs.ss.HealthCheck(); err != nil {
		status = "degraded"
		msg = "Storage Issue: " + err.Error()
	}

	return views.NewHealthView(status, msg), nil
}
