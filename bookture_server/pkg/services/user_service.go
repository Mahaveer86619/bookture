package services

import (
	"errors"

	"github.com/Mahaveer86619/bookture/pkg/db"
	"github.com/Mahaveer86619/bookture/pkg/errz"
	"github.com/Mahaveer86619/bookture/pkg/models"
	"github.com/Mahaveer86619/bookture/pkg/views"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{
		db: db.GetDB(),
	}
}

func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.
		Preload("Followers").
		Preload("Following").
		First(&user, "id = ?", userID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "User not found", err)
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, req views.UpdateProfileRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, errz.New(errz.NotFound, "User not found", err)
	}

	// Update fields if provided
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to update profile", err)
	}

	return &user, nil
}

func (s *UserService) FollowUser(followerID, targetID uuid.UUID) error {
	if followerID == targetID {
		return errz.New(errz.BadRequest, "You cannot follow yourself", nil)
	}

	err := s.db.Model(&models.User{Base: models.Base{ID: followerID}}).
		Association("Following").
		Append(&models.User{Base: models.Base{ID: targetID}})

	return err
}

func (s *UserService) UnfollowUser(followerID, targetID uuid.UUID) error {
	err := s.db.Model(&models.User{Base: models.Base{ID: followerID}}).
		Association("Following").
		Delete(&models.User{Base: models.Base{ID: targetID}})

	return err
}

func (s *UserService) GetFollowers(userID uuid.UUID) ([]models.User, error) {
	var user models.User
	err := s.db.Preload("Followers").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return user.Followers, nil
}

func (s *UserService) GetFollowing(userID uuid.UUID) ([]models.User, error) {
	var user models.User
	err := s.db.Preload("Following").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return user.Following, nil
}
