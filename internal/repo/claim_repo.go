package repo

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"warranty_days/internal/models"
)

type ClaimRepo struct {
	db *gorm.DB
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