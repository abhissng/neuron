package cryptography

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/blake2s"
)

// DefaultHashSizeBytes defines the default hash output size (in bytes).
// You can override this by setting the environment variable HASH_SIZE (1–32).
var DefaultHashSizeBytes = getDefaultHashSize()

// getDefaultHashSize reads HASH_SIZE from environment, or falls back to 8 bytes.
func getDefaultHashSize() int {
	return 16 // default to 16 bytes (~128-bit hash)
}

func GenerateHash(key, data []byte, size int) (string, error) {
	if len(key) < 16 {
		return "", fmt.Errorf("key must be at least 16 bytes")
	}

	if len(data) == 0 {
		return "", fmt.Errorf("data must not be empty")
	}

	switch size {
	case 16:
		return Generate128BitHash(key, data)
	case 32:
		return Generate256BitHash(key, data)
	default:
		return "", fmt.Errorf("invalid hash size: must be 16 or 32 bytes")
	}
}

// Generate128BitHash creates a small, secure keyed hash using BLAKE2s.
//
// Parameters:
//   - key: secret key (recommended at least 16 bytes)
//   - data: input to hash (e.g., OTP, message, etc.)
//
// Returns:
//   - base64-url-safe string (no padding)
//
// Example:
//
//	hash, err := Generate128BitHash([]byte("key123"), []byte("otp-123456"))
//	fmt.Println("Hash:", hash)
func Generate128BitHash(key, data []byte) (string, error) {
	if len(key) < 16 {
		return "", fmt.Errorf("key must be at least 16 bytes")
	}

	hasher, err := blake2s.New128(key)
	if err != nil {
		return "", fmt.Errorf("failed to initialize blake2s: %w", err)
	}

	hasher.Write(data)
	sum := hasher.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(sum), nil
}

// Generate256BitHash creates a small, secure keyed hash using BLAKE2s.
//
// Parameters:
//   - key: secret key (recommended at least 16 bytes)
//   - data: input to hash (e.g., OTP, message, etc.)
//
// Returns:
//   - base64-url-safe string (no padding)
//
// Example:
//
//	hash, err := Generate256BitHash([]byte("key123"), []byte("otp-123456"))
//	fmt.Println("Hash:", hash)
func Generate256BitHash(key, data []byte) (string, error) {
	if len(key) < 16 {
		return "", fmt.Errorf("key must be at least 16 bytes")
	}

	hasher, err := blake2s.New256(key)
	if err != nil {
		return "", fmt.Errorf("failed to initialize blake2s: %w", err)
	}

	hasher.Write(data)
	sum := hasher.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(sum), nil
}

// CompareHash safely compares two hash strings in constant time.
//
// Returns true if they match, false otherwise.
//
// Example:
//
//	ok := CompareHash(storedHash, receivedHash)
//	if ok { ... }
func CompareHash(hash1, hash2 string) bool {
	return subtle.ConstantTimeCompare([]byte(hash1), []byte(hash2)) == 1
}
