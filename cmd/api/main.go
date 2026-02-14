package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"warranty_days/internal/config"
	"warranty_days/internal/db"
	"warranty_days/internal/repo"
)

type warrantyYearResponse struct {
	Items     []repo.ClaimRepairDaysItem `json:"items"`
	TotalDays int                        `json:"total_days"`
}

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

	mux.HandleFunc("/claims/warranty-year", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vin := strings.TrimSpace(r.URL.Query().Get("vin"))
		if vin == "" {
			http.Error(w, "vin query param is required, example: /claims/warranty-year?vin=XXX", http.StatusBadRequest)
			return
		}

		items, totalDays, err := claimRepo.ListWarrantyYearRepairsByVIN(r.Context(), vin, time.Now())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "claims not found for vin", http.StatusNotFound)
				return
			}
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := warrantyYearResponse{
			Items:     items,
			TotalDays: totalDays,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
	})

	log.Println("server on", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
