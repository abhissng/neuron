package request

import (
	"errors"
	"strconv"

	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
)

// ConvertibleType defines the generic type constraints for parameters that can be converted.
// It includes integers, booleans, strings, and UUIDs.
type ConvertibleType interface {
	constraints.Integer | bool | string | uuid.UUID
}

// ParamOrigin represents the source of a parameter (route, query, or header).
type ParamOrigin int

const (
	Unknown ParamOrigin = iota
	RouteParam
	QueryParam
	HeaderParam
)

// String returns the string representation of ParamOrigin.
func (p ParamOrigin) String() string {
	switch p {
	case RouteParam:
		return "route"
	case QueryParam:
		return "query"
	case HeaderParam:
		return "header"
	default:
		return "unknown"
	}
}

// ParamValidatorFunc is a function type for validating parameters of type T.
type ParamValidatorFunc[T ConvertibleType] func(T) bool

// ParamConverterFunc is a function type for converting string values to type T.
type ParamConverterFunc[T ConvertibleType] func(string) result.Result[T]

// fetchParam retrieves a raw parameter value from the gin context based on its origin.
// It returns a Result containing the parameter value or an error if not found.
func fetchParam(c *gin.Context, paramName string, origin ParamOrigin) result.Result[string] {
	switch origin {
	case RouteParam:
		val := c.Param(paramName)
		return result.NewSuccess(&val)
	case QueryParam:
		val, exists := c.GetQuery(paramName)
		if exists {
			return result.NewSuccess(&val)
		}
		return result.NewFailure[string](blame.MissingParameterError(paramName))
	case HeaderParam:
		val := c.GetHeader(paramName)
		return result.NewSuccess(&val)
	default:
		return result.NewFailure[string](blame.InvalidSourceError(origin.String()))
	}
}

// fetchAndConvertParam is a generic function that fetches, converts, and validates parameters.
// It handles the complete parameter processing pipeline including validation.
func fetchAndConvertParam[T ConvertibleType](
	c *gin.Context,
	paramName string,
	required bool,
	origin ParamOrigin,
	converter ParamConverterFunc[T],
	validator ParamValidatorFunc[T],
) result.Result[T] {
	rawParamResult := fetchParam(c, paramName, origin)

	rawValue, err := rawParamResult.Value()
	if err != nil {
		if required {
			return result.NewFailure[T](blame.MissingParameterError(paramName))
		}
		return result.NewSuccess[T](nil)
	}

	convertedResult := converter(*rawValue)
	convertedValue, convErr := convertedResult.Value()
	if convErr != nil {
		return result.NewFailure[T](blame.MalformedParameterError(paramName))
	}

	if validator != nil && !validator(*convertedValue) {
		return result.NewFailure[T](blame.MalformedParameterError(paramName))
	}

	return result.NewSuccess(convertedValue)
}

// FetchIntParam fetches and converts a parameter to int64.
// It returns a Result containing the converted integer or an error if conversion fails.
func FetchIntParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[int64] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[int64] {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return result.NewFailure[int64](blame.TypeConversionError(paramName, value, "int64", err))
		}
		return result.NewSuccess(&parsed)
	}, nil)
}

// FetchTextParam fetches and validates a string parameter.
// It ensures the parameter is not empty and returns a Result with the string value.
func FetchTextParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[string] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[string] {
		if helpers.IsEmpty(value) {
			return result.NewFailure[string](blame.MissingParameterError(paramName))
		}
		return result.NewSuccess(&value)
	}, nil)
}

// FetchValidatedTextParam fetches a string parameter and applies custom validation.
// It combines string fetching with user-provided validation logic.
func FetchValidatedTextParam(
	c *gin.Context,
	paramName string,
	origin ParamOrigin,
	required bool,
	validator ParamValidatorFunc[string],
) result.Result[string] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[string] {
		if helpers.IsEmpty(value) {
			return result.NewFailure[string](blame.MissingParameterError(paramName))
		}
		return result.NewSuccess(&value)
	}, validator)
}

