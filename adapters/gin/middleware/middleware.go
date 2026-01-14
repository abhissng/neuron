package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/abhissng/neuron/adapters/gin/request"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/paseto"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
	"github.com/abhissng/neuron/utils/structures"
	"github.com/abhissng/neuron/utils/structures/claims"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware creates a Gin middleware that generates a request ID and a correlation ID,
// stores them in the request context, and logs both identifiers.
//
// The middleware reads the correlation ID from the request header named by constant.CorrelationIDHeader;
// if absent, it generates a new UUID. Both IDs are stored in the Gin context under constant.RequestID
// and constant.CorrelationID. The provided logger is used to emit a debug log containing the IDs.
func RequestIDMiddleware(log1 *log.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique requestId
		requestId := random.GenerateUUIDString()

		// Check if correlationId is passed in the headers
		correlationId := c.GetHeader(constant.CorrelationIDHeader)
		if correlationId == "" {
			correlationId = uuid.New().String() // Generate a new one if not provided
		}

		// Attach IDs to the context
		c.Set(constant.RequestID, requestId)
		c.Set(constant.CorrelationID, correlationId)

		// Log the IDs
		log1.Debug("Request ID and Correlation ID", log.String(constant.RequestID, requestId), log.String(constant.CorrelationIDHeader, correlationId))
		// Pass control to the next middleware/handler
		c.Next()
	}
}

// Middleware to create ServiceContext for each API request
func ServiceContextMiddleware(opts ...context.ServiceContextOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize ServiceContext
		opts = append(opts, context.WithGinContext(c))
		serviceCtx := context.NewServiceContext(
			opts...,
		)

		// Attach ServiceContext to Gin's context
		c.Set(constant.ServiceContext, serviceCtx)

		// Pass control to the next middleware/handler
		c.Next()
	}
}

// **Gin Middleware for HSTS**
func HSTSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Next()
	}
}

// Middleware to add the service name to the context from request parameters
func ServiceNameMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fetch the service name from request parameters
		serviceNameRes := request.FetchServiceNameFromParams(c)
		if !serviceNameRes.IsSuccess() {
			// Handle the failure case, e.g., return an error response
			c.JSON(400, gin.H{"error": "Service name is missing or invalid"})
			c.Abort()
			return
		}

		serviceName := serviceNameRes.ToValue()

		// Add the service name to the context
		c.Set(constant.Service, serviceName)

		// Pass control to the next middleware/handler
		c.Next()
	}
}

// **Gin Middleware for Compression**
func CompressionMiddleware() gin.HandlerFunc {
	return gzip.Gzip(gzip.BestSpeed)
}

// TODO create correct logic for autorefresh
// basically  token services needs to be called for auto- refresh
// **Gin Middleware for Auto-Refresh**
func AutoRefreshMiddleware(ctx *context.ServiceContext) result.Result[bool] {

	if ctx.PasetoMiddlewareOption().IsAutoRefresh() {

		tokenResult := request.FetchTextParam(ctx.Context, constant.AuthorizationHeader, request.HeaderParam, true)
		if !tokenResult.IsSuccess() {
			return result.NewFailure[bool](blame.MissingAuthCredential(errors.New("authorization header is not present")))
		}

		bearerToken, _ := tokenResult.Value()
		token := helpers.ExtractBearerToken(*bearerToken)
		if helpers.IsEmpty(token) {
			return result.NewFailure[bool](blame.MalformedAuthToken(errors.New("authorization header is not present")))
		}

		subjectResult := request.FetchXSubjectHeader(ctx.Context)
		if !subjectResult.IsSuccess() {
			return result.CastFailure[string, bool](subjectResult)
		}

		res := ctx.ValidateToken(token, nil)
		if !res.IsSuccess() {
			_, err := res.Value()
			return result.NewFailure[bool](err)
		}
		claim, _ := res.Value()

		now := time.Now()
		// If token is close to expiring, refresh it
		if claim.Exp.Sub(now) < ctx.PasetoMiddlewareOption().RefreshThreshold() {
			// TODO call here token service for refreshing
			newRefreshTokenResult := ctx.FetchRefreshToken(
				claims.WithIP(claim.Ip),
				claims.WithSubject(claim.Sub),
				claims.WithAudience(claim.Aud),
				claims.WithNotBefore(claim.Nbf),
			)
			if !newRefreshTokenResult.IsSuccess() {
				_, err := newRefreshTokenResult.Value()
				// TODO add error handling
				// ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
				return result.NewFailure[bool](err)
			}
			newToken, _ := newRefreshTokenResult.Value()

			// Return new token in response header
			ctx.Header(ctx.PasetoMiddlewareOption().NewAuthHeader(), newToken.Token)
		}
	}
	validToken := true
	return result.NewSuccess(&validToken)
}

