// Package jwts
package jwts

import (
	"errors"
	"fmt"
	"time"

	utility "blog_server/share/utils"

	"github.com/golang-jwt/jwt/v5"
)

// Token types
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type JWTUtil interface {
	GenerateRefreshToken(userID, email, role string) (string, time.Time, error)
	GenerateAccessToken(userID string) (string, time.Time, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
}

type jwtUtil struct {
	accessSecret  []byte        // For access tokens
	refreshSecret []byte        // For refresh tokens (DIFFERENT!)
	accessExpiry  time.Duration // e.g., 15 minutes
	refreshExpiry time.Duration // e.g., 7 days
}

type Claims struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func NewJWTUtils(JWTSecret string) JWTUtil {
	return &jwtUtil{
		accessSecret:  []byte(JWTSecret), // For access tokens
		refreshSecret: []byte(JWTSecret), // For refresh tokens (DIFFERENT!)
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 30 * 24 * time.Hour,
	}
}

// GenerateAccessToken implements JWTUtil.
func (ts *jwtUtil) GenerateRefreshToken(userID, name, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(ts.accessExpiry)

	// JWT-based access token
	claims := &Claims{
		ID:        userID,
		Name:      name,
		Role:      role,
		TokenType: TokenTypeAccess, // ← CORRECT TYPE
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
			ID:        utility.UniqueID(20), // Unique ID for access token
		},
	}

	// Sign with ACCESS secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ts.accessSecret)

	return tokenString, expiresAt, err
}

func (ts *jwtUtil) GenerateAccessToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(ts.refreshExpiry)

	// Method 1: JWT-based refresh token
	claims := &Claims{
		ID:        userID,
		TokenType: TokenTypeRefresh, // ← DIFFERENT TYPE
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
			ID:        utility.UniqueID(20), // Unique ID for refresh token
		},
	}

	// Sign with REFRESH secret (DIFFERENT from access!)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ts.refreshSecret)

	return tokenString, expiresAt, err
}

func (ts *jwtUtil) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Check token type
		if claims.TokenType != TokenTypeAccess {
			return nil, errors.New("invalid token type: expected access token")
		}

		// Use ACCESS secret for validation
		return ts.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (ts *jwtUtil) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Check token type
		if claims.TokenType != TokenTypeRefresh {
			return nil, errors.New("invalid token type: expected refresh token")
		}

		// Use REFRESH secret for validation (DIFFERENT!)
		return ts.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return claims, nil
}
