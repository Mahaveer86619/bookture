package services

import (
	"errors"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
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
	// 1. Hash Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 2. Create User
	user := models.User{
		Email:        email,
		PasswordHash: string(hashed),
		DisplayName:  displayName,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// 3. Generate Token
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
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
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

func (s *UserService) Refresh(refreshToken string) (*views.AuthResponse, error) {
	userID, err := middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Verify user still exists
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Rotate tokens: Generate a fresh pair
	newAccessToken, newRefreshToken, err := middleware.GenerateTokens(user.ID)
	if err != nil {
		return nil, err
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
		return nil, errors.New("user not found")
	}
	v := views.ToUserView(&user)
	return &v, nil
}

func (s *UserService) UpdateUser(id uint, displayName string) (*views.UserView, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("user not found")
	}

	user.DisplayName = displayName
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	v := views.ToUserView(&user)
	return &v, nil
}

func (s *UserService) DeleteUser(id uint) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("user not found")
	}

	if err := s.db.Unscoped().Delete(&user).Error; err != nil {
		return err
	}

	return nil
}
