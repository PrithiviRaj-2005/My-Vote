package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims defines the payload saved inside JWT tokens
type JWTClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// getSecret retrieves the secret key used for signing JWT tokens
func getSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pulsevote_default_super_secret_jwt_key_2026"
	}
	return []byte(secret)
}

// GenerateJWT creates a new signed JWT token for an authenticated user
func GenerateJWT(userID string, email string, name string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // Valid for 7 days

	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT verifies token signature and parses embedded claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure signing algorithm is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}
