// Package service for operations with auth data
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"warranty_days/internal/auth"
	"warranty_days/internal/models"
	"warranty_days/internal/repo"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenService interface {
	GenerateAccessToken(userID int64, email string) (string, error)
	GenerateRefreshToken(userID int64, email string) (string, error)
	ParseAndValidate(tokenStr string, expectedType string) (*auth.Claims, error)
}

type AuthService struct {
	userRepo      *repo.UserRepo
	tokenService  TokenService
	passwordCost  int
	minPasswordLn int
}

func NewAuthService(userRepo *repo.UserRepo, tokenService TokenService) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		tokenService:  tokenService,
		passwordCost:  bcrypt.DefaultCost,
		minPasswordLn: 8,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	email = normalizeEmail(email)

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(password) < s.minPasswordLn {
		return nil, ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.passwordCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		// Если пользователь уже есть.
		// На сервисном слое нормализуем DB-ошибку в бизнес-ошибку.
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	email = normalizeEmail(email)

	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	access, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.tokenService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(email string) error {
	if email == "" || len(email) > 254 {
		return ErrInvalidEmail
	}

	// ParseAddress принимает display-name форматы, поэтому дополнительно проверяем
	// что адрес без лишнего оформления.
	addr, err := mail.ParseAddress(email)
	if err != nil || addr == nil {
		return ErrInvalidEmail
	}
	if strings.ToLower(strings.TrimSpace(addr.Address)) != email {
		return ErrInvalidEmail
	}

	return nil
}

// Пока простой вариант по тексту ошибки.
// На этапе production лучше сделать проверку через pgconn.PgError.Code == "23505".
func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "idx_users_email_lower_unique")
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidCredentials
	}

	claims, err := s.tokenService.ParseAndValidate(refreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	access, err := s.tokenService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.tokenService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
