package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"go-boilerplate/internal/config"
	"go-boilerplate/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateToken(t *testing.T) {
	// Create temporary keys for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// careful with temp files in tests running concurrent
	privFile, err := os.CreateTemp("", "private_*.pem")
	require.NoError(t, err)
	defer os.Remove(privFile.Name())
	_, err = privFile.Write(privateKeyPEM)
	require.NoError(t, err)
	privFile.Close()

	pubFile, err := os.CreateTemp("", "public_*.pem")
	require.NoError(t, err)
	defer os.Remove(pubFile.Name())
	_, err = pubFile.Write(publicKeyPEM)
	require.NoError(t, err)
	pubFile.Close()

	cfg := config.JWTConfig{
		PrivateKeyPath:   privFile.Name(),
		PublicKeyPath:    pubFile.Name(),
		AccessExpiresIn:  15,
		RefreshExpiresIn: 10080,
	}

	userID := uint(123)
	// Test Token Pair
	accessToken, refreshToken, err := auth.GenerateTokenPair(userID, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// Validate Access Token
	claims, err := auth.ValidateToken(accessToken, cfg)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, auth.TokenTypeAccess, claims.TokenType)

	// Validate Refresh Token
	refreshClaims, err := auth.ValidateRefreshToken(refreshToken, cfg)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, auth.TokenTypeRefresh, refreshClaims.TokenType)

	// Test cross-validation failure (Refresh token masquerading as Access token)
	_, err = auth.ValidateToken(refreshToken, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")

	// Test invalid token string
	_, err = auth.ValidateToken(accessToken+"invalid", cfg)
	assert.Error(t, err)
}
