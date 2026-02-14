package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"warranty_days/internal/config"
	"warranty_days/internal/db"
	"warranty_days/internal/httpapi/handler"
	"warranty_days/internal/httpapi/router"
	"warranty_days/internal/repo"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config error: ", err)
	}

	gormDB, err := db.NewGorm(cfg.DatabaseURL())
	if err != nil {
		log.Fatal("gorm connect error: ", err)
	}

	claimRepo := repo.NewClaimRepo(gormDB)
	claimsHandler := handler.NewClaimsHandler(claimRepo)
	mux := router.NewMux(claimsHandler)

	log.Println("server on", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
