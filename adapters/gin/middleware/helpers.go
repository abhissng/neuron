package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/timeutil"
	"github.com/gin-gonic/gin"
)

// GetServiceContext retrieves the ServiceContext from the gin.Context.
// It returns an error if the ServiceContext is not found or is of the wrong type.
func GetServiceContext(c *gin.Context) (*context.ServiceContext, error) {
	sc, exists := c.Get(constant.ServiceContext)
	if !exists {
		return nil, errors.New("ServiceContext not found in gin.Context")
	}

	serviceCtx, ok := sc.(*context.ServiceContext)
	if !ok {
		return nil, errors.New("invalid type for ServiceContext in gin.Context")
	}

	return serviceCtx, nil
}

// SetSessionCookie writes a cookie using the request host/scheme defaults.
// - env: "prod", "staging", "dev", ...
// - domainOverride: optional; pass empty string to auto-detect
func SetSessionCookie(c *gin.Context, sessionID, env, domainOverride string, ttl time.Duration) {
	req := c.Request

	// 1. Determine the Origin Host (Where the user is)
	origin := req.Header.Get("Origin")
	originHost := helpers.HostFromOrigin(origin)

	// 2. Determine the Server Host (Where the API is)
	host := strings.Split(req.Host, ":")[0]

	// ... (Keep your Scheme/Secure detection logic here) ...
	scheme := "http"
	if req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}

	secure := false
	switch strings.ToLower(env) {
	case "prod":
		secure = true
	case "staging":
		secure = scheme == "https"
	default:
		secure = scheme == "https"
	}

	// 3. FIXED DOMAIN LOGIC
	var domain string
	if domainOverride != "" {
		domain = domainOverride
	} else {
		// If the ORIGIN is localhost (Developer Mode), clear the domain.
		// This creates a "Host-Only" cookie for staging.kitchenao.com,
		// which reduces friction for cross-site browser rules.
		if strings.Contains(originHost, "localhost") || originHost == "127.0.0.1" || originHost == "::1" {
			domain = ""
		} else {
			// Otherwise, set it to the server host (standard behavior)
			domain = host
		}
	}

	// 4. SameSite Logic (Keep existing, but ensure localhost gets None + Secure)
	// Treat differing origin host as cross-site
	isCrossSite := originHost != "" && host != "" && originHost != host

	sameSite := http.SameSiteLaxMode
	if isCrossSite && secure {
		// Essential for localhost -> staging communication
		sameSite = http.SameSiteNoneMode
	}

	cookie := &http.Cookie{
		Name:     constant.SessionID,
		Value:    sessionID,
		Path:     "/",
		Domain:   domain, // Now empty if origin is localhost
		Expires:  timeutil.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		Secure:   secure, // Must be TRUE for SameSite=None
		HttpOnly: true,
		SameSite: sameSite,
	}

	http.SetCookie(c.Writer, cookie)
}
