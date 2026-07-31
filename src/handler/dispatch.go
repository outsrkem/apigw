package handler

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// Global constraint constants for backend request timeout limit
const (
	MinBackendTimeoutMs = 200   // Minimum guaranteed timeout threshold: 200 milliseconds
	MaxBackendTimeoutMs = 30000 // Maximum capped timeout threshold: 30 seconds
)

// GetSafeTimeout standardize raw timeout value to avoid extreme unreasonable configuration
// Clamp the input value between MinBackendTimeoutMs and MaxBackendTimeoutMs
func GetSafeTimeout(raw int) int {
	if raw < MinBackendTimeoutMs {
		return MinBackendTimeoutMs
	}
	if raw > MaxBackendTimeoutMs {
		return MaxBackendTimeoutMs
	}
	return raw
}

// RewritePath intercept url prefix to rewrite upstream request path, consistent with Nginx prefix truncation rules
// path: original request full path
// rawPath: route configured matching prefix path
// backendUri: fixed target backend base path
// mode: route match mode, EXACT or PREFIX
// return final rewritten request path for upstream access
func RewritePath(path, rawPath, backendUri, mode string) string {
	if rawPath == "" || path == "" {
		return backendUri
	}
	if mode == "EXACT" {
		return backendUri
	}
	if backendUri == "" {
		return path
	}
	// Cut off matched prefix and splice remaining suffix to backend base path
	suffix := strings.TrimPrefix(path, rawPath)
	raw := backendUri + suffix
	// Remove the consecutive slashes at the beginning
	for strings.HasPrefix(raw, "/") {
		raw = raw[1:]
	}
	return "/" + raw
}

// getRealClientIP get client public IP address with priority rule:
// X-Real-IP Header > X-Forwarded-For Header > direct connection remote address
func getRealClientIP(c *app.RequestContext) string {
	xRealIP := string(c.Request.Header.Get("X-Real-IP"))
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}

	xff := string(c.Request.Header.Get("X-Forwarded-For"))
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	return c.RemoteAddr().String()
}
