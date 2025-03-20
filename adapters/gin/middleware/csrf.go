package middleware

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-gonic/gin"
)

// CSRFToken represents a cross-site request forgery token
type CSRFToken struct {
	Value     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CSRFManager centrally manages CSRF tokens across requests
type CSRFManager struct {
	secretKey      []byte
	cookieName     string
	headerName     string
	path           string
	secureCookie   bool
	sameSite       http.SameSite
	tokenLifetime  time.Duration
	excludedRoutes []string

	// Map of session ID to token
	tokens     map[string]*CSRFToken
	tokenMutex sync.RWMutex
}

// NewCSRFManager creates a new CSRF manager
func NewCSRFManager(secretKey string, excludedRoutes []string) *CSRFManager {
	return &CSRFManager{
		secretKey:      []byte(secretKey),
		cookieName:     constant.CSRFTokenCookie,
		headerName:     constant.CSRFTokenHeader,
		path:           "/",
		secureCookie:   true,
		sameSite:       http.SameSiteStrictMode,
		tokenLifetime:  24 * time.Hour,
		excludedRoutes: excludedRoutes,
		tokens:         make(map[string]*CSRFToken),
	}
}

// CreateCSRFConfig initializes the CSRF configuration settings.
func CreateCSRFConfig(secretKey string, excludedRoutes []string) *CSRFManager {
	return NewCSRFManager(secretKey, excludedRoutes)
}

// CreateToken generates and stores a new token for the given session
func (m *CSRFManager) CreateToken(sessionID string) (*CSRFToken, error) {
	// Generate a unique token
	tokenData := fmt.Sprintf("%s:%d:%s", sessionID, time.Now().UnixNano(), m.secretKey)
	hasher := sha256.New()
	hasher.Write([]byte(tokenData))
	tokenValue := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	// Create token with expiration
	now := time.Now()
	token := &CSRFToken{
		Value:     tokenValue,
		CreatedAt: now,
		ExpiresAt: now.Add(m.tokenLifetime),
	}

	// Store token
	m.tokenMutex.Lock()
	m.tokens[sessionID] = token
	m.tokenMutex.Unlock()

	return token, nil
}

// GetToken retrieves a token for the given session
func (m *CSRFManager) GetToken(sessionID string) *CSRFToken {
	m.tokenMutex.RLock()
	token, exists := m.tokens[sessionID]
	m.tokenMutex.RUnlock()

	if !exists {
		return nil
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		m.tokenMutex.Lock()
		delete(m.tokens, sessionID)
		m.tokenMutex.Unlock()
		return nil
	}

	return token
}

// ValidateToken checks if the provided token matches the stored one
func (m *CSRFManager) ValidateToken(sessionID, tokenValue string) bool {
	token := m.GetToken(sessionID)
	if token == nil {
		return false
	}

	return token.Value == tokenValue
}

// SetCSRFCookie sets the CSRF token cookie
func (m *CSRFManager) SetCSRFCookie(w http.ResponseWriter, token *CSRFToken) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    token.Value,
		Path:     m.path,
		Secure:   m.secureCookie,
		HttpOnly: true,
		SameSite: m.sameSite,
		Expires:  token.ExpiresAt,
	})
}

// GetOrCreateSessionID gets the existing session ID or creates a new one
func (m *CSRFManager) GetOrCreateSessionID(r *http.Request, w http.ResponseWriter) (string, error) {
	// Try to get existing session ID
	sessionCookie, err := r.Cookie(constant.SessionID)
	if err == nil && sessionCookie.Value != "" {
		return sessionCookie.Value, nil
	}

	// Generate a new session ID
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constant.SessionID,
		Value:    sessionID,
		Path:     "/",
		Secure:   m.secureCookie,
		HttpOnly: true,
		SameSite: m.sameSite,
		Expires:  time.Now().Add(m.tokenLifetime),
	})

	return sessionID, nil
}

// HandleCSRF processes the CSRF token for a request
func (m *CSRFManager) HandleCSRF(w http.ResponseWriter, r *http.Request) (*CSRFToken, error) {
	// Get or create the session ID
	sessionID, err := m.GetOrCreateSessionID(r, w)
	if err != nil {
		return nil, errors.New("failed to get session ID")
	}

	// For root path requests, always generate a new token
	if r.URL.Path == "/" {
		token, err := m.CreateToken(sessionID)
		if err != nil {
			return nil, errors.New("failed to generate CSRF token")
		}

		// Set the token cookie
		m.SetCSRFCookie(w, token)
		return token, nil
	}

	// For excluded routes, skip CSRF validation
	for _, route := range m.excludedRoutes {
		if route == r.URL.Path {
			return nil, nil
		}
	}

	// For other paths, get the token but don't validate for GET requests
	token := m.GetToken(sessionID)
	if token == nil {
		// No token exists yet - this means the client hasn't visited "/" first
		return nil, errors.New("CSRF token not found - visit root path first")
	}

	// For non-GET requests, check the header token
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
		headerToken := r.Header.Get(m.headerName)
		if headerToken == "" {
			return nil, errors.New("CSRF token header missing")
		}

		if !m.ValidateToken(sessionID, headerToken) {
			return nil, errors.New("CSRF token invalid")
		}
	}

	return token, nil
}

// GetCSRFToken is a helper to get the CSRF token from the Gin context
func GetCSRFToken(c *gin.Context) string {
	value, exists := c.Get(constant.CSRFTokenHeader)
	if !exists {
		return ""
	}
	return value.(string)
}

// CSRFMiddleware is a middleware to handle CSRF protection
func CSRFMiddleware(csrfManager *CSRFManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if route is excluded from CSRF protection
		if isExcludedRoute(csrfManager, c.FullPath()) {
			c.Next()
			return
		}

		// Handle CSRF token logic
		token, err := csrfManager.HandleCSRF(c.Writer, c.Request)

		// Handle root path separately for token initialization
		if c.FullPath() == "/" {
			handleRootPath(c, token)
			c.Next()
			return
		}

		// Handle CSRF validation errors for non-GET requests
		if err != nil {
			handleCSRFError(c, err)
			return
		}

		// Store token in context if it exists
		if token != nil {
			c.Set(constant.CSRFTokenHeader, token.Value)
		}

		c.Next()
	}
}

// Helper function to check if the route is excluded from CSRF protection
func isExcludedRoute(csrfManager *CSRFManager, route string) bool {
	return slices.Contains(csrfManager.excludedRoutes, route)
}

// Handle special case for root path and store CSRF token in context
func handleRootPath(c *gin.Context, token *CSRFToken) {
	if token != nil {
		c.Set(constant.CSRFTokenHeader, token.Value)
	}
}

// Handle CSRF validation errors and send appropriate responses
func handleCSRFError(c *gin.Context, err error) {
	if err.Error() == "CSRF token not found - visit root path first" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Please visit the root path (/) first to initialize CSRF protection",
		})
		return
	}

	// For non-GET requests
	if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("CSRF validation failed: %s", err.Error()),
		})
	}
}

/*

// TODO change the logic so that everytime new csrf can be created initially and then used accross application
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

*/
