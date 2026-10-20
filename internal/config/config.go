package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	AppPort         string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiresHours int
	AdminEmail      string
	AdminPassword   string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	jwtExpiresHours, err := envInt("JWT_EXPIRES_HOURS", 24)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		AppEnv:          envString("APP_ENV", "development"),
		AppPort:         envString("APP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTExpiresHours: jwtExpiresHours,
		AdminEmail:      os.Getenv("ADMIN_EMAIL"),
		AdminPassword:   os.Getenv("ADMIN_PASSWORD"),
	}

	required := map[string]string{
		"DATABASE_URL":   config.DatabaseURL,
		"JWT_SECRET":     config.JWTSecret,
		"ADMIN_EMAIL":    config.AdminEmail,
		"ADMIN_PASSWORD": config.AdminPassword,
	}

	var missing []string

	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf(
			"missing required environment variables: %s",
			strings.Join(missing, ", "),
		)
	}

	return config, nil
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback
	}

	return value
}

func envInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", key)
	}

	return parsed, nil
}
