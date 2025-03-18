package paseto

import (
	"time"

	"github.com/o1egl/paseto"
)

// GetPasetoObj creates a new Paseto V2 instance.
func GetPasetoObj() *paseto.V2 {
	// Comment: Returns a new instance of the Paseto V2 library for token operations.
	return paseto.NewV2()
}

// PasetoMiddlewareOptions defines options for the Paseto middleware.
type PasetoMiddlewareOptions struct {
	isAutoRefresh    bool          // Indicates whether auto-refresh is enabled.
	authHeader       string        // The name of the authorization header (e.g., "Authorization").
	newAuthHeader    string        // The name of the header for the new token (e.g., "New-Authorization").
	refreshThreshold time.Duration // Time before token expiration to trigger refresh.
	excludedServices []string      // List of services to exclude from token validation.
}

// Getters for PasetoMiddlewareOptions fields.
func (p *PasetoMiddlewareOptions) IsAutoRefresh() bool {
	// Comment: Returns whether auto-refresh is enabled.
	return p.isAutoRefresh
}

// AuthHeader returns the name of the authorization header.
func (p *PasetoMiddlewareOptions) AuthHeader() string {
	// Comment: Returns the name of the authorization header.
	return p.authHeader
}

// NewAuthHeader returns the name of the header for the new token (if auto-refresh is enabled).
func (p *PasetoMiddlewareOptions) NewAuthHeader() string {
	// Comment: Returns the name of the header for the new token (if auto-refresh is enabled).
	return p.newAuthHeader
}

// RefreshThreshold returns the time threshold before token expiration to trigger refresh.
func (p *PasetoMiddlewareOptions) RefreshThreshold() time.Duration {
	// Comment: Returns the time threshold before token expiration to trigger refresh.
	return p.refreshThreshold
}

// ExcludedServices returns the list of services to exclude from token validation.
func (p *PasetoMiddlewareOptions) ExcludedServices() []string {
	// Comment: Returns the list of services to exclude from token validation.
	return p.excludedServices
}

// HasExcludedService returns true if the list of excluded services is not empty.
func (p *PasetoMiddlewareOptions) HasExcludedService() bool {
	return len(p.excludedServices) > 0
}

// TokenDetails holds information about a token.
type TokenDetails struct {
	Token     string
	ExpiresAt time.Time
	ID        string // Unique identifier for the token
}