// FetchUUIDParam fetches and converts a parameter to UUID.
// It parses the string value as a UUID and returns a Result with the parsed UUID.
func FetchUUIDParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[uuid.UUID] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[uuid.UUID] {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return result.NewFailure[uuid.UUID](blame.TypeConversionError(paramName, value, "uuid.UUID", err))
		}
		return result.NewSuccess(&parsed)
	}, nil)
}

// FetchBoolParam fetches and converts a parameter to boolean.
// It parses string values like "true", "false", "1", "0" to boolean.
func FetchBoolParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[bool] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[bool] {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return result.NewFailure[bool](blame.TypeConversionError(paramName, value, "bool", err))
		}
		return result.NewSuccess(&parsed)
	}, nil)
}

// RetrieveFromGinContext retrieves an arbitrary value from the gin context by key.
// It performs type assertion and returns a Result with the typed value.
func RetrieveFromGinContext[T any](c *gin.Context, key string) result.Result[T] {
	val, exists := c.Get(key)
	if !exists {
		return result.NewFailure[T](blame.GinContextKeyFetchError(key, nil))
	}

	typedVal, ok := val.(T)
	if !ok {
		serialized, _ := codec.Encode(val, codec.JSON)
		return result.NewFailure[T](blame.TypeConversionError(key, string(serialized), "bool", nil))
	}

	return result.NewSuccess(&typedVal)
}

// ExtractDataFromRequestBody extracts and unmarshals JSON data from the request body.
// It binds the JSON payload to the specified type T.
func ExtractDataFromRequestBody[T any](c *gin.Context) result.Result[T] {
	var payload T
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		return result.NewFailure[T](blame.RequestBodyDataExtractionFailed(err))
	}
	return result.NewSuccess(&payload)
}

// ExtractDataFromForm extracts and binds form data from the request.
// It supports both URL-encoded and multipart form data.
func ExtractDataFromForm[T any](c *gin.Context) result.Result[T] {
	var form T
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		return result.NewFailure[T](blame.RequestFormDataExtractionFailed(bindErr))
	}
	return result.NewSuccess(&form)
}

// FetchBusinessIDFromParams fetches the business ID from route parameters.
// It converts the parameter to BusinessID type and validates it.
func FetchBusinessIDFromParams(c *gin.Context) result.Result[types.BusinessID] {
	idResult := FetchUUIDParam(c, constant.BusinessID, RouteParam, true)
	if idResult.IsSuccess() {
		val, _ := idResult.Value()
		return result.NewSuccess(types.CreateRef(types.BusinessID(*val)))
	}
	_, err := idResult.Value()
	return result.NewFailure[types.BusinessID](blame.BusinessIdPathParamMissing(err.FetchCauses()...))
}

// ParseUnixTimeFromParams parses a Unix timestamp from query parameters.
// It converts the timestamp to Milliseconds type for time handling.
func ParseUnixTimeFromParams(key string, mandatory bool, c *gin.Context) result.Result[types.Milliseconds] {
	paramResult := FetchIntParam(c, key, QueryParam, mandatory)
	if paramResult.IsSuccess() {
		value, _ := paramResult.Value()
		if value != nil {
			unixTime := types.Milliseconds(*value)
			return result.NewSuccess(&unixTime)
		}
		return result.NewSuccess[types.Milliseconds](nil)
	}
	return result.NewFailure[types.Milliseconds](blame.TimeQueryParamInvalid())
}

// RetrieveUserIdFromContext retrieves the user ID from the gin context.
// This is typically set by authentication middleware.
func RetrieveUserIdFromContext(c *gin.Context) result.Result[types.UserID] {
	userIdResult := RetrieveFromGinContext[types.UserID](c, constant.UserID)
	if userIdResult.IsSuccess() {
		return userIdResult
	}
	_, err := userIdResult.Value()
	return result.NewFailure[types.UserID](blame.UserIdContextMissing(constant.UserID, err.FetchCauses()...))
}

