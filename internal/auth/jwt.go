package auth

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	UserEmailContextKey = "user_email"
)

type Config struct {
	JWTSecret         string
	DebugLoginEnabled bool
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingEmail = errors.New("token email is required")
)

func (c Config) JWTEnabled() bool {
	return strings.TrimSpace(c.JWTSecret) != ""
}

func GenerateToken(cfg Config, email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrMissingEmail
	}
	if !cfg.JWTEnabled() {
		return "", ErrInvalidToken
	}
	claims := Claims{
		Email:            email,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func ValidateToken(cfg Config, tokenString string) (*Claims, error) {
	if !cfg.JWTEnabled() {
		return nil, ErrInvalidToken
	}
	token, err := jwt.ParseWithClaims(strings.TrimSpace(tokenString), &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims.Email = strings.TrimSpace(claims.Email)
	if claims.Email == "" {
		return nil, ErrMissingEmail
	}
	return claims, nil
}
