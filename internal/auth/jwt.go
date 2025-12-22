package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

func getAccessSecret() []byte {
	secret := os.Getenv("ACCESS_TOKEN_SECRET")
	if secret == "" {
		panic("ACCESS_TOKEN_SECRET not set")
	}
	return []byte(secret)
}

func getRefreshSecret() []byte {
	secret := os.Getenv("REFRESH_TOKEN_SECRET")
	if secret == "" {
		panic("REFRESH_TOKEN_SECRET not set")
	}
	return []byte(secret)
}

// Claims defines JWT payload
type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

/*
========================
 TOKEN GENERATION
========================
*/

func GenerateAccessToken(userID string) (string, error) {
	return generateToken(userID, getAccessSecret(), accessTokenTTL)
}

func GenerateRefreshToken(userID string) (string, error) {
	return generateToken(userID, getRefreshSecret(), refreshTokenTTL)
}


func generateToken(userID string, secret []byte, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

/*
========================
 TOKEN VALIDATION
========================
*/

// ValidateAccessToken validates access token
func ValidateAccessToken(tokenStr string) (*Claims, error) {
	return validateToken(tokenStr, getAccessSecret())
}

// ValidateRefreshToken validates refresh token
func ValidateRefreshToken(tokenStr string) (*Claims, error) {
	// TODO (future):
	// 1. Check Redis blacklist
	// 2. Reject if token revoked

	return validateToken(tokenStr, getRefreshSecret())
}

func validateToken(tokenStr string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
