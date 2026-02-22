package random

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// GenerateUUID generates a UUID
func GenerateUUIDString() string {
	// Implement your UUID generation logic here
	return uuid.New().String()
}

func GenerateUUID() uuid.UUID {
	// Implement your UUID generation logic here
	return uuid.New()
}

// JoinComponentsToID joins multiple strings into a single ID
func JoinComponentsToID(components ...string) string {
	return strings.Join(components, "-")
}

// GenerateRandomNumber generates a random number with up to `n` digits (max 20)
func GenerateRandomNumber(n int) (string, error) {
	if n <= 0 || n > 20 {
		n = 20 // Default max length
	}

	// Generate a number in the range 10^(n-1) to (10^n)-1
	min := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n-1)), nil)
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	max.Sub(max, big.NewInt(1))

	num, err := rand.Int(rand.Reader, new(big.Int).Sub(max, min))
	if err != nil {
		return "", err
	}
	num.Add(num, min) // Shift to min range

	return num.String(), nil
}

// GenerateRandomAlphanumeric generates a random alphanumeric string of `n` length
func GenerateRandomAlphanumeric(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if n <= 0 {
		n = 10 // Default length
	}

	var sb strings.Builder
	sb.Grow(n)

	for i := 0; i < n; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(charset[index.Int64()])
	}

	return sb.String(), nil
}

// GenerateTokenID generates a random token ID
func GenerateTokenID() (string, error) {
	bytes := make([]byte, 16)  // 128 bits (UUID size)
	_, err := rand.Read(bytes) // Use crypto/rand
	if err != nil {
		return "", fmt.Errorf("error generating token ID: %w", err) // Wrap error
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomStringNumber generates a random string number of `n` length
func GenerateRandomStringNumber(n int) string {
	str, err := GenerateRandomNumber(n)
	if err != nil {
		return ""
	}
	return str
}
