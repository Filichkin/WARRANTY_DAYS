package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"warranty_days/internal/models"
)

type ClaimRepo struct {
	db *gorm.DB
}

type ClaimRepairDaysItem struct {
	Claim      models.Claim `json:"claim"`
	RepairDays int          `json:"repair_days"`
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

func (r *ClaimRepo) ListWarrantyYearRepairsByVIN(ctx context.Context, vin string, now time.Time) ([]ClaimRepairDaysItem, int, error) {
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
			return nil, 0, err
		}
		return nil, 0, err
	}

	warrantyStart, warrantyEnd := currentWarrantyYearWindow(retailHolder.RetailDate, now)

	var claims []models.Claim
	err = r.db.WithContext(ctx).
		Where("LOWER(vin) = LOWER(?)", vin).
		Where("ro_open_date <= ? AND ro_close_date >= ?", warrantyEnd, warrantyStart).
		Order("ro_open_date ASC, id ASC").
		Find(&claims).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]ClaimRepairDaysItem, 0, len(claims))
	totalDays := 0

	for _, claim := range claims {
		effectiveOpen := maxDate(claim.RoOpenDate, warrantyStart)
		effectiveClose := minDate(claim.RoCloseDate, warrantyEnd)
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

	return items, totalDays, nil
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
