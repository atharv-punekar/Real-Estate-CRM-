package tests

import (
	"testing"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword_Success(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := utils.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // Hash should be different from password
}

func TestCheckPassword_Success(t *testing.T) {
	password := "mySecurePassword123"

	// Hash the password
	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)

	// Verify correct password
	result := utils.CheckPassword(hash, password)
	assert.True(t, result)
}

func TestCheckPassword_Failure(t *testing.T) {
	password := "mySecurePassword123"
	wrongPassword := "wrongPassword456"

	// Hash the correct password
	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)

	// Try to verify with wrong password
	result := utils.CheckPassword(hash, wrongPassword)
	assert.False(t, result)
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "mySecurePassword123"
	emptyPassword := ""

	// Hash the correct password
	hash, err := utils.HashPassword(password)
	assert.NoError(t, err)

	// Try to verify with empty password
	result := utils.CheckPassword(hash, emptyPassword)
	assert.False(t, result)
}

func TestGenerateRandomPassword(t *testing.T) {
	length := 16

	password := utils.GenerateRandomPassword(length)

	assert.NotEmpty(t, password)
	assert.Equal(t, length, len(password))
}

func TestGenerateRandomPassword_DifferentOutputs(t *testing.T) {
	length := 16

	password1 := utils.GenerateRandomPassword(length)
	password2 := utils.GenerateRandomPassword(length)

	// Two random passwords should be different
	assert.NotEqual(t, password1, password2)
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "mySecurePassword123"

	// Hash the same password twice
	hash1, err1 := utils.HashPassword(password)
	hash2, err2 := utils.HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	// Bcrypt generates different salts, so hashes should be different
	assert.NotEqual(t, hash1, hash2)
	// But both should verify the same password
	assert.True(t, utils.CheckPassword(hash1, password))
	assert.True(t, utils.CheckPassword(hash2, password))
}
