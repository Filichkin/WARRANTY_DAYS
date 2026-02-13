package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:      os.Getenv("APP_ENV"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
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