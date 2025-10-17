package paseto

import (
	"crypto/ed25519"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

const (
	RefreshThreshold = 5 * time.Minute
)

// **Option Type for Functional Pattern**
type PasetoOption func(*PasetoManager)

// **Constructor Using Option Pattern**
func NewPasetoManager(opts ...PasetoOption) *PasetoManager {
	pw := &PasetoManager{
		basicTokenExpiry: time.Minute * 5,
	}

	for _, opt := range opts {
		opt(pw)
	}

	return pw
}

// **Functional Options**

// WithKeys sets the private and public keys for the PASETO wrapper.
func WithKeys(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) PasetoOption {
	return func(p *PasetoManager) {
		p.privateKey = privateKey // Auth service only
		p.publicKey = publicKey   // All other services
	}
}

// WithPrivateKey sets the private key for the PASETO wrapper.
func WithPrivateKey(privateKey ed25519.PrivateKey) PasetoOption {
	return func(p *PasetoManager) {
		p.privateKey = privateKey // Auth service only
	}
}

// WithPublicKey sets the public key for the PASETO wrapper.
func WithPublicKey(publicKey ed25519.PublicKey) PasetoOption {
	return func(p *PasetoManager) {
		p.publicKey = publicKey // All other services
	}
}

// WithIssuer sets the issuer for the PASETO wrapper.
func WithIssuer(issuer string) PasetoOption {
	return func(p *PasetoManager) {
		p.issuer = issuer
	}
}

// WithExpiry sets the access and refresh token expirations for the PASETO wrapper.
func WithExpiry(accessToken, refreshToken time.Duration) PasetoOption {
	return func(p *PasetoManager) {
		p.accessTokenExpiry = accessToken
		p.refreshTokenExpiry = refreshToken
	}
}

// WithAccessTokenExpiry sets the access token expiry for the PASETO wrapper.
func WithAccessTokenExpiry(accessToken time.Duration) PasetoOption {
	return func(p *PasetoManager) {
		p.accessTokenExpiry = accessToken
	}
}

// WithRefreshTokenExpiry sets the refresh token expiry for the PASETO wrapper.
func WithRefreshTokenExpiry(refreshToken time.Duration) PasetoOption {
	return func(p *PasetoManager) {
		p.refreshTokenExpiry = refreshToken
	}
}

// WithBasicTokenExpiry sets the basic token expiry for the PASETO wrapper.
func WithBasicTokenExpiry(basicToken time.Duration) PasetoOption {
	return func(p *PasetoManager) {
		p.basicTokenExpiry = basicToken
	}
}

// WithPasetoMiddlewareOption sets the middleware options for the PASETO wrapper.
func WithPasetoMiddlewareOption(opts ...PasetoMiddlewareOption) PasetoOption {
	return func(p *PasetoManager) {
		// Set default middleware options
		defaultOptions := &PasetoMiddlewareOptions{
			isAutoRefresh:    false,
			authHeader:       constant.AuthorizationHeader,
			newAuthHeader:    constant.XRefreshToken,
			refreshThreshold: RefreshThreshold,
		}

		// Apply user-provided options
		for _, opt := range opts {
			opt(defaultOptions)
		}

		// Assign final options to the wrapper
		p.pasetoMiddlewareOption = defaultOptions
	}
}

// PasetoMiddlewareOption sets options for the Paseto middleware.
type PasetoMiddlewareOption func(*PasetoMiddlewareOptions)

// WithAutoRefresh enables token auto-refresh
func WithAutoRefresh(enabled bool) PasetoMiddlewareOption {
	return func(o *PasetoMiddlewareOptions) {
		o.isAutoRefresh = enabled
	}
}

// WithAuthHeader sets the authorization header name
func WithAuthHeader(header string) PasetoMiddlewareOption {
	if helpers.IsEmpty(header) {
		header = constant.RefreshToken
	}
	return func(o *PasetoMiddlewareOptions) {
		o.authHeader = header
	}
}

// WithNewAuthHeader sets the new token response header name
func WithNewAuthHeader(header string) PasetoMiddlewareOption {
	if helpers.IsEmpty(header) {
		header = constant.XRefreshToken
	}
	return func(o *PasetoMiddlewareOptions) {
		o.newAuthHeader = header
	}
}

// WithRefreshThreshold sets the time before expiration to trigger refresh
func WithRefreshThreshold(duration time.Duration) PasetoMiddlewareOption {
	return func(o *PasetoMiddlewareOptions) {
		o.refreshThreshold = duration
	}
}

/*
// WithExcludedServices sets the excluded services for the PASETO wrapper.
func WithExcludedServices(services []string) PasetoMiddlewareOption {
	return func(o *PasetoMiddlewareOptions) {
		o.excludedServices = services
	}
}
*/

// WithExcludedOptions sets the excluded options for the PASETO wrapper.
func WithExcludedOptions(options *ExcludedOptions) PasetoMiddlewareOption {
	return func(o *PasetoMiddlewareOptions) {
		o.excludedOptions = options
	}
}
