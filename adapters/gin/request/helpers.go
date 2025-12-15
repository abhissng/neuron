package request

import (
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/google/uuid"
)

type EssentialHeaders struct {
	OrgId        types.OrgID  `json:"org_id"`
	UserId       types.UserID `json:"user_id"`
	LocationId   uuid.UUID    `json:"location_id"`
	UserRole     string       `json:"user_role"`
	FeatureFlags string       `json:"feature_flags"`
}

// NewEssentialHeaders creates and returns an empty EssentialHeaders value.
func NewEssentialHeaders() *EssentialHeaders {
	return &EssentialHeaders{}
}

// config for behaviour
type essentialHeadersConfig struct {
	RequireFeatureFlags bool
	RequireLocationID   bool
}

// Option type
type EssentialHeadersOption func(*essentialHeadersConfig)

// require X-Feature-Flags header
func WithFeatureFlagRequired() EssentialHeadersOption {
	return func(c *essentialHeadersConfig) {
		c.RequireFeatureFlags = true
	}
}

// require X-Location-Id header
func WithLocationIdRequired() EssentialHeadersOption {
	return func(c *essentialHeadersConfig) {
		c.RequireLocationID = true
	}
}

// GetEssentialHeadersValues extracts essential header values from the request using the provided options.
func GetEssentialHeadersValues(ctx *context.ServiceContext, options ...EssentialHeadersOption) (*EssentialHeaders, blame.Blame) {
	defer func() { helpers.RecoverException(recover()) }()
	cfg := &essentialHeadersConfig{
		RequireFeatureFlags: false,
		RequireLocationID:   false,
	}

	for _, o := range options {
		o(cfg)
	}

	orgIdResult := FetchXOrgIdHeader(ctx.Context)
	if !orgIdResult.IsSuccess() {
		ctx.SlogError("unable to get the org id", log.Blame(orgIdResult.Blame()))
		return nil, orgIdResult.Blame()
	}

	userIdResult := FetchXUserIdHeader(ctx.Context)
	if !userIdResult.IsSuccess() {
		ctx.SlogError("unable to get the user id", log.Blame(userIdResult.Blame()))
		return nil, userIdResult.Blame()
	}

	var userRole string
	userRoleResult := FetchXUserRoleHeader(ctx.Context)
	if !userRoleResult.IsSuccess() {
		ctx.SlogError("unable to get the user role", log.Blame(userRoleResult.Blame()))
		return nil, userRoleResult.Blame()
	}

	userRole = *userRoleResult.ToValue()

	// -------- FEATURE FLAGS ----------
	featureFlag := ""

	featureFlagsResult := FetchXFeatureFlagsHeader(ctx.Context)
	if !featureFlagsResult.IsSuccess() {
		if cfg.RequireFeatureFlags {
			ctx.SlogError("missing feature flags header (required)", log.Blame(featureFlagsResult.Blame()))
			return nil, featureFlagsResult.Blame()
		}
	} else {
		featureFlag = *featureFlagsResult.ToValue()
	}

	// -------- LOCATION ID ----------
	var locationId uuid.UUID
	locationResult := FetchXLocationIdHeader(ctx.Context)

	if !locationResult.IsSuccess() {
		if cfg.RequireLocationID {
			ctx.SlogError("missing location id header (required)", log.Blame(locationResult.Blame()))
			return nil, locationResult.Blame()
		}
	} else {
		locationId = *locationResult.ToValue()
	}

	orgId, ok := types.CastTo[types.OrgID](*orgIdResult.ToValue())
	if !ok {
		return nil, blame.TypeConversionError("org_id", "uuid.UUID", "OrgID", nil)
	}

	userId, ok := types.CastTo[types.UserID](*userIdResult.ToValue())
	if !ok {
		return nil, blame.TypeConversionError("user_id", "uuid.UUID", "UserID", nil)
	}

	return &EssentialHeaders{
		OrgId:        orgId,
		UserId:       userId,
		UserRole:     userRole,
		FeatureFlags: featureFlag,
		LocationId:   locationId,
	}, nil
}

type RequestAuthValues struct {
	Token         string
	CorrelationID types.CorrelationID
	XSubject      string
}

type RequestAuthConfig struct {
	RequireToken         bool
	RequireCorrelationID bool
	XSubject             bool
}

type RequestAuthOption func(*RequestAuthConfig)

// WithRequireToken marks the auth token as required when fetching request auth values.
func WithRequireToken() RequestAuthOption {
	return func(cfg *RequestAuthConfig) {
		cfg.RequireToken = true
	}
}

// WithRequireCorrelationID marks the correlation ID header as required when fetching request auth values.
func WithRequireCorrelationID() RequestAuthOption {
	return func(cfg *RequestAuthConfig) {
		cfg.RequireCorrelationID = true
	}
}

// WithRequireXSubject marks the X-Subject header as required when fetching request auth values.
func WithRequireXSubject() RequestAuthOption {
	return func(cfg *RequestAuthConfig) {
		cfg.XSubject = true
	}
}

// GetRequestAuthValues returns the request auth values
func GetRequestAuthValues(ctx *context.ServiceContext, options ...RequestAuthOption) (*RequestAuthValues, blame.Blame) {
	defer func() { helpers.RecoverException(recover()) }()

	// Default configuration - both required by default
	cfg := &RequestAuthConfig{
		RequireToken:         true,
		RequireCorrelationID: true,
	}

	// Apply options
	for _, opt := range options {
		opt(cfg)
	}

	// -------- TOKEN ----------
	var token string
	tokenResult := FetchPasetoBearerToken(ctx.Context)
	if !tokenResult.IsSuccess() {
		if cfg.RequireToken {
			ctx.SlogError("unable to get the authentication token", log.Blame(tokenResult.Blame()))
			return nil, tokenResult.Blame()
		}
	} else {
		token = *tokenResult.ToValue()
	}

	// -------- CORRELATION ID ----------
	var correlationID types.CorrelationID
	correlationIdResult := FetchCorrelationIdFromHeaders(ctx.Context)
	if !correlationIdResult.IsSuccess() {
		if cfg.RequireCorrelationID {
			ctx.SlogError("unable to get the correlation ID", log.Blame(correlationIdResult.Blame()))
			return nil, correlationIdResult.Blame()
		}
	} else {
		correlationID = *correlationIdResult.ToValue()
	}

	var xSubject string
	xSubjectResult := FetchXSubjectHeader(ctx.Context)
	if !xSubjectResult.IsSuccess() {
		if cfg.XSubject {
			ctx.SlogError("unable to get the X-Subject header", log.Blame(xSubjectResult.Blame()))
			return nil, xSubjectResult.Blame()
		}
	} else {
		xSubject = *xSubjectResult.ToValue()
	}

	return &RequestAuthValues{
		Token:         token,
		CorrelationID: correlationID,
		XSubject:      xSubject,
	}, nil
}

// MustGetRequestAuthValues returns the request auth values or if an error occurs
func MustGetRequestAuthValues(ctx *context.ServiceContext) (*RequestAuthValues, blame.Blame) {
	return GetRequestAuthValues(
		ctx,
		WithRequireToken(),
		WithRequireCorrelationID(),
		WithRequireXSubject(),
	)
}
