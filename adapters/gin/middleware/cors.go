package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// CORSMiddleware returns a gin.HandlerFunc that handles CORS requests
func CORSMiddleware(additionalHeaders ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		w := c.Writer

		// Set security headers
		setSecurityHeaders(w)

		// Check for preflight request
		if isCORSPreflightRequest(r) {
			if handlePreflightRequest(r, w, additionalHeaders...) {
				c.JSON(http.StatusOK, gin.H{"message": "OK"})
				return
			}
		}

		// Handle regular CORS request
		handleRegularCORSRequest(w)
		c.Next()
	}
}

// setSecurityHeaders sets security headers for the response
func setSecurityHeaders(w http.ResponseWriter) {
	// Set common security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
}

// isCORSPreflightRequest returns true if the request is a CORS preflight request
func isCORSPreflightRequest(r *http.Request) bool {
	// Determine if the request is a CORS preflight
	return r.Method == http.MethodOptions &&
		r.Header.Get("Origin") != "" &&
		r.Header.Get("Access-Control-Request-Method") != ""
}

// handlePreflightRequest handles preflight requests
func handlePreflightRequest(r *http.Request, w http.ResponseWriter, additionalHeaders ...string) bool {
	// Handle preflight request and allow methods
	method := r.Header.Get("Access-Control-Request-Method")
	if ok := isMethodAllowed(method); ok {
		origin := getAllowedOrigin(r)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", getAllowedMethods())
		headers := getAllowedHeaders()
		if len(additionalHeaders) > 0 {
			headers += ", " + strings.Join(additionalHeaders, ", ")
		}
		w.Header().Set("Access-Control-Allow-Headers", headers)
		w.Header().Set("Access-Control-Allow-Credentials", "true") // If you allow credentials
		w.Header().Add("Vary", "Origin")
		return true
	}
	return false
}

// handleRegularCORSRequest handles regular CORS requests
func handleRegularCORSRequest(w http.ResponseWriter) {
	// Regular CORS request processing (non-preflight)
	origin := getAllowedOrigin(nil) // Get allowed origin dynamically
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Add("Vary", "Origin")
}

// isMethodAllowed returns true if the method is allowed
func isMethodAllowed(method string) bool {
	// Check if the requested method is allowed
	allowedMethods := getAllowedMethodsList()
	for _, m := range allowedMethods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// getAllowedMethodsList returns a list of allowed HTTP methods for CORS
func getAllowedMethodsList() []string {
	// Return list of allowed HTTP methods for CORS
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodHead, // Additional standard methods
		"CONNECT",       // For completeness in HTTP
		"TRACE",         // Less commonly used
	}
}

// getAllowedMethods returns a comma-separated string of allowed HTTP methods for CORS
func getAllowedMethods() string {
	// Return allowed methods as a string
	return strings.Join(getAllowedMethodsList(), ", ")
}

// getAllowedHeaders returns a comma-separated string of allowed headers for CORS
func getAllowedHeaders() string {
	// Return allowed headers for CORS
	return `Content-Type,
	 Content-Length,
	 Accept-Encoding,
	 X-CSRF-Token,
	 Authorization,
	 accept,
	 origin,
	 Cache-Control,
	 X-Correlation-ID,
	 X-Requested-With,
	 X-Subject,
	 X-Signature,
	 X-Paseto-Token,
	 X-Refresh-Token,
	 X-User-Role`
}

// getAllowedOrigin returns the allowed origin for CORS with wildcard support
func getAllowedOrigin(r *http.Request) string {
	allowedOrigins := viper.GetStringSlice(constant.CorsAllowedOriginsKey)

	if len(allowedOrigins) == 0 || slices.Contains(allowedOrigins, "*") {
		return "*"
	}

	if r == nil {
		return "*"
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return "*"
	}

	for _, allowedOrigin := range allowedOrigins {
		// Exact match
		if allowedOrigin == origin {
			return origin
		}

		// Wildcard pattern like https://*.abhishek.com
		if strings.Contains(allowedOrigin, "*") {
			prefix := strings.Split(allowedOrigin, "*")[0]
			suffix := strings.Split(allowedOrigin, "*")[1]
			if strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
				return origin
			}
		}
	}

	return "*"
}
