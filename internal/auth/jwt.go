// Package auth for token operations
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Claims struct {
	UserID    int64  `json:"uid"`
	Email     string `json:"email"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	nowFn      func() time.Time
}

func NewJWTService(secret, issuer string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		nowFn:      time.Now,
	}
}

func (s *JWTService) GenerateAccessToken(userID int64, email string) (string, error) {
	return s.generateToken(userID, email, TokenTypeAccess, s.accessTTL)
}

func (s *JWTService) GenerateRefreshToken(userID int64, email string) (string, error) {
	return s.generateToken(userID, email, TokenTypeRefresh, s.refreshTTL)
}

func (s *JWTService) ParseAndValidate(tokenStr string, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != s.issuer {
		return nil, ErrInvalidToken
	}

	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) generateToken(userID int64, email, tokenType string, ttl time.Duration) (string, error) {
	now := s.nowFn()

	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			Audience:  []string{"warranty_days_api"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}
	return signed, nil
}
