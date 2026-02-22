package nats

import (
	"errors"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures"
	"github.com/abhissng/neuron/utils/types"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// FetchXOrgIdHeaderFromNatsMsg fetches the X-Org-Id header from the NATS message and parses it as UUID.
func FetchXOrgIdHeaderFromNatsMsg(msg *nats.Msg) result.Result[uuid.UUID] {
	return fetchUUIDFromNatsHeader(msg, constant.XOrgId)
}

// FetchXUserIdHeaderFromNatsMsg fetches the X-User-Id header from the NATS message and parses it as UUID.
func FetchXUserIdHeaderFromNatsMsg(msg *nats.Msg) result.Result[uuid.UUID] {
	return fetchUUIDFromNatsHeader(msg, constant.XUserId)
}

// FetchXUserRoleHeaderFromNatsMsg fetches the X-User-Role header from the NATS message.
func FetchXUserRoleHeaderFromNatsMsg(msg *nats.Msg) result.Result[string] {
	return fetchStringFromNatsHeader(msg, constant.XUserRole)
}

// FetchXFeatureFlagsHeaderFromNatsMsg fetches the X-Feature-Flags header from the NATS message.
func FetchXFeatureFlagsHeaderFromNatsMsg(msg *nats.Msg) result.Result[string] {
	return fetchStringFromNatsHeader(msg, constant.XFeatureFlags)
}

// FetchXLocationIdHeaderFromNatsMsg fetches the X-Location-Id header from the NATS message and parses it as UUID.
func FetchXLocationIdHeaderFromNatsMsg(msg *nats.Msg) result.Result[uuid.UUID] {
	return fetchUUIDFromNatsHeader(msg, constant.XLocationId)
}

func fetchStringFromNatsHeader(msg *nats.Msg, key string) result.Result[string] {
	if msg == nil || msg.Header == nil {
		return result.NewFailure[string](blame.HeadersNotFound(errors.New("nats headers not found")))
	}
	val := msg.Header.Get(key)
	if helpers.IsEmpty(val) {
		return result.NewFailure[string](blame.HeadersNotFound(errors.New("nats header value is empty for key: " + key)))
	}
	return result.NewSuccess(&val)
}

func fetchUUIDFromNatsHeader(msg *nats.Msg, key string) result.Result[uuid.UUID] {
	if msg == nil || msg.Header == nil {
		return result.NewFailure[uuid.UUID](blame.HeadersNotFound(errors.New("nats headers not found")))
	}
	val := msg.Header.Get(key)
	if helpers.IsEmpty(val) {
		return result.NewFailure[uuid.UUID](blame.HeadersNotFound(errors.New("nats header value is empty for key: " + key)))
	}
	parsed, err := uuid.Parse(val)
	if err != nil {
		return result.NewFailure[uuid.UUID](blame.TypeConversionError(key, val, "uuid.UUID", err))
	}
	return result.NewSuccess(&parsed)
}

// GetEssentialHeadersValuesFromNatsMsg extracts essential header values from the NATS message using the same semantics as GetEssentialHeadersValues.
// If logger is non-nil, errors are logged via logger.Error; otherwise helpers.Println with constant.ERROR is used.
func GetEssentialHeadersValuesFromNatsMsg(msg *nats.Msg, logger *log.Log, options ...structures.EssentialHeadersOption) (*structures.EssentialHeaders, blame.Blame) {
	defer func() { helpers.RecoverException(recover()) }()
	if logger == nil {
		return nil, blame.GeneralKnownError(errors.New("logger is nil"))
	}
	cfg := structures.NewEssentialHeadersConfig()
	for _, o := range options {
		o(cfg)
	}

	orgIdResult := FetchXOrgIdHeaderFromNatsMsg(msg)
	if !orgIdResult.IsSuccess() {
		logger.Error("unable to get the org id", log.Blame(orgIdResult.Blame()))
		return nil, orgIdResult.Blame()
	}

	userIdResult := FetchXUserIdHeaderFromNatsMsg(msg)
	if !userIdResult.IsSuccess() {
		logger.Error("unable to get the user id", log.Blame(userIdResult.Blame()))
		return nil, userIdResult.Blame()
	}

	userRoleResult := FetchXUserRoleHeaderFromNatsMsg(msg)
	if !userRoleResult.IsSuccess() {
		logger.Error("unable to get the user role", log.Blame(userRoleResult.Blame()))
		return nil, userRoleResult.Blame()
	}
	userRole := *userRoleResult.ToValue()

	featureFlag := ""
	featureFlagsResult := FetchXFeatureFlagsHeaderFromNatsMsg(msg)
	if !featureFlagsResult.IsSuccess() {
		if cfg.RequireFeatureFlags {
			logger.Error("missing feature flags header (required)", log.Blame(featureFlagsResult.Blame()))
			return nil, featureFlagsResult.Blame()
		}
	} else {
		featureFlag = *featureFlagsResult.ToValue()
	}

	var locationId uuid.UUID
	locationResult := FetchXLocationIdHeaderFromNatsMsg(msg)
	if !locationResult.IsSuccess() {
		if cfg.RequireLocationID {
			logger.Error("missing location id header (required)", log.Blame(locationResult.Blame()))
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

	return &structures.EssentialHeaders{
		OrgId:        orgId,
		UserId:       userId,
		UserRole:     userRole,
		FeatureFlags: featureFlag,
		LocationId:   locationId,
	}, nil
}

// FetchCorrelationIdFromNatsMsg fetches the X-Correlation-ID header from the NATS message.
func FetchCorrelationIdFromNatsMsg(msg *nats.Msg) result.Result[types.CorrelationID] {
	if msg == nil || msg.Header == nil {
		return result.NewFailure[types.CorrelationID](blame.CorrelationIDHeaderMissing(constant.CorrelationIDHeader, errors.New("nats message or header is nil")))
	}
	val := msg.Header.Get(constant.CorrelationIDHeader)
	if helpers.IsEmpty(val) {
		return result.NewFailure[types.CorrelationID](blame.CorrelationIDHeaderMissing(constant.CorrelationIDHeader, errors.New("correlation id header is empty")))
	}
	return result.NewSuccess(types.CreateRef(types.CorrelationID(val)))
}

// FetchXSubjectHeaderFromNatsMsg fetches the X-Subject header from the NATS message.
func FetchXSubjectHeaderFromNatsMsg(msg *nats.Msg) result.Result[string] {
	if msg == nil || msg.Header == nil {
		return result.NewFailure[string](blame.HeadersNotFound(errors.New("nats headers not found")))
	}
	val := msg.Header.Get(constant.XSubject)
	if helpers.IsEmpty(val) {
		return result.NewFailure[string](blame.XSubjectHeaderMissing(errors.New("subject header is not present")))
	}
	return result.NewSuccess(&val)
}

// FetchPasetoBearerTokenFromNatsMsg fetches and extracts the PASETO bearer token from the Authorization header of the NATS message.
func FetchPasetoBearerTokenFromNatsMsg(msg *nats.Msg) result.Result[string] {
	if msg == nil || msg.Header == nil {
		return result.NewFailure[string](blame.MissingAuthCredential(errors.New("authorization header is not present")))
	}
	authVal := msg.Header.Get(constant.AuthorizationHeader)
	if helpers.IsEmpty(authVal) {
		return result.NewFailure[string](blame.MissingAuthCredential(errors.New("authorization header is not present")))
	}
	token := helpers.ExtractBearerToken(authVal)
	if helpers.IsEmpty(token) {
		return result.NewFailure[string](blame.MalformedAuthToken(errors.New("authorization header is not present")))
	}
	return result.NewSuccess(&token)
}

// GetRequestAuthValuesFromNatsMsg extracts request auth values (token, correlation ID, X-Subject) from the NATS message.
// If logger is non-nil, errors are logged via logger.Error; otherwise helpers.Println with constant.ERROR is used.
func GetRequestAuthValuesFromNatsMsg(msg *nats.Msg, logger *log.Log, options ...structures.RequestAuthOption) (*structures.RequestAuthValues, blame.Blame) {
	defer func() { helpers.RecoverException(recover()) }()
	if logger == nil {
		return nil, blame.GeneralKnownError(errors.New("logger is nil"))
	}
	cfg := &structures.RequestAuthConfig{
		RequireToken:         false,
		RequireCorrelationID: false,
		RequireXSubject:      false,
	}
	for _, opt := range options {
		opt(cfg)
	}

	var token string
	tokenResult := FetchPasetoBearerTokenFromNatsMsg(msg)
	if !tokenResult.IsSuccess() {
		if cfg.RequireToken {
			logger.Error("unable to get the authentication token", log.Blame(tokenResult.Blame()))
			return nil, tokenResult.Blame()
		}
	} else {
		token = *tokenResult.ToValue()
	}

	var correlationID types.CorrelationID
	correlationIdResult := FetchCorrelationIdFromNatsMsg(msg)
	if !correlationIdResult.IsSuccess() {
		if cfg.RequireCorrelationID {
			logger.Error("unable to get the correlation ID", log.Blame(correlationIdResult.Blame()))
			return nil, correlationIdResult.Blame()
		}
	} else {
		correlationID = *correlationIdResult.ToValue()
	}

	var xSubject string
	xSubjectResult := FetchXSubjectHeaderFromNatsMsg(msg)
	if !xSubjectResult.IsSuccess() {
		if cfg.RequireXSubject {
			logger.Error("unable to get the X-Subject header", log.Blame(xSubjectResult.Blame()))
			return nil, xSubjectResult.Blame()
		}
	} else {
		xSubject = *xSubjectResult.ToValue()
	}

	return &structures.RequestAuthValues{
		Token:         token,
		CorrelationID: correlationID,
		XSubject:      xSubject,
	}, nil
}

// MustGetRequestAuthValuesFromNatsMsg returns request auth values with all auth options required, or returns blame on failure.
func MustGetRequestAuthValuesFromNatsMsg(msg *nats.Msg, logger *log.Log) (*structures.RequestAuthValues, blame.Blame) {
	return GetRequestAuthValuesFromNatsMsg(
		msg,
		logger,
		structures.WithRequireToken(),
		structures.WithRequireCorrelationID(),
		structures.WithRequireXSubject(),
	)
}
