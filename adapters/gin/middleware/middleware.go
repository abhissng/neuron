package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/abhissng/neuron/adapters/gin/request"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/adapters/paseto"
	"github.com/abhissng/neuron/adapters/session"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
	"github.com/abhissng/neuron/utils/structures/claims"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Middleware to generate requestId and correlationId
func RequestIDMiddleware(log1 *log.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique requestId
		requestId := random.GenerateUUID()

		// Check if correlationId is passed in the headers
		correlationId := c.GetHeader(constant.CorrelationIDHeader)
		if correlationId == "" {
			correlationId = uuid.New().String() // Generate a new one if not provided
		}

		// Attach IDs to the context
		c.Set(constant.RequestID, requestId)
		c.Set(constant.CorrelationID, correlationId)

		// Log the IDs
		log1.Debug("Request ID and Correlation ID", log.String("request-id", requestId), log.String("correlation-id", correlationId))
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

	if ctx.PasetoMiddlewareOption() != nil && ctx.PasetoMiddlewareOption().HasExcludedOption() {
		blame := handleExcludedOptions(ctx)
		if blame != nil {
			return result.NewFailure[bool](blame)
		}
		return result.NewSuccess(helpers.Valid())
	}

	tokenResult := request.FetchPasetoBearerToken(ctx.Context)
	if !tokenResult.IsSuccess() {
		_, blameInfo := tokenResult.Value()
		return result.NewFailure[bool](blameInfo)
	}
	token := tokenResult.ToValue()

	subjectResult := request.FetchXSubjectHeader(ctx.Context)
	if !subjectResult.IsSuccess() {
		return result.CastFailure[string, bool](subjectResult)
	}

	res := ctx.ValidateToken(*token, nil, paseto.WithValidateEssentialTags)
	if !res.IsSuccess() {
		_, err := res.Value()
		return result.NewFailure[bool](err)
	}

	validToken := true
	return result.NewSuccess(&validToken)
}

// **Gin Middleware for Correlation ID**
func VerifyCorrelationId(ctx *context.ServiceContext) result.Result[bool] {
	if ctx.Context == nil {
		return result.NewFailure[bool](blame.MissingCorrelationID())
	}

	if ctx.Context.GetHeader(constant.CorrelationIDHeader) == "" {
		return result.NewFailure[bool](blame.MissingCorrelationID())
	}
	valid := true
	return result.NewSuccess(&valid)
}

func SessionMiddleware(sm *session.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(constant.SessionID)
		if err != nil {
			// No cookie found — just continue without session
			c.Next()
			return
		}

		// Try retrieving session data
		sessionData, err := sm.GetSession(c, sessionID)
		if err != nil {
			// Invalid or expired session — clear cookie (optional)
			c.SetCookie(constant.SessionID, "", -1, "/", "", false, true)
			c.Next()
			return
		}

		// Valid session — attach to context
		c.Set("session", sessionData)
		c.Next()
	}
}

func handleExcludedOptions(ctx *context.ServiceContext) blame.Blame {
	excluded := ctx.PasetoMiddlewareOption().ExcludedOptions()

	if excluded.HasExcludedService() {
		serviceName, err := ctx.GetGinCtxServiceName()
		if err != nil {
			return blame.MissingServiceName(err)
		}

		if helpers.IsFoundInSlice(serviceName.String(), excluded.ExcludedServices()) {
			return nil
		}
	}

	if excluded.HasExcludedRecords() {
		recordsName, err := ctx.GetGinCtxRecordsName()
		if err != nil {
			return blame.MissingRecordsName(err)
		}

		if helpers.IsFoundInSlice(*recordsName, excluded.ExcludedRecords()) {
			return nil
		}
	}

	return nil
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
