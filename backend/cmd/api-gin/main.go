package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"warranty_days/internal/auth"
	"warranty_days/internal/config"
	"warranty_days/internal/db"
	"warranty_days/internal/httpapi/handler"
	ginrouter "warranty_days/internal/httpapi_gin/router"
	"warranty_days/internal/logging"
	"warranty_days/internal/repo"
	"warranty_days/internal/service"
)

func main() {
	loadDotEnv(".env", "../.env", "../../.env")

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
	userRepo := repo.NewUserRepo(gormDB)

	jwtSvc := auth.NewJWTService(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	authSvc := service.NewAuthService(userRepo, jwtSvc)

	claimsHandler := handler.NewClaimsHandler(claimRepo, logger)
	authHandler := handler.NewAuthHandler(authSvc)

	ginEngine := ginrouter.NewEngine(claimsHandler, authHandler, jwtSvc, logger)

	logger.Info("gin server starting", "http_addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, ginEngine); err != nil {
		logger.Error("http server stopped", "error", err)
		os.Exit(1)
	}
}

func loadDotEnv(paths ...string) {
	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			return
		}
	}
}
