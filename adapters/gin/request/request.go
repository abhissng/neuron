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

// Generic Type Constraints
type ConvertibleType interface {
	constraints.Integer | bool | string | uuid.UUID
}

// Parameter Source Enum
type ParamOrigin int

const (
	Unknown ParamOrigin = iota
	RouteParam
	QueryParam
	HeaderParam
)

// String Conversion
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

// Parameter Processing Functions
type ParamValidatorFunc[T ConvertibleType] func(T) bool
type ParamConverterFunc[T ConvertibleType] func(string) result.Result[T]

// Fetch Raw Parameter
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

// Generic Parameter Getter
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

// Integer Parameter
func FetchIntParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[int64] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[int64] {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return result.NewFailure[int64](blame.TypeConversionError(paramName, value, "int64", err))
		}
		return result.NewSuccess(&parsed)
	}, nil)
}

// String Parameter
func FetchTextParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[string] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[string] {
		if helpers.IsEmpty(value) {
			return result.NewFailure[string](blame.MissingParameterError(paramName))
		}
		return result.NewSuccess(&value)
	}, nil)
}

// Validated String Parameter
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

// UUID Parameter
func FetchUUIDParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[uuid.UUID] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[uuid.UUID] {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return result.NewFailure[uuid.UUID](blame.TypeConversionError(paramName, value, "uuid.UUID", err))
		}
		return result.NewSuccess(&parsed)
	}, nil)
}

// Boolean Parameter
func FetchBoolParam(c *gin.Context, paramName string, origin ParamOrigin, required bool) result.Result[bool] {
	return fetchAndConvertParam(c, paramName, required, origin, func(value string) result.Result[bool] {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return result.NewFailure[bool](blame.TypeConversionError(paramName, value, "bool", err))
		}
		return result.NewSuccess(&parsed)
	}, nil)
}

// Retrieve Arbitrary Value from Gin Context
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

// Extract Data from Request Body
func ExtractDataFromRequestBody[T any](c *gin.Context) result.Result[T] {
	var payload T
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		return result.NewFailure[T](blame.RequestBodyDataExtractionFailed(err))
	}
	return result.NewSuccess(&payload)
}

// Extract Data from Request Form
func ExtractDataFromForm[T any](c *gin.Context) result.Result[T] {
	var form T
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		return result.NewFailure[T](blame.RequestFormDataExtractionFailed(bindErr))
	}
	return result.NewSuccess(&form)
}

// Fetch Business ID from Params
func FetchBusinessIDFromParams(c *gin.Context) result.Result[types.BusinessID] {
	idResult := FetchIntParam(c, constant.BusinessID, RouteParam, true)
	if idResult.IsSuccess() {
		val, _ := idResult.Value()
		return result.NewSuccess(types.CreateRef(types.BusinessID(*val)))
	}
	_, err := idResult.Value()
	return result.NewFailure[types.BusinessID](blame.BusinessIdPathParamMissing(err.FetchCauses()...))
}

// Parse Unix Time from Params
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

// Retrieve User ID from Context
func RetrieveUserIdFromContext(c *gin.Context) result.Result[types.UserID] {
	userIdResult := RetrieveFromGinContext[types.UserID](c, constant.UserID)
	if userIdResult.IsSuccess() {
		return userIdResult
	}
	_, err := userIdResult.Value()
	return result.NewFailure[types.UserID](blame.UserIdContextMissing(constant.UserID, err.FetchCauses()...))
}

// Fetch User ID from Params
func FetchUserIdFromParams(c *gin.Context) result.Result[types.UserID] {
	userIdParam := FetchIntParam(c, constant.UserID, QueryParam, true)
	if userIdParam.IsSuccess() {
		value, _ := userIdParam.Value()
		return result.NewSuccess(types.CreateRef(types.UserID(*value)))
	}
	_, err := userIdParam.Value()
	return result.NewFailure[types.UserID](blame.UserIdQueryParamMissing(constant.UserID, err.FetchCauses()...))
}

// Fetch Business ID from Headers
func FetchBusinessIdFromHeaders(c *gin.Context) result.Result[types.BusinessID] {
	businessIdHeader := FetchIntParam(c, constant.BusinessID, HeaderParam, true)
	if businessIdHeader.IsSuccess() {
		businessIdValue, _ := businessIdHeader.Value()
		return result.NewSuccess(types.CreateRef(types.BusinessID(*businessIdValue)))
	}
	_, err := businessIdHeader.Value()
	return result.NewFailure[types.BusinessID](blame.BusinessIdHeaderMissing(constant.BusinessID, err.FetchCauses()...))
}

