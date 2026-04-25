package config

import (
	"os"
	"strconv"
)

type Config struct {
	Address     string
	DatabaseURL string
	JWTSecret   string
	TokenHours  int
	AdminCode   string
}

func Load() Config {
	tokenHours, err := strconv.Atoi(getEnv("TOKEN_HOURS", "24"))
	if err != nil || tokenHours <= 0 {
		tokenHours = 24
	}

	return Config{
		Address:     getEnv("ADDRESS", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/meeting_room?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "change_this_secret"),
		TokenHours:  tokenHours,
		AdminCode:   os.Getenv("ADMIN_CODE"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