// **Gin Middleware for Paseto Verification**
func PasetoVerifyMiddleware(ctx *context.ServiceContext) result.Result[bool] {
	var err blame.Blame

	if ctx.PasetoMiddlewareOption() != nil && ctx.PasetoMiddlewareOption().HasExcludedOption() {
		blame := handleExcludedOptions(ctx, ctx.PasetoMiddlewareOption().ExcludedOptions())
		if blame == nil {
			return result.NewSuccess(helpers.Valid())
		}
		// 🧩 If exclusion handling fails → fall back to normal verification
		ctx.SlogWarn("excluded option check failed, falling back to normal paseto verification", log.Blame(blame))
	}

	tokenResult := request.FetchPasetoBearerToken(ctx.Context)
	if !tokenResult.IsSuccess() {
		err = tokenResult.Blame()
		return result.NewFailure[bool](err)
	}
	token := tokenResult.ToValue()

	subjectResult := request.FetchXSubjectHeader(ctx.Context)
	if !subjectResult.IsSuccess() {
		return result.NewFailure[bool](subjectResult.Blame())
	}

	extra := make(map[string]any)
	extra["subject"] = *(subjectResult.ToValue())
	extra["ip"] = ctx.ClientIP()
	extra["audience"] = ctx.Request.UserAgent()

	res := ctx.ValidateToken(*token, extra, paseto.WithValidateEssentialTags)
	if !res.IsSuccess() {
		ctx.SlogError("validation failed for paseto token", log.Blame(res.Blame()))
		return result.NewFailure[bool](res.Blame())
	}

	validToken := true
	return result.NewSuccess(&validToken)
}

// **Gin Middleware for Correlation ID**
func VerifyCorrelationId(ctx *context.ServiceContext) result.Result[bool] {
	if ctx.Context == nil {
		return result.NewFailure[bool](blame.MissingCorrelationID())
	}

	if ctx.GetHeader(constant.CorrelationIDHeader) == "" {
		return result.NewFailure[bool](blame.MissingCorrelationID())
	}
	valid := true
	return result.NewSuccess(&valid)
}

// SessionVerifyMiddleware validates user sessions stored by SessionManager
func SessionVerifyMiddleware(ctx *context.ServiceContext) result.Result[bool] {
	// 🧠 Handle exclusion rules (like in PasetoVerifyMiddleware)
	if ctx.SessionManager != nil && ctx.SessionMiddlewareOption() != nil && ctx.SessionMiddlewareOption().HasExcludedOption() {
		blame := handleExcludedOptions(ctx, ctx.SessionMiddlewareOption().ExcludedOptions())
		if blame == nil {
			return result.NewSuccess(helpers.Valid())
		}
		// 🧩 If exclusion handling fails → fall back to normal verification
		ctx.SlogWarn("excluded option check failed, falling back to normal session verification", log.Blame(blame))
	}

	var err error
	var sessionID string

	defer func() {
		if ctx.SessionManager != nil && err != nil && sessionID != "" {
			go func() {
				if destroyErr := ctx.DestroySession(ctx.Context, sessionID); destroyErr != nil {
					ctx.SlogError("failed to destroy session", log.Err(destroyErr))
				}
			}()
		}
	}()

	// 🧩 Extract session ID cookie
	sessionID, err = ctx.Cookie(constant.SessionID)
	if err != nil || sessionID == "" {
		ctx.SlogError("session cookie is missing", log.Err(err))
		return result.NewFailure[bool](blame.SessionMalformed(errors.New("session cookie is missing")))
	}

	// 🧩 Fetch session data
	sessionData, err := ctx.GetSession(ctx.Context, sessionID)
	if err != nil {
		ctx.SlogError("session not found", log.Err(err))
		// Clear expired/invalid cookie
		ctx.SetCookie(constant.SessionID, "", -1, "/", "", false, true)
		return result.NewFailure[bool](blame.SessionNotFound())
	}

	// 🧩 Optional custom validator
	res := ctx.ValidateSession(ctx.Context, sessionID, nil)
	if !res.IsSuccess() {
		ctx.SlogError("session validation failed", log.Blame(res.Blame()))
		return result.NewFailure[bool](res.Blame())
	}

	// 🧩 Attach session to Gin context (for downstream handlers)
	ctx.Set(constant.SessionID, sessionData)

	valid := true
	return result.NewSuccess(&valid)
}

func handleExcludedOptions(ctx *context.ServiceContext, excluded *structures.ExcludedOptions) blame.Blame {

	if excluded.HasExcludedService() {
		serviceName, err := ctx.GetGinCtxServiceName()
		if err != nil {
			return blame.MissingServiceName(err)
		}

		if helpers.IsFoundInSlice(serviceName.String(), excluded.ExcludedServices()) {
			return nil
		}
		return blame.NewBasicError("service not allowed")
	}

	if excluded.HasExcludedRecords() {
		recordsName, err := ctx.GetGinCtxRecordsName()
		if err != nil {
			return blame.MissingRecordsName(err)
		}

		if helpers.IsFoundInSlice(*recordsName, excluded.ExcludedRecords()) {
			return nil
		}
		return blame.NewBasicError("records not allowed")
	}

	return blame.NewBasicError("excluded options not allowed")
}

// BasicAuthMiddleware implements simple HTTP Basic Auth
func BasicAuthMiddleware(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(constant.AuthorizationHeader)

		// Check if Authorization header is present
		if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header missing or invalid",
			})
			return
		}

		// Decode the base64 credentials
		payload, err := codec.Decode[string]([]byte(strings.TrimPrefix(authHeader, "Basic ")), codec.Base64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid base64 credentials"})
			return
		}

		// Split "username:password"
		parts := strings.SplitN(string(payload), ":", 2)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credential format"})
			return
		}

		reqUser, reqPass := parts[0], parts[1]

		// Validate credentials
		if reqUser != username || reqPass != password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Auth successful → continue
		c.Next()
	}
}

// Middleware to Inject anything in the gin context
func InjectMiddleware(key string, value any) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize ServiceContext
		c.Set(key, value)

		// Pass control to the next middleware/handler
		c.Next()
	}
}