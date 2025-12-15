package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT string

	PROFILE string

	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string

	ID_SALT string

	JWT_SECRET string

	STORAGE_DRIVER string
	STORAGE_PATH   string
}

var AppConfig Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Error loading .env file: %v", err)
	}

	AppConfig = Config{
		PORT: getEnv("PORT", "7000"),

		PROFILE: getEnv("PROFILE", "dev"),

		DB_HOST:     getEnv("DB_HOST", "localhost"),
		DB_PORT:     getEnv("DB_PORT", "5432"),
		DB_USER:     getEnv("DB_USER", "postgres"),
		DB_PASSWORD: getEnv("DB_PASSWORD", "password"),
		DB_NAME:     getEnv("DB_NAME", "bookture"),

		ID_SALT: getEnv("ID_SALT", "bookture-secret-salt-change-me"),

		JWT_SECRET: getEnv("JWT_SECRET", "your_secret_key"),

		STORAGE_DRIVER: getEnv("STORAGE_DRIVER", "local"),
		STORAGE_PATH:   getEnv("STORAGE_PATH", "./uploads"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := lookupEnv(key); exists {
		return value
	}
	return defaultValue
}

var lookupEnv = func(key string) (string, bool) {
	return os.LookupEnv(key)
}
