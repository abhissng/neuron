package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/abhissng/neuron/utils/helpers"
	"github.com/dgrijalva/jwt-go"
)

// Claims represents the JWT claims
type Claims struct {
	ServiceName string   `json:"service"`
	Roles       []string `json:"roles"`
	jwt.StandardClaims
}

func NewJWTClaims(serviceName string, roles []string, expiryDuration time.Duration) *Claims {
	expirationTime := time.Now().Add(expiryDuration).Unix()
	claims := Claims{
		ServiceName: serviceName,
		Roles:       roles,
		StandardClaims: jwt.StandardClaims{
			Issuer:    serviceName,
			ExpiresAt: expirationTime,
			IssuedAt:  time.Now().Unix(),
		},
	}
	return &claims
}

// GenerateJWT generates a JWT token with the given secret and expiration time
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

// ValidateJWT validates a JWT token with the given secret and expected claims
func ValidateJWT(tokenString string, secret string, validRoles []string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
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
	// if claims.Issuer != expectedIssuer {
	// 	return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	// }
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}
	if claims.IssuedAt > time.Now().Unix() {
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
