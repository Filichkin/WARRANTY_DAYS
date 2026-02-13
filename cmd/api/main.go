package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"

	"warranty_days/internal/config"
	"warranty_days/internal/db"
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

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/claims", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vin := strings.TrimSpace(r.URL.Query().Get("vin"))
		if vin == "" {
			http.Error(w, "vin query param is required, example: /claims?vin=XXX", http.StatusBadRequest)
			return
		}

		claims, err := claimRepo.ListByVINCaseInsensitive(r.Context(), vin)
		if err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(claims)
	})

	log.Println("server on", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}