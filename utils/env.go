package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// GetEnvAsInt retrieves an integer environment variable or returns a default value.
func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s, using default: %d\n", key, defaultValue)
		return defaultValue
	}
	return value
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Error loading %s", key)
		return defaultValue
	}
	return value
}
