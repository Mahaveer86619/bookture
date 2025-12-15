package services

import (
	"errors"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/middleware"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{
		db: db.GetBooktureDB().DB,
	}
}

func (s *UserService) Register(email, password, displayName string) (*views.AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hashed),
		DisplayName:  displayName,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errz.New(errz.Conflict, "Email already registered", err)
	}

	accessToken, refreshToken, err := middleware.GenerateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return &views.AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         views.ToUserView(&user),
	}, nil
}

func (s *UserService) Login(email, password string) (*views.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errz.New(errz.Unauthorized, "Invalid credentials", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errz.New(errz.Unauthorized, "Invalid credentials", err)
	}

	accessToken, refreshToken, err := middleware.GenerateTokens(user.ID)
	if err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to generate tokens", err)
	}

	return &views.AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         views.ToUserView(&user),
	}, nil
}

func (s *UserService) Refresh(refreshToken string) (*views.AuthResponse, error) {
	userID, err := middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errz.New(errz.Unauthorized, "Invalid or expired refresh token", err)
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errz.New(errz.Unauthorized, "User account not found", err)
	}

	newAccessToken, newRefreshToken, err := middleware.GenerateTokens(user.ID)
	if err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to rotate tokens", err)
	}

	return &views.AuthResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User:         views.ToUserView(&user),
	}, nil
}

func (s *UserService) GetUser(id uint) (*views.UserView, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "User not found", err)
		}
		return nil, errz.New(errz.InternalServerError, "Database error", err)
	}

	v := views.ToUserView(&user)
	return &v, nil
}

func (s *UserService) UpdateUser(id uint, displayName string) (*views.UserView, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "User not found", err)
		}
		return nil, errz.New(errz.InternalServerError, "Database error", err)
	}

	user.DisplayName = displayName
	if err := s.db.Save(&user).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to update user", err)
	}
	v := views.ToUserView(&user)
	return &v, nil
}

func (s *UserService) DeleteUser(id uint) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errz.New(errz.NotFound, "User not found", err)
		}
		return errz.New(errz.InternalServerError, "Database error", err)
	}

	if err := s.db.Unscoped().Delete(&user).Error; err != nil {
		return errz.New(errz.InternalServerError, "Failed to delete user", err)
	}

	return nil
}
