package tests

import (
	"testing"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT_Success(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	role := "org_admin"
	orgID := "org-456"

	token, err := utils.GenerateJWT(secret, userID, role, orgID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseJWT_Success(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	role := "org_admin"
	orgID := "org-456"

	// Generate a token first
	token, err := utils.GenerateJWT(secret, userID, role, orgID)
	assert.NoError(t, err)

	// Parse the token
	claims, err := utils.ParseJWT(token, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, orgID, claims.OrgID)
}

func TestParseJWT_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	invalidToken := "invalid.token.here"

	claims, err := utils.ParseJWT(invalidToken, secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWT_WrongSecret(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret"
	userID := "user-123"
	role := "org_admin"
	orgID := "org-456"

	// Generate token with correct secret
	token, err := utils.GenerateJWT(secret, userID, role, orgID)
	assert.NoError(t, err)

	// Try to parse with wrong secret
	claims, err := utils.ParseJWT(token, wrongSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	role := "org_admin"
	orgID := "org-456"

	// Manually create an expired token
	claims := utils.JWTClaims{
		UserID: userID,
		Role:   role,
		OrgID:  orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	// Try to parse expired token
	parsedClaims, err := utils.ParseJWT(tokenString, secret)

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
}
