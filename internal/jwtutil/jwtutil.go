package jwtutil

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessCookieName  = "nl_access"
	RefreshCookieName = "nl_refresh"
	AccessTTL         = 15 * time.Minute
	RefreshTTL        = 30 * 24 * time.Hour
)

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

func secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-in-production"
	}
	return []byte(s)
}

func SignAccess(userID string) (string, error) {
	return sign(userID, AccessTTL)
}

func sign(userID string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret())
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