// FetchUserIdFromParams fetches the user ID from query parameters.
// It validates and converts the parameter to UserID type.
func FetchUserIdFromParams(c *gin.Context) result.Result[types.UserID] {
	userIdParam := FetchUUIDParam(c, constant.UserID, QueryParam, true)
	if userIdParam.IsSuccess() {
		value, _ := userIdParam.Value()
		return result.NewSuccess(types.CreateRef(types.UserID(*value)))
	}
	_, err := userIdParam.Value()
	return result.NewFailure[types.UserID](blame.UserIdQueryParamMissing(constant.UserID, err.FetchCauses()...))
}

// FetchBusinessIdFromHeaders fetches the business ID from request headers.
// It's commonly used for multi-tenant applications.
func FetchBusinessIdFromHeaders(c *gin.Context) result.Result[types.BusinessID] {
	businessIdHeader := FetchUUIDParam(c, constant.BusinessID, HeaderParam, true)
	if businessIdHeader.IsSuccess() {
		businessIdValue, _ := businessIdHeader.Value()
		return result.NewSuccess(types.CreateRef(types.BusinessID(*businessIdValue)))
	}
	_, err := businessIdHeader.Value()
	return result.NewFailure[types.BusinessID](blame.BusinessIdHeaderMissing(constant.BusinessID, err.FetchCauses()...))
}

// RetrieveUserIdFromHeaders retrieves the user ID from request headers.
// It validates the header value and converts it to UserID type.
func RetrieveUserIdFromHeaders(c *gin.Context) result.Result[types.UserID] {
	userIdHeader := FetchUUIDParam(c, constant.UserID, HeaderParam, true)
	if userIdHeader.IsSuccess() {
		userValue, _ := userIdHeader.Value()
		return result.NewSuccess(types.CreateRef(types.UserID(*userValue)))
	}
	_, err := userIdHeader.Value()
	return result.NewFailure[types.UserID](blame.UserIdHeaderMissing(constant.UserID, err.FetchCauses()...))
}

// FetchCorrelationIdFromHeaders fetches the correlation ID from request headers.
// Correlation IDs are used for distributed tracing and request tracking.
func FetchCorrelationIdFromHeaders(c *gin.Context) result.Result[types.CorrelationID] {
	correlationIdHeader := FetchTextParam(c, constant.CorrelationID, HeaderParam, true)
	if correlationIdHeader.IsSuccess() {
		entityValue, _ := correlationIdHeader.Value()
		return result.NewSuccess(types.CreateRef(types.CorrelationID(*entityValue)))
	}
	_, err := correlationIdHeader.Value()
	return result.NewFailure[types.CorrelationID](blame.CorrelationIDHeaderMissing(constant.CorrelationID, err.FetchCauses()...))
}

// RetrieveSignatureFromHeaders retrieves the signature from request headers.
// This is used for request authentication and verification.
func RetrieveSignatureFromHeaders(c *gin.Context) result.Result[string] {
	signatureHeader := FetchTextParam(c, constant.XSignature, HeaderParam, true)
	if signatureHeader.IsSuccess() {
		signature, _ := signatureHeader.Value()
		return result.NewSuccess(signature)
	}
	return result.NewFailure[string](blame.AuthSignatureMissing())
}

// ExtractPasetoAuthTokenFromHeaders extracts the PASETO authentication token from headers.
// PASETO tokens are used for secure authentication and authorization.
func ExtractPasetoAuthTokenFromHeaders(c *gin.Context) result.Result[string] {
	authTokenHeader := FetchTextParam(c, constant.XPasetoToken, HeaderParam, true)
	if authTokenHeader.IsSuccess() {
		token, _ := authTokenHeader.Value()
		return result.NewSuccess(token)
	}
	return result.NewFailure[string](blame.MissingAuthCredential(errors.New("paseto authorization header is missing")))
}

// FetchXSubjectHeader fetches the X-Subject header from the request.
// This header typically contains the subject identifier for the request.
func FetchXSubjectHeader(c *gin.Context) result.Result[string] {
	subjectHeader := FetchTextParam(c, constant.XSubject, HeaderParam, true)
	if subjectHeader.IsSuccess() {
		subject, _ := subjectHeader.Value()
		return result.NewSuccess(subject)
	}
	return result.NewFailure[string](blame.XSubjectHeaderMissing(errors.New("subject header is not present")))
}

