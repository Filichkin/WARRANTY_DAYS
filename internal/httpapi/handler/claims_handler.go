// Package handler
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"warranty_days/internal/repo"
)

type ClaimsHandler struct {
	claimRepo *repo.ClaimRepo
}

type warrantyYearResponse struct {
	Items     []repo.ClaimRepairDaysItem `json:"items"`
	TotalDays int                        `json:"total_days"`
}

func NewClaimsHandler(claimRepo *repo.ClaimRepo) *ClaimsHandler {
	return &ClaimsHandler{claimRepo: claimRepo}
}

func (h *ClaimsHandler) Health(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (h *ClaimsHandler) GetClaimsByVIN(w http.ResponseWriter, r *http.Request) {
	vin := strings.TrimSpace(r.URL.Query().Get("vin"))
	if vin == "" {
		http.Error(w, "vin query param is required, example: /claims?vin=XXX", http.StatusBadRequest)
		return
	}

	claims, err := h.claimRepo.ListByVINCaseInsensitive(r.Context(), vin)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(claims)
}

func (h *ClaimsHandler) GetWarrantyYearClaims(w http.ResponseWriter, r *http.Request) {
	vin := strings.TrimSpace(r.URL.Query().Get("vin"))
	if vin == "" {
		http.Error(w, "vin query param is required, example: /claims/warranty-year?vin=XXX", http.StatusBadRequest)
		return
	}

	items, totalDays, err := h.claimRepo.ListWarrantyYearRepairsByVIN(r.Context(), vin, time.Now())
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
}
