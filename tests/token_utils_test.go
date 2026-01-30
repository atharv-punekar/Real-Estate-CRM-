package tests

import (
	"testing"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInviteToken_Success(t *testing.T) {
	token, err := utils.GenerateInviteToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// Token should be 64 characters (32 bytes in hex)
	assert.Equal(t, 64, len(token))
}

func TestGenerateInviteToken_Unique(t *testing.T) {
	token1, err1 := utils.GenerateInviteToken()
	token2, err2 := utils.GenerateInviteToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	// Two tokens should be different
	assert.NotEqual(t, token1, token2)
}

func TestGetTokenExpiry(t *testing.T) {
	before := time.Now()
	expiry := utils.GetTokenExpiry()
	after := time.Now()

	// Expiry should be 7 days from now
	expectedMin := before.Add(7 * 24 * time.Hour)
	expectedMax := after.Add(7 * 24 * time.Hour)

	assert.True(t, expiry.After(expectedMin) || expiry.Equal(expectedMin))
	assert.True(t, expiry.Before(expectedMax) || expiry.Equal(expectedMax))
}

func TestIsTokenExpired_NotExpired(t *testing.T) {
	// Token expires in the future
	futureTime := time.Now().Add(1 * time.Hour)

	result := utils.IsTokenExpired(&futureTime)

	assert.False(t, result)
}

func TestIsTokenExpired_Expired(t *testing.T) {
	// Token expired in the past
	pastTime := time.Now().Add(-1 * time.Hour)

	result := utils.IsTokenExpired(&pastTime)

	assert.True(t, result)
}

func TestIsTokenExpired_Nil(t *testing.T) {
	// Nil expiry should be treated as expired
	result := utils.IsTokenExpired(nil)

	assert.True(t, result)
}

func TestIsTokenExpired_ExactlyNow(t *testing.T) {
	// Token expires right now (edge case)
	now := time.Now()

	result := utils.IsTokenExpired(&now)

	// Should be expired or very close to it
	assert.True(t, result)
}
