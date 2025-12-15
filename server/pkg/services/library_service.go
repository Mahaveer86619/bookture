package services

import (
	"errors"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
	"gorm.io/gorm"
)

type LibraryService struct {
	db *gorm.DB
}

func NewLibraryService() *LibraryService {
	return &LibraryService{
		db: db.GetBooktureDB().DB,
	}
}

func (s *LibraryService) CreateNewLibrary(userID uint, name string) (*views.LibraryView, error) {
	lib := models.Library{
		UserID: userID,
		Name:   name,
	}

	if err := s.db.Create(&lib).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to create library", err)
	}

	v := views.ToLibraryView(&lib)
	return &v, nil
}

func (s *LibraryService) GetUserLibraries(userID uint) ([]views.LibraryView, error) {
	var libraries []models.Library
	if err := s.db.Preload("Books").Where("user_id = ?", userID).Find(&libraries).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to fetch libraries", err)
	}

	return views.ToLibraryViews(libraries), nil
}

func (s *LibraryService) GetLibrary(id uint, userID uint) (*views.LibraryView, error) {
	var lib models.Library
	if err := s.db.Preload("Books").Where("id = ? AND user_id = ?", id, userID).First(&lib).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "Library not found or access denied", err)
		}
		return nil, errz.New(errz.InternalServerError, "Database error", err)
	}

	v := views.ToLibraryView(&lib)
	return &v, nil
}

func (s *LibraryService) UpdateLibrary(id uint, userID uint, name string) (*views.LibraryView, error) {
	var lib models.Library
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&lib).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "Library not found or access denied", err)
		}
		return nil, errz.New(errz.InternalServerError, "Database error", err)
	}

	lib.Name = name
	if err := s.db.Save(&lib).Error; err != nil {
		return nil, err
	}

	v := views.ToLibraryView(&lib)
	return &v, nil
}

func (s *LibraryService) DeleteLibrary(id uint, userID uint) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Library{}).Unscoped()
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return errz.New(errz.NotFound, "Library not found or access denied", res.Error)
		}
		return errz.New(errz.InternalServerError, "Database error", res.Error)
	}

	if res.RowsAffected == 0 {
		return errz.New(errz.NotFound, "Library not found or access denied", errors.New("no rows affected"))
	}

	return nil
}
