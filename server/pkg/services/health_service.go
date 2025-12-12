package services

import "github.com/Mahaveer86619/bookture/server/pkg/views"

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (hs *HealthService) CheckHealth() (views.HealthView, error) {
	return views.NewHealthView("healthy", "Service is running smoothly"), nil
}