// ExtractRefreshTokenFromHeaders extracts the refresh token from request headers.
// Refresh tokens are used to obtain new access tokens without re-authentication.
func ExtractRefreshTokenFromHeaders(c *gin.Context) result.Result[string] {
	authTokenHeader := FetchTextParam(c, constant.XRefreshToken, HeaderParam, true)
	if authTokenHeader.IsSuccess() {
		token, _ := authTokenHeader.Value()
		return result.NewSuccess(token)
	}

	return result.NewFailure[string](blame.MissingAuthCredential(errors.New("paseto authorization header is missing for refresh token")))
}

// FetchServiceNameFromParams fetches the service name from route parameters.
// This is used for service identification and routing.
func FetchServiceNameFromParams(c *gin.Context) result.Result[types.Service] {
	serviceParam := FetchTextParam(c, constant.Service, RouteParam, true)
	if serviceParam.IsSuccess() {
		value, _ := serviceParam.Value()
		return result.NewSuccess(types.CreateRef(types.Service(*value)))
	}
	_, err := serviceParam.Value()
	return result.NewFailure[types.Service](blame.ServiceQueryParamMissing(constant.Service, err.FetchCauses()...))
}

// FetchPasetoBearerToken fetches and extracts the PASETO bearer token from authorization headers.
// It validates the Bearer token format and extracts the actual token value.
func FetchPasetoBearerToken(c *gin.Context) result.Result[string] {
	tokenResult := FetchTextParam(c, constant.AuthorizationHeader, HeaderParam, true)
	if !tokenResult.IsSuccess() {
		return result.NewFailure[string](blame.MissingAuthCredential(errors.New("authorization header is not present")))
	}

	bearerToken, _ := tokenResult.Value()
	token := helpers.ExtractBearerToken(*bearerToken)
	if helpers.IsEmpty(token) {
		return result.NewFailure[string](blame.MalformedAuthToken(errors.New("authorization header is not present")))
	}

	return result.NewSuccess(&token)
}

// FetchXUserRoleHeader fetches the X-User-Role header from the request.
// This header contains the user's role information for authorization.
func FetchXUserRoleHeader(c *gin.Context) result.Result[string] {
	userRoleHeader := FetchTextParam(c, constant.XUserRole, HeaderParam, true)
	if userRoleHeader.IsSuccess() {
		userRole := userRoleHeader.ToValue()
		return result.NewSuccess(userRole)
	}
	return result.NewFailure[string](blame.MissingXUserRole())
}

// FetchXOrgIdHeader fetches the X-Org-Id header and converts it to UUID.
// This header identifies the organization in multi-tenant applications.
func FetchXOrgIdHeader(c *gin.Context) result.Result[uuid.UUID] {
	orgIdHeader := FetchUUIDParam(c, constant.XOrgId, HeaderParam, true)
	if orgIdHeader.IsSuccess() {
		orgId := orgIdHeader.ToValue()
		return result.NewSuccess(orgId)
	}
	return result.NewFailure[uuid.UUID](blame.MissingXOrgId())
}

// FetchXUserIdHeader fetches the X-User-Id header and converts it to UUID.
// This header contains the user identifier in UUID format.
func FetchXUserIdHeader(c *gin.Context) result.Result[uuid.UUID] {
	userIdHeader := FetchUUIDParam(c, constant.XUserId, HeaderParam, true)
	if userIdHeader.IsSuccess() {
		userId := userIdHeader.ToValue()
		return result.NewSuccess(userId)
	}
	return result.NewFailure[uuid.UUID](blame.MissingXUserId())
}

// FetchXFeatureFlagsHeader fetches the X-Feature-Flags header from the request.
// This header contains the feature flags for the user.
func FetchXFeatureFlagsHeader(c *gin.Context) result.Result[string] {
	featureFlagsHeader := FetchTextParam(c, constant.XFeatureFlags, HeaderParam, true)
	if featureFlagsHeader.IsSuccess() {
		featureFlags := featureFlagsHeader.ToValue()
		return result.NewSuccess(featureFlags)
	}
	return result.NewFailure[string](blame.MissingFeatureFlags())
}
