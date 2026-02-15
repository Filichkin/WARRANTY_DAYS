// Package handler
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"warranty_days/internal/repo"
)

type ClaimsHandler struct {
	claimRepo *repo.ClaimRepo
	logger    *slog.Logger
}

type warrantyYearResponse struct {
	Items          []repo.ClaimRepairDaysItem `json:"items"`
	TotalDays      int                        `json:"total_days"`
	RetailDate     time.Time                  `json:"retail_date"`
	WarrantyPeriod warrantyPeriodResponse     `json:"warranty_period"`
}

type warrantyPeriodResponse struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func NewClaimsHandler(claimRepo *repo.ClaimRepo, logger *slog.Logger) *ClaimsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ClaimsHandler{claimRepo: claimRepo, logger: logger}
}

func (h *ClaimsHandler) Health(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (h *ClaimsHandler) GetClaimsByVIN(w http.ResponseWriter, r *http.Request) {
	vin := strings.TrimSpace(r.URL.Query().Get("vin"))
	if vin == "" {
		h.logger.WarnContext(r.Context(), "vin query param is missing", "path", r.URL.Path)
		http.Error(w, "vin query param is required, example: /claims?vin=XXX", http.StatusBadRequest)
		return
	}

	claims, err := h.claimRepo.ListByVINCaseInsensitive(r.Context(), vin)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to fetch claims by vin", "vin", vin, "error", err)
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
		h.logger.WarnContext(r.Context(), "vin query param is missing", "path", r.URL.Path)
		http.Error(w, "vin query param is required, example: /claims/warranty-year?vin=XXX", http.StatusBadRequest)
		return
	}

	items, totalDays, retailDate, warrantyStart, warrantyEnd, err := h.claimRepo.ListWarrantyYearRepairsByVIN(
		r.Context(),
		vin,
		time.Now(),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.InfoContext(r.Context(), "claims not found for vin", "vin", vin)
			http.Error(w, "claims not found for vin", http.StatusNotFound)
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to fetch warranty-year claims", "vin", vin, "error", err)
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := warrantyYearResponse{
		Items:      items,
		TotalDays:  totalDays,
		RetailDate: retailDate,
		WarrantyPeriod: warrantyPeriodResponse{
			Start: warrantyStart,
			End:   warrantyEnd,
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}
