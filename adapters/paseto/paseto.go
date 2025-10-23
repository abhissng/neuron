package paseto

import (
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/claims"
	"github.com/abhissng/neuron/utils/types"
)

// **Paseto Wrapper Type**
type PasetoManager struct {
	privateKey             ed25519.PrivateKey // For auth service (token generation)
	publicKey              ed25519.PublicKey  // For other services (token validation)
	issuer                 string
	basicTokenExpiry       time.Duration
	accessTokenExpiry      time.Duration
	refreshTokenExpiry     time.Duration
	pasetoMiddlewareOption *PasetoMiddlewareOptions
}

// **Token Generation**

// FetchToken generates a new access token
func (p *PasetoManager) FetchToken(options ...claims.StandardClaimsOption) result.Result[TokenDetails] {
	return p.createToken(p.issuer, p.accessTokenExpiry, options...)
}

// FetchRefreshToken generates a new refresh token
func (p *PasetoManager) FetchRefreshToken(options ...claims.StandardClaimsOption) result.Result[TokenDetails] {
	return p.createToken(p.issuer, p.refreshTokenExpiry, options...)
}

// FetchBasicToken generates a new basic token
func (p *PasetoManager) FetchBasicToken(options ...claims.StandardClaimsOption) result.Result[TokenDetails] {
	return p.createToken(p.issuer, p.basicTokenExpiry, options...)
}

// createToken generates a new token with the given issuer, expiry, and options
func (p *PasetoManager) createToken(issuer string, expiry time.Duration, options ...claims.StandardClaimsOption) result.Result[TokenDetails] {

	// Create standard claims
	standardClaims := claims.NewStandardClaims(issuer, expiry, options...).WithPid()

	// Encrypt the token
	token, err := GetPasetoObj().Sign(p.privateKey, standardClaims, nil)
	if err != nil {
		return result.NewFailure[TokenDetails](blame.CreateTokenFailed())
	}

	tokenDetails := TokenDetails{
		Token:     token,
		ExpiresAt: standardClaims.Exp,
		ID:        standardClaims.Jti,
	}

	return result.NewSuccess(&tokenDetails)
}

// TokenValidator defines a function that validates claims.
type TokenValidator func(claim *claims.StandardClaims, extra map[string]any) error

// ValidateToken validates a token and runs multiple custom validators.
func (p *PasetoManager) ValidateToken(
	token string,
	extra map[string]any,
	validators ...TokenValidator,
) result.Result[claims.StandardClaims] {
	var claim claims.StandardClaims

	// Decrypt and verify token
	err := GetPasetoObj().Verify(token, p.publicKey, &claim, nil)
	if err != nil {
		return result.NewFailure[claims.StandardClaims](blame.MalformedAuthToken(err))
	}

	// Validate standard fields
	if claim.Iss != p.issuer {
		return result.NewFailure[claims.StandardClaims](blame.UntrustedTokenIssuer())
	}
	if helpers.IsEmpty(claim.Exp) {
		return result.NewFailure[claims.StandardClaims](blame.MalformedAuthToken(nil))
	}
	if time.Now().After(claim.Exp) {
		return result.NewFailure[claims.StandardClaims](blame.ExpiredAuthToken())
	}

	// Run custom validators
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator(&claim, extra); err != nil {
			return result.NewFailure[claims.StandardClaims](blame.AuthValidationFailed(err))
		}
	}

	return result.NewSuccess(&claim)
}

// PasetoMiddlewareOption returns the middleware options for the PASETO wrapper.
func (p *PasetoManager) PasetoMiddlewareOption() *PasetoMiddlewareOptions {
	return p.pasetoMiddlewareOption
}

// WithValidateEssentialTags ensures core payload fields are correct.
func WithValidateEssentialTags(claim *claims.StandardClaims, extra map[string]any) error {
	if claim.Pid == "" {
		return errors.New("payload id is missing")
	}

	if claim.Pid != claims.GetRandomPid(claim.Sub, claim.Iss, claim.Jti) {
		return errors.New("payload id does not match")
	}

	if !helpers.IsEmpty(claim.Ip) {
		ip, ok := types.CastTo[string](extra["ip"])
		if ok && claim.Ip != ip {
			return errors.New("ip does not match")
		}
	}

	if !helpers.IsEmpty(claim.Sub) {
		subject, ok := types.CastTo[string](extra["subject"])
		if ok && claim.Sub != subject {
			return errors.New("subject does not match")
		}
	}

	return nil
}
