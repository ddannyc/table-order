package router

import "github.com/gin-gonic/gin"

// ConfigureTrustedProxies pins which upstream may set X-Forwarded-For. An empty
// list trusts no proxy, so gin's ClientIP() resolves the real TCP peer and a
// client-spoofed XFF is ignored (the rate limiter keys on a non-forgeable value).
// Returns an error for a malformed proxy entry so the caller can fail fast rather
// than silently fall back to trusting nothing.
func ConfigureTrustedProxies(r *gin.Engine, trustedProxies []string) error {
	return r.SetTrustedProxies(trustedProxies)
}
