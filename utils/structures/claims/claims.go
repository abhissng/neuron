package claims

import (
	"time"

	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
)

// StandardClaims represents the standard claims in a token.
type StandardClaims struct {
	Iss  string         `json:"iss"`            // Issuer
	Exp  time.Time      `json:"exp"`            // Expiration Time
	Iat  time.Time      `json:"iat"`            // Issued At Time
	Jti  string         `json:"jti"`            // JWT ID (Unique identifier for the token)
	Aud  string         `json:"aud,omitempty"`  // Audience (Optional)
	Nbf  time.Time      `json:"nbf,omitempty"`  // Not Before (Optional)
	Sub  string         `json:"sub,omitempty"`  // Subject (Optional)
	Ip   string         `json:"ip,omitempty"`   // IP address (Optional)
	Pid  string         `json:"pid"`            // Payload ID created at payload time
	Data map[string]any `json:"data,omitempty"` // Data (Optional)
}

// StandardClaimsOption is a functional option for configuring StandardClaims.
type StandardClaimsOption func(*StandardClaims)

// WithAudience sets the Audience claim.
func WithAudience(aud string) StandardClaimsOption {
	return func(c *StandardClaims) {
		c.Aud = aud
	}
}

// WithNotBefore sets the NotBefore claim.
func WithNotBefore(nbf time.Time) StandardClaimsOption {
	return func(c *StandardClaims) {
		c.Nbf = nbf
	}
}

// WithSubject sets the Subject claim.
func WithSubject(sub string) StandardClaimsOption {
	return func(c *StandardClaims) {
		c.Sub = sub
	}
}

// WithIP sets the IP claim.
func WithIP(ip string) StandardClaimsOption {
	return func(c *StandardClaims) {
		c.Ip = ip
	}
}

// WithData sets the Data claim.
func WithData(data map[string]any) StandardClaimsOption {
	return func(c *StandardClaims) {
		if c.Data == nil {
			c.Data = make(map[string]any)
		}
		c.Data = data
	}
}

// WithPid sets the Payload ID.
func (c *StandardClaims) WithPid() *StandardClaims {
	// Comment: Sets the Payload ID for the claims.
	c.Pid = GetRandomPid(c.Sub, c.Iss, c.Jti)
	return c
}

// GetRandomPid returns a payload ID derived from subject, issuer, and JWT ID.
func GetRandomPid(subject, issuer, jti string) string {
	// Comment: Generates a unique Payload ID based on subject, issuer, and JWT ID.
	return random.JoinComponentsToID(subject, issuer, jti)
}

// NewStandardClaims creates a new StandardClaims struct with options.
func NewStandardClaims(issuer string, expiry time.Duration, options ...StandardClaimsOption) *StandardClaims {
	tokenID, _ := random.GenerateTokenID()
	if helpers.IsEmpty(tokenID) {
		tokenID, _ = random.GenerateRandomAlphanumeric(15)
	}
	// Comment: Creates a new StandardClaims struct with the provided issuer, expiry, and optional configurations.
	claims := &StandardClaims{
		Iss: issuer,
		Exp: time.Now().Add(expiry),
		Iat: time.Now(),
		Jti: tokenID,
	}

	for _, option := range options {
		option(claims)
	}

	return claims
}

// Issuer returns the token issuer (Iss claim).
func (c *StandardClaims) Issuer() string {
	// Comment: Returns the issuer of the token.
	return c.Iss
}

// Expiration returns the token expiration time (Exp claim).
func (c *StandardClaims) Expiration() time.Time {
	// Comment: Returns the expiration time of the token.
	return c.Exp
}

// IssuedAt returns the token issued-at time (Iat claim).
func (c *StandardClaims) IssuedAt() time.Time {
	// Comment: Returns the issued at time of the token.
	return c.Iat
}

// JWTID returns the unique token identifier (Jti claim).
func (c *StandardClaims) JWTID() string {
	// Comment: Returns the unique identifier of the token.
	return c.Jti
}

// Audience returns the intended audience (Aud claim), if set.
func (c *StandardClaims) Audience() string {
	// Comment: Returns the intended audience of the token (optional).
	return c.Aud
}

// NotBefore returns the not-before time (Nbf claim), if set.
func (c *StandardClaims) NotBefore() time.Time {
	// Comment: Returns the time before which the token should not be accepted (optional).
	return c.Nbf
}

// Subject returns the token subject (Sub claim), if set.
func (c *StandardClaims) Subject() string {
	// Comment: Returns the subject of the token (optional).
	return c.Sub
}

// IP returns the client IP claim (Ip), if set.
func (c *StandardClaims) IP() string {
	// Comment: Returns the IP address associated with the token (optional).
	return c.Ip
}

// GetData returns the custom data map (Data claim), if set.
func (c *StandardClaims) GetData() map[string]any {
	// Comment: Returns the data associated with the token (optional).
	return c.Data
}
