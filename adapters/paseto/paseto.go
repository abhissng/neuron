package paseto

import (
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/claims"
)

// **Paseto Wrapper Type**
type PasetoWrapper struct {
	privateKey             ed25519.PrivateKey // For auth service (token generation)
	publicKey              ed25519.PublicKey  // For other services (token validation)
	issuer                 string
	accessTokenExpiry      time.Duration
	refreshTokenExpiry     time.Duration
	pasetoMiddlewareOption *PasetoMiddlewareOptions
}

// **Token Generation**

// FetchToken generates a new access token
func (p *PasetoWrapper) FetchToken(options ...claims.StandardClaimsOption) result.Result[TokenDetails] {
	return p.createToken(p.issuer, p.accessTokenExpiry, options...)
}

// FetchRefreshToken generates a new refresh token
func (p *PasetoWrapper) FetchRefreshToken(options ...claims.StandardClaimsOption) result.Result[TokenDetails] {
	return p.createToken(p.issuer, p.refreshTokenExpiry, options...)
}

// createToken generates a new token with the given issuer, expiry, and options
func (p *PasetoWrapper) createToken(issuer string, expiry time.Duration, options ...claims.StandardClaimsOption) result.Result[TokenDetails] {

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

// ValidateToken validates a token
func (p *PasetoWrapper) ValidateToken(token string, validatePayload func(payload *claims.StandardClaims) error) result.Result[claims.StandardClaims] {
	var claim claims.StandardClaims
	// Decrypt the token
	err := GetPasetoObj().Verify(token, p.publicKey, &claim, nil)
	if err != nil {
		return result.NewFailure[claims.StandardClaims](blame.MalformedAuthToken(err))
	}

	// Validate standard claims
	if claim.Iss != p.issuer {
		return result.NewFailure[claims.StandardClaims](blame.UntrustedTokenIssuer())
	}

	if helpers.IsEmpty(claim.Exp) {
		return result.NewFailure[claims.StandardClaims](blame.MalformedAuthToken(nil))
	}

	if time.Now().After(claim.Exp) {
		return result.NewFailure[claims.StandardClaims](blame.ExpiredAuthToken())
	}

	// Validate custom payload
	if validatePayload != nil {
		if err := validatePayload(&claim); err != nil {
			return result.NewFailure[claims.StandardClaims](blame.AuthValidationFailed(err))
		}
	}

	return result.NewSuccess(&claim)
}

// PasetoMiddlewareOption returns the middleware options for the PASETO wrapper.
func (p *PasetoWrapper) PasetoMiddlewareOption() *PasetoMiddlewareOptions {
	return p.pasetoMiddlewareOption
}

// ValidateEssentialTags validates the essential tags in the standard claims
func ValidateEssentialTags(claim *claims.StandardClaims) error {
	if claim.Pid == "" {
		return errors.New("payload id is missing")
	}

	if claim.Pid != claims.GetRandomPid(claim.Sub, claim.Iss, claim.Jti) {
		return errors.New("payload id does not match")
	}

	return nil
}
