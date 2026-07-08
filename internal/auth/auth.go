package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
)

type AccountType string

const (
	AccountTypeAdvertiser AccountType = "advertiser"
	AccountTypePublisher  AccountType = "publisher"
	AccountTypeAdmin      AccountType = "admin"
)

type AccountStatus string

const (
	AccountStatusActive        AccountStatus = "active"
	AccountStatusSuspended     AccountStatus = "suspended"
	AccountStatusPendingReview AccountStatus = "pending_review"
)

type Claims struct {
	AccountID uuid.UUID   `json:"account_id"`
	Email     string      `json:"email"`
	Type      AccountType `json:"type"`
	jwt.RegisteredClaims
}

type JWTConfig struct {
	SecretKey     string
	AccessTokenTTL time.Duration
	RefreshTTL    time.Duration
}

type AuthService struct {
	config JWTConfig
}

func NewAuthService(config JWTConfig) *AuthService {
	return &AuthService{config: config}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateAccessToken(accountID uuid.UUID, email string, accountType AccountType) (string, error) {
	claims := Claims{
		AccountID: accountID,
		Email:     email,
		Type:      accountType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

func (s *AuthService) GenerateRefreshToken(accountID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   accountID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, ErrTokenExpired
		}
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	accountID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return accountID, nil
}

func (s *AuthService) HasPermission(accountType AccountType, requiredType AccountType) bool {
	if accountType == AccountTypeAdmin {
		return true
	}
	return accountType == requiredType
}
