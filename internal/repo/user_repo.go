package repo

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"warranty_days/internal/models"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	user.Email = normalizeEmail(user.Email)

	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	email = normalizeEmail(email)

	var user models.User
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", email).
		Take(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
