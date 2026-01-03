package services

import (
	"github.com/Mahaveer86619/bookture/pkg/db"
	"gorm.io/gorm"
)

type HealthService struct {
	db *gorm.DB
}

func NewHealthService() *HealthService {
	return &HealthService{
		db: db.GetDB(),
	}
}

func (s *HealthService) Check() (bool, error) {
	sqlDB, err := s.db.DB()
	if err != nil {
		return false, err
	}

	if err := sqlDB.Ping(); err != nil {
		return false, err
	}

	return true, nil
}
