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
	VIN        string                       `json:"vin"`
	RetailDate time.Time                    `json:"retail_date"`
	Periods    []warrantyYearPeriodResponse `json:"periods"`
}

type warrantyYearPeriodResponse struct {
	WarrantyPeriod warrantyPeriodResponse `json:"warranty_period"`
	TotalDays      int                    `json:"total_days"`
	Items          []warrantyPeriodItem   `json:"items"`
}

type warrantyPeriodResponse struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type warrantyPeriodItem struct {
	Claim      warrantyPeriodClaim `json:"claim"`
	RepairDays int                 `json:"repair_days"`
}

type warrantyPeriodClaim struct {
	ID          int64     `json:"id"`
	RoOpenDate  time.Time `json:"ro_open_date"`
	RoCloseDate time.Time `json:"ro_close_date"`
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

	repoResp, err := h.claimRepo.ListWarrantyYearRepairsByVIN(
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

	periods := make([]warrantyYearPeriodResponse, 0, len(repoResp.Periods))
	for _, period := range repoResp.Periods {
		items := make([]warrantyPeriodItem, 0, len(period.Items))
		for _, item := range period.Items {
			items = append(items, warrantyPeriodItem{
				Claim: warrantyPeriodClaim{
					ID:          item.Claim.ID,
					RoOpenDate:  item.Claim.RoOpenDate,
					RoCloseDate: item.Claim.RoCloseDate,
				},
				RepairDays: item.RepairDays,
			})
		}

		periods = append(periods, warrantyYearPeriodResponse{
			WarrantyPeriod: warrantyPeriodResponse{
				Start: period.WarrantyStart,
				End:   period.WarrantyEnd,
			},
			TotalDays: period.TotalDays,
			Items:     items,
		})
	}

	resp := warrantyYearResponse{
		VIN:        repoResp.VIN,
		RetailDate: repoResp.RetailDate,
		Periods:    periods,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}
