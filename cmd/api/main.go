package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"warranty_days/internal/config"
	"warranty_days/internal/db"
	"warranty_days/internal/httpapi/handler"
	"warranty_days/internal/httpapi/router"
	"warranty_days/internal/logging"
	"warranty_days/internal/repo"
)

func main() {
	_ = godotenv.Load()

	bootstrapLogger := logging.New(os.Getenv("APP_ENV"), os.Getenv("LOG_LEVEL"), os.Stdout)
	slog.SetDefault(bootstrapLogger)

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error("config error", "error", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.AppEnv, cfg.LogLevel, os.Stdout)
	slog.SetDefault(logger)

	gormDB, err := db.NewGorm(cfg.DatabaseURL())
	if err != nil {
		logger.Error("gorm connect error", "error", err)
		os.Exit(1)
	}

	claimRepo := repo.NewClaimRepo(gormDB)
	claimsHandler := handler.NewClaimsHandler(claimRepo, logger)
	mux := router.NewMux(claimsHandler, logger)

	logger.Info("server starting", "http_addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil {
		logger.Error("http server stopped", "error", err)
		os.Exit(1)
	}
}
