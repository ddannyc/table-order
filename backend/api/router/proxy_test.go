package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// clientIP spins up an engine configured with the given trusted proxies and
// returns what ClientIP() resolves for a request from peer with the given XFF.
func clientIP(t *testing.T, trusted []string, peer, xff string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := ConfigureTrustedProxies(r, trusted); err != nil {
		t.Fatalf("ConfigureTrustedProxies: %v", err)
	}
	var got string
	r.GET("/ip", func(c *gin.Context) { got = c.ClientIP(); c.Status(200) })

	req, _ := http.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = peer
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	r.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

// With no trusted proxies, a spoofed XFF is ignored — ClientIP is the real peer.
func TestConfigureTrustedProxies_EmptyIgnoresSpoofedXFF(t *testing.T) {
	if ip := clientIP(t, nil, "203.0.113.9:5555", "1.2.3.4"); ip != "203.0.113.9" {
		t.Errorf("trust-none should resolve the TCP peer, got %q", ip)
	}
}

// When the peer is a trusted proxy, its forwarded client IP is honored.
func TestConfigureTrustedProxies_TrustsConfiguredEdge(t *testing.T) {
	if ip := clientIP(t, []string{"192.0.2.1/32"}, "192.0.2.1:5555", "1.2.3.4"); ip != "1.2.3.4" {
		t.Errorf("trusted edge should forward the client IP, got %q", ip)
	}
}

// A malformed proxy entry surfaces an error (caller fails fast) rather than
// silently leaving the engine trusting nothing.
func TestConfigureTrustedProxies_RejectsBadCIDR(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := ConfigureTrustedProxies(gin.New(), []string{"not-a-cidr"}); err == nil {
		t.Error("expected an error for a malformed trusted-proxy entry")
	}
}
