package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error: %v", err)
		log.Println("No .env file found, relying on system environment variables")
	}
	port := getEnv("PORT", "7000")

	return Config{
		PORT: port,
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
