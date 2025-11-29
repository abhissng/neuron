package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
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
	host := req.Host
	// detect request scheme (best-effort)
	scheme := "http"
	if req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	secure := false
	switch env {
	case "prod":
		secure = true
	case "staging":
		// staging often uses HTTPS; keep secure true if scheme shows https
		secure = scheme == "https"
	default:
		// dev: default to false unless TLS
		secure = scheme == "https"
	}

	// Decide domain: if override provided use it; otherwise, only set Domain when not localhost
	var domain string
	if domainOverride != "" {
		domain = domainOverride
	} else if helpers.IsLocalhostHost(host) {
		domain = "" // host-only cookie (do NOT set Domain for localhost)
	} else {
		domain = strings.Split(host, ":")[0]
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     constant.SessionID,
		Value:    sessionID,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode, // use None for cross-site requests; requires Secure=true in browsers
	})
}
