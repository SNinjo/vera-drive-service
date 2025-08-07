package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Domain             string
	Port               string
	DatabaseURL        string
	IdentityServiceURL string
	ALLOWED_ORIGINS    []string
}

func NewConfig(logger *zap.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Warn("No .env file found or error loading .env file", zap.Error(err))
	}
	var domain string
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		domain = "0.0.0.0"
	} else {
		domain = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Domain:             domain,
		Port:               port,
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		IdentityServiceURL: os.Getenv("IDENTITY_SERVICE_URL"),
		ALLOWED_ORIGINS:    strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	}
}
