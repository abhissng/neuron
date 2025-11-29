package request

import (
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/types"
)

type EssentialHeaders struct {
	OrgId        types.OrgID  `json:"org_id"`
	UserId       types.UserID `json:"user_id"`
	UserRole     string       `json:"user_role"`
	FeatureFlags string       `json:"feature_flags"`
}

func NewEssentialHeaders() *EssentialHeaders {
	return &EssentialHeaders{
		OrgId:        types.OrgID{},
		UserId:       types.UserID{},
		UserRole:     "",
		FeatureFlags: "",
	}
}

func GetEssentialHeadersValues(ctx *context.ServiceContext) (*EssentialHeaders, blame.Blame) {
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

	userRoleResult := FetchXUserRoleHeader(ctx.Context)
	if !userRoleResult.IsSuccess() {
		ctx.SlogError("unable to get the user role", log.Blame(userRoleResult.Blame()))
		return nil, userRoleResult.Blame()
	}

	// it might be possible that feature flag is missing in some request
	featureFlag := ""
	featureFlagsResult := FetchXFeatureFlagsHeader(ctx.Context)
	if !featureFlagsResult.IsSuccess() {
		ctx.SlogWarn("unable to get the feature flags", log.Blame(featureFlagsResult.Blame()))
	} else {
		featureFlag = *featureFlagsResult.ToValue()
	}

	orgId, ok := types.CastTo[types.OrgID](*orgIdResult.ToValue())
	if !ok {
		return nil, blame.TypeConversionError("org_id", "uuid.UUID", "IDType", nil)
	}

	userId, ok := types.CastTo[types.UserID](*userIdResult.ToValue())
	if !ok {
		return nil, blame.TypeConversionError("user_id", "uuid.UUID", "IDType", nil)
	}

	return &EssentialHeaders{
		OrgId:        orgId,
		UserId:       userId,
		UserRole:     *userRoleResult.ToValue(),
		FeatureFlags: featureFlag,
	}, nil
}
