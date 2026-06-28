package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimit(2, time.Minute, ByIP), func(c *gin.Context) { c.Status(http.StatusOK) })

	var codes []int
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/x", nil)
		req.RemoteAddr = "1.2.3.4:5678"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != 200 || codes[1] != 200 || codes[2] != http.StatusTooManyRequests {
		t.Errorf("expected [200 200 429], got %v", codes)
	}
}

// With no trusted proxies, a spoofed X-Forwarded-For must NOT change the limiter
// key — both requests from the same TCP peer share one bucket, so the second is
// throttled regardless of the forged header.
func TestRateLimit_IgnoresSpoofedXFF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}
	r.GET("/x", RateLimit(1, time.Minute, ByIP), func(c *gin.Context) { c.Status(http.StatusOK) })

	var codes []int
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", "/x", nil)
		req.RemoteAddr = "203.0.113.7:5555"
		req.Header.Set("X-Forwarded-For", "9.9.9."+itoa(i)) // attacker rotates this
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != 200 || codes[1] != http.StatusTooManyRequests {
		t.Errorf("spoofed XFF should not reset the limit; expected [200 429], got %v", codes)
	}
}

func itoa(i int) string { return strconv.Itoa(i) }

func TestRateLimit_SeparateKeysIndependent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimit(1, time.Minute, ByIP), func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, ip := range []string{"1.1.1.1:1", "2.2.2.2:2"} {
		req, _ := http.NewRequest("GET", "/x", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("first hit for %s should pass, got %d", ip, w.Code)
		}
	}
}
