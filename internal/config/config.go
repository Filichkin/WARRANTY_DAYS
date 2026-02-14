// Package config
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv        string
	LogLevel      string
	HTTPAddr      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTSecret     string
	JWTIssuer     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

func Load() (Config, error) {
	accessTTL, err := parseDurationEnv("JWT_ACCESS_TTL", "15m")
	if err != nil {
		return Config{}, err
	}

	refreshTTL, err := parseDurationEnv("JWT_REFRESH_TTL", "168h") // 7 days
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:        os.Getenv("APP_ENV"),
		LogLevel:      os.Getenv("LOG_LEVEL"),
		HTTPAddr:      os.Getenv("HTTP_ADDR"),
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		DBSSLMode:     os.Getenv("DB_SSLMODE"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTIssuer:     os.Getenv("JWT_ISSUER"),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
	}
	// дефолты
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "127.0.0.1"
	}
	if cfg.DBPort == "" {
		cfg.DBPort = "5432"
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "warranty_days"
	}

	// обязательные поля
	if cfg.DBUser == "" {
		return Config{}, errors.New("DB_USER is required")
	}
	if cfg.DBPassword == "" {
		return Config{}, errors.New("DBPassword is required")
	}
	if cfg.DBName == "" {
		return Config{}, errors.New("DB_NAME is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if cfg.JWTAccessTTL <= 0 {
		return Config{}, errors.New("JWT_ACCESS_TTL must be > 0")
	}
	if cfg.JWTRefreshTTL <= 0 {
		return Config{}, errors.New("JWT_REFRESH_TTL must be > 0")
	}

	return cfg, nil
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func parseDurationEnv(key, fallback string) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		raw = fallback
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s has invalid duration %q: %w", key, raw, err)
	}
	return d, nil
}
