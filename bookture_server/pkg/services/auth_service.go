package services

import (
	"errors"

	"github.com/Mahaveer86619/bookture/pkg/db"
	"github.com/Mahaveer86619/bookture/pkg/errz"
	"github.com/Mahaveer86619/bookture/pkg/middleware"
	"github.com/Mahaveer86619/bookture/pkg/models"
	"github.com/Mahaveer86619/bookture/pkg/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{
		db: db.GetDB(),
	}
}

func (s *AuthService) Register(username, email, password string) (*models.User, error) {
	// 1. Check if exists
	var count int64
	s.db.Model(&models.User{}).Where("email = ? OR username = ?", email, username).Count(&count)
	if count > 0 {
		return nil, errz.New(errz.Conflict, "User with this email or username already exists", nil)
	}

	// 2. Hash Password (Using Utils)
	hashed, err := utils.HashPassword(password)
	if err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to hash password", err)
	}

	// 3. Create
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashed,
	}

	if err := db.GetDB().Create(user).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Database error", err)
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, string, *models.User, error) {
	var user models.User
	// Preload counts if needed, for now just basic user
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil, errz.New(errz.Unauthorized, "Invalid credentials", err)
		}
		return "", "", nil, err
	}

	// Validate (Using Utils)
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", nil, errz.New(errz.Unauthorized, "Invalid credentials", nil)
	}

	// Generate JWT
	access, refresh, err := middleware.GenerateTokens(user.ID)
	if err != nil {
		return "", "", nil, errz.New(errz.InternalServerError, "Token generation failed", err)
	}

	return access, refresh, &user, nil
}

func (s *AuthService) RefreshTokens(refreshToken string) (string, string, error) {
	// Validate Refresh Token
	userID, err := middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return "", "", errz.New(errz.Unauthorized, "User no longer exists", err)
	}

	// Generate New Tokens
	access, refresh, err := middleware.GenerateTokens(userID)
	if err != nil {
		return "", "", errz.New(errz.InternalServerError, "Token generation failed", err)
	}

	return access, refresh, nil
}
