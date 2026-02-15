// Package repo provides database repositories for claims and related entities
package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"warranty_days/internal/models"

	"gorm.io/gorm"
)

type ClaimRepo struct {
	db *gorm.DB
}

type ClaimRepairDaysItem struct {
	Claim      models.Claim `json:"claim"`
	RepairDays int          `json:"repair_days"`
}

type WarrantyYearPeriod struct {
	WarrantyStart time.Time            `json:"warranty_start"`
	WarrantyEnd   time.Time            `json:"warranty_end"`
	TotalDays     int                  `json:"total_days"`
	Items         []ClaimRepairDaysItem `json:"items"`
}

type WarrantyYearsResponse struct {
	VIN        string               `json:"vin"`
	RetailDate time.Time            `json:"retail_date"`
	Periods    []WarrantyYearPeriod `json:"periods"`
}

func NewClaimRepo(db *gorm.DB) *ClaimRepo {
	return &ClaimRepo{db: db}
}

func (r *ClaimRepo) ListByVINCaseInsensitive(ctx context.Context, vin string) ([]models.Claim, error) {
	vin = strings.TrimSpace(vin)

	var claims []models.Claim
	err := r.db.WithContext(ctx).
		Where("LOWER(vin) = LOWER(?)", vin).
		Order("id DESC").
		Find(&claims).Error

	return claims, err
}

func (r *ClaimRepo) ListWarrantyYearRepairsByVIN(
	ctx context.Context,
	vin string,
	now time.Time,
) (WarrantyYearsResponse, error) {
	vin = strings.TrimSpace(vin)
	if now.IsZero() {
		now = time.Now()
	}

	var retailHolder struct {
		RetailDate time.Time
	}

	err := r.db.WithContext(ctx).
		Model(&models.Claim{}).
		Select("retail_date").
		Where("LOWER(vin) = LOWER(?)", vin).
		Order("retail_date ASC").
		Take(&retailHolder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return WarrantyYearsResponse{}, err
		}
		return WarrantyYearsResponse{}, err
	}

	periods := buildWarrantyPeriods(retailHolder.RetailDate, now)
	if len(periods) == 0 {
		return WarrantyYearsResponse{
			VIN:        vin,
			RetailDate: retailHolder.RetailDate,
			Periods:    []WarrantyYearPeriod{},
		}, nil
	}

	oldestStart := periods[len(periods)-1].WarrantyStart
	newestEnd := periods[0].WarrantyEnd

	var claims []models.Claim
	err = r.db.WithContext(ctx).
		Where("LOWER(vin) = LOWER(?)", vin).
		Where("ro_open_date <= ? AND ro_close_date >= ?", newestEnd, oldestStart).
		Order("ro_open_date ASC, id ASC").
		Find(&claims).Error
	if err != nil {
		return WarrantyYearsResponse{}, err
	}

	for periodIdx := range periods {
		items := make([]ClaimRepairDaysItem, 0)
		totalDays := 0

		for _, claim := range claims {
			effectiveOpen := maxDate(claim.RoOpenDate, periods[periodIdx].WarrantyStart)
			effectiveClose := minDate(claim.RoCloseDate, periods[periodIdx].WarrantyEnd)
			if effectiveClose.Before(effectiveOpen) {
				continue
			}

			repairDays := daysInclusive(effectiveOpen, effectiveClose)
			items = append(items, ClaimRepairDaysItem{
				Claim:      claim,
				RepairDays: repairDays,
			})
			totalDays += repairDays
		}

		periods[periodIdx].Items = items
		periods[periodIdx].TotalDays = totalDays
	}

	return WarrantyYearsResponse{
		VIN:        vin,
		RetailDate: retailHolder.RetailDate,
		Periods:    periods,
	}, nil
}

func currentWarrantyYearWindow(retailDate time.Time, now time.Time) (time.Time, time.Time) {
	retailDate = toUTCDate(retailDate)
	now = toUTCDate(now)

	years := now.Year() - retailDate.Year()
	start := toUTCDate(retailDate.AddDate(years, 0, 0))
	if start.After(now) {
		start = toUTCDate(start.AddDate(-1, 0, 0))
	}

	end := toUTCDate(start.AddDate(1, 0, 0).AddDate(0, 0, -1))
	return start, end
}

func buildWarrantyPeriods(retailDate time.Time, now time.Time) []WarrantyYearPeriod {
	retailDate = toUTCDate(retailDate)
	now = toUTCDate(now)

	currentStart, currentEnd := currentWarrantyYearWindow(retailDate, now)
	periods := make([]WarrantyYearPeriod, 0)

	for start, end := currentStart, currentEnd; !start.Before(retailDate); {
		periods = append(periods, WarrantyYearPeriod{
			WarrantyStart: start,
			WarrantyEnd:   end,
			TotalDays:     0,
			Items:         []ClaimRepairDaysItem{},
		})

		start = toUTCDate(start.AddDate(-1, 0, 0))
		end = toUTCDate(end.AddDate(-1, 0, 0))
	}

	return periods
}

func daysInclusive(start time.Time, end time.Time) int {
	start = toUTCDate(start)
	end = toUTCDate(end)
	return int(end.Sub(start).Hours()/24) + 1
}

func toUTCDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func maxDate(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minDate(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
