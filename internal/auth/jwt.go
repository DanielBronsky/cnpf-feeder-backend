package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AuthCookieName = "cnpf_auth"

var jwtSecret []byte

// InitJWT initializes JWT secret from environment
func InitJWT() error {
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		return fmt.Errorf("missing AUTH_SECRET in environment variables")
	}
	jwtSecret = []byte(secret)
	return nil
}

// AuthClaims represents JWT claims
type AuthClaims struct {
	Sub   string `json:"sub"` // userId
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// SignToken creates a JWT token
func SignToken(userID, email string) (string, error) {
	if jwtSecret == nil {
		if err := InitJWT(); err != nil {
			return "", err
		}
	}

	claims := AuthClaims{
		Sub:   userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 days
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyToken verifies and parses JWT token
func VerifyToken(tokenString string) (*AuthClaims, error) {
	if jwtSecret == nil {
		if err := InitJWT(); err != nil {
			return nil, err
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
