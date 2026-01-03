package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/pkg/config"
	"github.com/Mahaveer86619/bookture/pkg/errz"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var jwtKey = []byte(config.AppConfig.JWTSecret)

type contextKey string

const userIDKey contextKey = "user_id"

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID uuid.UUID) (string, string, error) {
	accessExpiration := time.Now().Add(15 * time.Minute)
	accessClaims := &Claims{
		UserID:    userID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiration),
			Issuer:    "bookture-server",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	refreshExpiration := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &Claims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
			Issuer:    "bookture-server",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, errz.New(errz.Unauthorized, "Invalid refresh token", err)
	}

	if claims.TokenType != "refresh" {
		return uuid.Nil, errz.New(errz.Unauthorized, "Invalid token type", nil)
	}

	return claims.UserID, nil
}

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return errz.New(errz.Unauthorized, "Authorization header required", nil)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			return errz.New(errz.Unauthorized, "Invalid or expired token", err)
		}

		ctx := context.WithValue(c.Request().Context(), userIDKey, claims.UserID)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func GetUserID(r *http.Request) uuid.UUID {
	val := r.Context().Value(userIDKey)
	if val == nil {
		return uuid.Nil
	}

	if id, ok := val.(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
