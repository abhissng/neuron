package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/abhissng/neuron/utils/helpers"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the custom JWT claims structure with service information.
// It embeds jwt.RegisteredClaims and adds service-specific fields.
type Claims struct {
	ServiceName string   `json:"service"`
	Roles       []string `json:"roles"`
	jwt.RegisteredClaims
}

// NewJWTClaims creates a new JWT claims structure with the specified parameters.
// It sets up standard claims like issuer, expiration, and issued-at times.
func NewJWTClaims(serviceName string, roles []string, expiryDuration time.Duration) *Claims {
	now := time.Now()
	claims := Claims{
		ServiceName: serviceName,
		Roles:       roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    serviceName,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return &claims
}

// GenerateJWT creates and signs a JWT token with HMAC SHA256.
// It includes service name, roles, and standard claims with the specified expiration.
func GenerateJWT(serviceName string, roles []string, secret string, expiryDuration time.Duration) (string, error) {

	claims := NewJWTClaims(serviceName, roles, expiryDuration)

	// Create token with HMAC SHA256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT parses and validates a JWT token against the provided secret and roles.
// It performs signature verification, expiration checks, and role validation.
func ValidateJWT(tokenString string, secret string, validRoles []string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	// Ensure the token is valid
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	// Extract the claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("failed to extract claims")
	}

	// Validate standard claims
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token has expired")
	}
	if claims.IssuedAt != nil && claims.IssuedAt.After(time.Now()) {
		return nil, errors.New("token issued-at time is in the future")
	}

	// Validate custom claims
	if len(claims.Roles) == 0 {
		return nil, errors.New("roles claim is missing or empty")
	}
	for _, role := range claims.Roles {
		if !helpers.IsFoundInSlice(role, validRoles) {
			return nil, fmt.Errorf("invalid role: %s", role)
		}
	}

	return claims, nil
}
