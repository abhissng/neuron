package middleware

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	csrf "github.com/spacecafe/gobox/gin-csrf"
	problems "github.com/spacecafe/gobox/gin-problems"
	csrfLog "github.com/spacecafe/gobox/logger"
)

// CreateCSRFConfig initializes the CSRF configuration settings.
func CreateCSRFConfig(secretKey string, excludedRoutes []string) csrf.Config {
	return csrf.Config{
		CookieName:     "X-CSRF-Token",
		HeaderName:     "X-CSRF-Token",
		SameSite:       "None",
		TokenLength:    32,
		Signer:         sha256.New,
		SecretKey:      []byte(secretKey),
		Logger:         csrfLog.Default(),
		ExcludedRoutes: excludedRoutes,
		Path:           "/",
		SecureCookie:   true,
	}
}

// CSRFMiddleware handles CSRF token generation, validation, and enforcement.
func CSRFMiddleware(log *log.Log, secretKey string, excludedRoutes []string) gin.HandlerFunc {
	csrfConfig := CreateCSRFConfig(secretKey, excludedRoutes)

	return func(ctx *gin.Context) {
		fullPath := ctx.FullPath()
		isExcluded := slices.Contains(csrfConfig.ExcludedRoutes, fullPath)

		if fullPath == "/" {
			handleRootRequest(ctx, &csrfConfig, log)
		} else if fullPath != "" && !isExcluded {
			validateCSRFToken(ctx, &csrfConfig, log)
		}

		ctx.Next()
	}
}

// handleRootRequest processes requests to the root path, generating a CSRF token if needed.
func handleRootRequest(ctx *gin.Context, config *csrf.Config, logger *log.Log) {
	clientIP := ctx.ClientIP()
	allowedOrigins := viper.GetString("CSRF_ALLOWED_ORIGINS")

	// Check if the client IP is allowed
	if allowedOrigins != "*" && !slices.Contains(strings.Split(allowedOrigins, ", "), clientIP) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Generate and set CSRF token as a cookie
	token, err := csrf.NewToken(config)
	if err != nil {
		logger.Warn("Failed to generate CSRF token", log.Err(err))
		_ = ctx.Error(problems.ProblemCSRFMalfunction)
		ctx.Abort()
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     config.CookieName,
		Value:    token.String(),
		Path:     config.Path,
		Secure:   ctx.Request.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	ctx.Set("csrf/token", token)
	logger.Debug("CSRF token generated for root path")
}

// validateCSRFToken checks the validity of the CSRF token for non-root, non-excluded routes.
func validateCSRFToken(ctx *gin.Context, config *csrf.Config, logger *log.Log) {
	token, err := csrf.NewTokenFromCookie(config, ctx)

	if errors.Is(err, csrf.ErrCookieRetrieval) {
		logger.Warn("CSRF cookie not found")
		_ = ctx.Error(problems.ProblemCSRFMissing)
		ctx.AbortWithStatus(http.StatusExpectationFailed)
		return
	}

	if err != nil {
		logger.Error("Invalid CSRF token", log.Err(err))
		_ = ctx.Error(problems.ProblemCSRFInvalid)
		ctx.AbortWithStatus(http.StatusExpectationFailed)
		return
	}

	// Optionally validate CSRF token from headers (commented out for flexibility)
	// if !token.Compare(ctx.GetHeader(config.HeaderName)) {
	// 	log.Warn("CSRF token mismatch")
	// 	_ = ctx.Error(problems.ProblemCSRFInvalid)
	// 	ctx.AbortWithStatus(http.StatusExpectationFailed)
	// 	return
	// }

	ctx.Set("csrf/token", token)
	logger.Debug("CSRF token validated for path:", log.Any("Full Path", ctx.FullPath()))
}
