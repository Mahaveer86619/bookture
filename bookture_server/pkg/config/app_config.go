package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port int

	DatabaseHost    string
	DatabasePort    int
	DatabaseUser    string
	DatabasePass    string
	DatabaseName    string
	DatabaseSSLMode string

	JWTSecret string
}

var AppConfig Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file: %v", err)
	}

	AppConfig = Config{
		Port:            GetEnvAsInt("APP_PORT", 8080),
		DatabaseHost:    GetEnv("DB_HOST", "localhost"),
		DatabasePort:    GetEnvAsInt("DB_PORT", 5432),
		DatabaseUser:    GetEnv("DB_USER", "bookture_user"),
		DatabasePass:    GetEnv("DB_PASS", "password"),
		DatabaseName:    GetEnv("DB_NAME", "bookture_db"),
		DatabaseSSLMode: GetEnv("DB_SSLMODE", "disable"),
		JWTSecret:       GetEnv("JWT_SECRET", "default_jwt_secret"),
	}
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(name string, defaultVal int) int {
	if valueStr, exists := os.LookupEnv(name); exists {
		var value int
		_, err := fmt.Sscanf(valueStr, "%d", &value)
		if err == nil {
			return value
		}
	}
	return defaultVal
}
