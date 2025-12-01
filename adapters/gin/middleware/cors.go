package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
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

		// Handle regular CORS request (pass request so we can echo origin when appropriate)
		handleRegularCORSRequest(r, w)
		c.Next()
		// === Re-assert essential CORS headers on the final response ===
		// This prevents a downstream handler or proxy from accidentally removing them.
		origin := getAllowedOrigin(r)
		if origin != "" {
			h := w.Header()
			if h.Get("Access-Control-Allow-Origin") == "" {
				h.Set("Access-Control-Allow-Origin", origin)
			}
			if h.Get("Access-Control-Allow-Credentials") == "" {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if h.Get("Access-Control-Expose-Headers") == "" {
				h.Set("Access-Control-Expose-Headers", "Set-Cookie")
			}
			// also ensure Allow-Headers/Methods present on final response if absent
			if h.Get("Access-Control-Allow-Headers") == "" {
				h.Set("Access-Control-Allow-Headers", getAllowedHeaders(additionalHeaders...))
			}
			if h.Get("Access-Control-Allow-Methods") == "" {
				h.Set("Access-Control-Allow-Methods", getAllowedMethods())
			}
			h.Add("Vary", "Origin")
		}
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
		// If origin is not allowed, return false so caller can respond with default (403/empty)
		if origin == "" {
			return false
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", getAllowedMethods())
		w.Header().Set("Access-Control-Allow-Headers", getAllowedHeaders(additionalHeaders...))
		// Always allow credentials for allowed origins (caller intends to set credentials).
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Vary", "Origin")
		return true
	}
	return false
}

// handleRegularCORSRequest handles regular CORS requests
func handleRegularCORSRequest(r *http.Request, w http.ResponseWriter, additionalHeaders ...string) {
	origin := getAllowedOrigin(r)
	if origin == "" {
		return
	}

	// Required on actual request
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Vary", "Origin")

	// Additional recommended headers
	w.Header().Set("Access-Control-Allow-Headers", getAllowedHeaders(additionalHeaders...))
	w.Header().Set("Access-Control-Allow-Methods", getAllowedMethods())
	w.Header().Set("Access-Control-Expose-Headers", "Set-Cookie")
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

// getAllowedHeaders returns a comma-separated string of allowed headers for CORS,
// appending any optional additional headers provided.
func getAllowedHeaders(additionalHeaders ...string) string {
	// Define the base list of allowed headers
	headers := []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"Accept",
		"Origin",
		"Cache-Control",
		"X-Correlation-ID",
		"X-Requested-With",
		"X-Subject",
		"X-Signature",
		"X-Paseto-Token",
		"X-Refresh-Token",
		"X-User-Role",
		"X-Org-Id",
		"X-User-Id",
		"X-Feature-Flags",
		"X-Location-Id",
		"X-RateLimit-Limit",
		"Retry-After",
	}

	// If additional headers are provided, append them to the list
	if len(additionalHeaders) > 0 {
		headers = append(headers, additionalHeaders...)
	}

	// Return allowed headers as a string (joined by comma)
	return strings.Join(headers, ", ")
}

// getAllowedOrigin returns the allowed origin for CORS with wildcard support.
// IMPORTANT:
//   - If allowedOrigins contains "*" and a request Origin header is present, this function will echo the request Origin.
//     This avoids returning Access-Control-Allow-Origin: * which is incompatible with Allow-Credentials: true.
//   - If the request Origin is not present or not allowed, returns an empty string.
func getAllowedOrigin(r *http.Request) string {
	allowedOrigins := viper.GetStringSlice(constant.CorsAllowedOriginsKey)
	helpers.Println(constant.DEBUG, "allowedOrigins ", allowedOrigins)

	// If no request available, return a safe default:
	if r == nil {
		// If allowedOrigins explicitly configured, return first configured origin (or "*")
		if len(allowedOrigins) == 0 {
			return "*"
		}
		// If there is a configured wildcard, return "*" (no request to echo)
		if slices.Contains(allowedOrigins, "*") {
			return "*"
		}
		// Otherwise return first configured origin as a fallback
		return allowedOrigins[0]
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// No Origin header on request — nothing to do
		return ""
	}

	// If allowedOrigins empty -> act permissive but echo origin (to support credentials)
	if len(allowedOrigins) == 0 {
		return origin
	}

	// If wildcard present in allowedOrigins, echo the origin (don't return "*")
	for _, ao := range allowedOrigins {
		if ao == "*" {
			return origin
		}
	}

	// Try exact matches
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == origin {
			return origin
		}
	}

	// Try wildcard patterns like https://*.example.com
	for _, allowedOrigin := range allowedOrigins {
		if strings.Contains(allowedOrigin, "*") {
			parts := strings.SplitN(allowedOrigin, "*", 2)
			prefix := parts[0]
			suffix := ""
			if len(parts) > 1 {
				suffix = parts[1]
			}
			if strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
				return origin
			}
		}
	}

	// Not allowed
	return ""
}
