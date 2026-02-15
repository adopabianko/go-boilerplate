package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"go-boilerplate/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID    string    `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func parsePrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	return jwt.ParseRSAPrivateKeyFromPEM(keyData)
}

func parsePublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}
	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}

func GenerateTokenPair(userID string, cfg config.JWTConfig) (accessToken, refreshToken string, err error) {
	key, err := parsePrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return "", "", err
	}

	// Access Token
	accessClaims := &Claims{
		UserID:    userID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.AccessExpiresIn) * time.Minute)),
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(key)
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := &Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.RefreshExpiresIn) * time.Minute)),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(key)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Deprecated: Use GenerateTokenPair instead. Keeping for backward compatibility if needed, but updated to use AccessExpiresIn
func GenerateToken(userID string, cfg config.JWTConfig) (string, error) {
	accessToken, _, err := GenerateTokenPair(userID, cfg)
	return accessToken, err
}

func ValidateToken(tokenString string, cfg config.JWTConfig) (*Claims, error) {
	return validateTokenWithType(tokenString, cfg, TokenTypeAccess)
}

func ValidateRefreshToken(tokenString string, cfg config.JWTConfig) (*Claims, error) {
	return validateTokenWithType(tokenString, cfg, TokenTypeRefresh)
}

func validateTokenWithType(tokenString string, cfg config.JWTConfig, expectedType TokenType) (*Claims, error) {
	key, err := parsePublicKey(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.TokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