// Retrieve User ID from Headers
func RetrieveUserIdFromHeaders(c *gin.Context) result.Result[types.UserID] {
	userIdHeader := FetchIntParam(c, constant.UserID, HeaderParam, true)
	if userIdHeader.IsSuccess() {
		userValue, _ := userIdHeader.Value()
		return result.NewSuccess(types.CreateRef(types.UserID(*userValue)))
	}
	_, err := userIdHeader.Value()
	return result.NewFailure[types.UserID](blame.UserIdHeaderMissing(constant.UserID, err.FetchCauses()...))
}

// Fetch Correlation ID from Headers
func FetchCorrelationIdFromHeaders(c *gin.Context) result.Result[types.CorrelationID] {
	correlationIdHeader := FetchTextParam(c, constant.CorrelationID, HeaderParam, true)
	if correlationIdHeader.IsSuccess() {
		entityValue, _ := correlationIdHeader.Value()
		return result.NewSuccess(types.CreateRef(types.CorrelationID(*entityValue)))
	}
	_, err := correlationIdHeader.Value()
	return result.NewFailure[types.CorrelationID](blame.CorrelationIDHeaderMissing(constant.CorrelationID, err.FetchCauses()...))
}

// Retrieve Signature from Headers
func RetrieveSignatureFromHeaders(c *gin.Context) result.Result[string] {
	signatureHeader := FetchTextParam(c, constant.XSignature, HeaderParam, true)
	if signatureHeader.IsSuccess() {
		signature, _ := signatureHeader.Value()
		return result.NewSuccess(signature)
	}
	return result.NewFailure[string](blame.AuthSignatureMissing())
}

// Extract PASETO Auth Token from Headers
func ExtractPasetoAuthTokenFromHeaders(c *gin.Context) result.Result[string] {
	authTokenHeader := FetchTextParam(c, constant.XPasetoToken, HeaderParam, true)
	if authTokenHeader.IsSuccess() {
		token, _ := authTokenHeader.Value()
		return result.NewSuccess(token)
	}
	return result.NewFailure[string](blame.MissingAuthCredential(errors.New("paseto authorization header is missing")))
}

// Fetch X-Subject Header
func FetchXSubjectHeader(c *gin.Context) result.Result[string] {
	subjectHeader := FetchTextParam(c, constant.XSubject, HeaderParam, true)
	if subjectHeader.IsSuccess() {
		subject, _ := subjectHeader.Value()
		return result.NewSuccess(subject)
	}
	return result.NewFailure[string](blame.XSubjectHeaderMissing(errors.New("subject header is not present")))
}

// Extract Refresh Token from Headers
func ExtractRefreshTokenFromHeaders(c *gin.Context) result.Result[string] {
	authTokenHeader := FetchTextParam(c, constant.XRefreshToken, HeaderParam, true)
	if authTokenHeader.IsSuccess() {
		token, _ := authTokenHeader.Value()
		return result.NewSuccess(token)
	}

	return result.NewFailure[string](blame.MissingAuthCredential(errors.New("paseto authorization header is missing for refresh token")))
}

// Service Parameter
func FetchServiceNameFromParams(c *gin.Context) result.Result[types.Service] {
	serviceParam := FetchTextParam(c, constant.Service, RouteParam, true)
	if serviceParam.IsSuccess() {
		value, _ := serviceParam.Value()
		return result.NewSuccess(types.CreateRef(types.Service(*value)))
	}
	_, err := serviceParam.Value()
	return result.NewFailure[types.Service](blame.ServiceQueryParamMissing(constant.Service, err.FetchCauses()...))
}

// Fetch Paseto Bearer Token from Headers of gin context
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

// Fetch X-User-Role Header
func FetchXUserRoleHeader(c *gin.Context) result.Result[string] {
	userRoleHeader := FetchTextParam(c, constant.XUserRole, HeaderParam, true)
	if userRoleHeader.IsSuccess() {
		userRole := userRoleHeader.ToValue()
		return result.NewSuccess(userRole)
	}
	return result.NewFailure[string](blame.MissingXUserRole())
}

// Fetch X-Org-Id Header
func FetchXOrgIdHeader(c *gin.Context) result.Result[string] {
	orgIdHeader := FetchTextParam(c, constant.XOrgId, HeaderParam, true)
	if orgIdHeader.IsSuccess() {
		orgId := orgIdHeader.ToValue()
		return result.NewSuccess(orgId)
	}
	return result.NewFailure[string](blame.MissingXOrgId())
}

// Fetch X-User-Id Header
func FetchXUserIdHeader(c *gin.Context) result.Result[string] {
	userIdHeader := FetchTextParam(c, constant.XUserId, HeaderParam, true)
	if userIdHeader.IsSuccess() {
		userId := userIdHeader.ToValue()
		return result.NewSuccess(userId)
	}
	return result.NewFailure[string](blame.MissingXUserId())
}
