package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateInviteToken generates a secure random token for invite links
func GenerateInviteToken() (string, error) {
	bytes := make([]byte, 32) // 64 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetTokenExpiry returns expiry time for invite tokens (7 days from now)
func GetTokenExpiry() time.Time {
	return time.Now().Add(7 * 24 * time.Hour)
}

// IsTokenExpired checks if a token has expired
func IsTokenExpired(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return true
	}
	return time.Now().After(*expiresAt)
}
