package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(config.AppConfig.JWT_SECRET)

type contextKey string

const userIDKey contextKey = "user_id"

type Claims struct {
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID uint) (string, string, error) {
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

func ValidateRefreshToken(tokenString string) (uint, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return 0, errz.New(errz.Unauthorized, "Invalid refresh token", err)
	}

	if claims.TokenType != "refresh" {
		return 0, errz.New(errz.Unauthorized, "Invalid token type", nil)
	}

	return claims.UserID, nil
}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errz.HandleErrors(w, errz.New(errz.Unauthorized, "Authorization header required", nil))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			errz.HandleErrors(w, errz.New(errz.Unauthorized, "Invalid or expired token", err))
			return
		}

		if claims.TokenType != "access" {
			errz.HandleErrors(w, errz.New(errz.Unauthorized, "Invalid token type for authentication", nil))
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

func GetUserID(r *http.Request) uint {
	val := r.Context().Value(userIDKey)
	if val == nil {
		return 0
	}

	if id, ok := val.(uint); ok {
		return id
	}
	return 0
}
