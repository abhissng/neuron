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
