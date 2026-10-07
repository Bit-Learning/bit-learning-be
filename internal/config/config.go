package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Environment  string
	HTTPAddr     string
	DatabaseURL  string
	JWTSecret    string
	JWTTTL       time.Duration
	AutoMigrate  bool
	OTLPEndpoint string
	ServiceName  string
}

func Load() (Config, error) {
	ttl, err := time.ParseDuration(getenv("JWT_TTL", "15m"))
	if err != nil {
		return Config{}, errors.New("invalid JWT_TTL: " + err.Error())
	}
	cfg := Config{
		Environment:  getenv("APP_ENV", "development"),
		HTTPAddr:     getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTTTL:       ttl,
		AutoMigrate:  getenv("AUTO_MIGRATE", "true") == "true",
		OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		ServiceName:  getenv("OTEL_SERVICE_NAME", "bit-learning-api"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
