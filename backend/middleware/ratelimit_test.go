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

// The sliding window resets: once enough time passes, a previously-throttled key
// is allowed again. Uses the injectable clock so there are no sleeps.
func TestRateLimiter_WindowResets(t *testing.T) {
	clock := mockTime{}
	rl := &rateLimiter{hits: map[string][]time.Time{}, limit: 2, window: time.Minute, now: clock.now}
	if !rl.allow("k") {
		t.Fatal("first hit should be allowed")
	}
	if !rl.allow("k") {
		t.Fatal("second hit should be allowed")
	}
	if rl.allow("k") {
		t.Fatal("third hit within window should be blocked")
	}
	clock.advance(time.Minute + time.Second)
	if !rl.allow("k") {
		t.Error("after the window elapses the key should be allowed again")
	}
}

// Abandoned keys (all hits expired) are evicted so the map can't grow without bound.
func TestRateLimiter_EvictsExpiredKeys(t *testing.T) {
	clock := mockTime{}
	rl := &rateLimiter{hits: map[string][]time.Time{}, limit: 5, window: time.Minute, now: clock.now}
	rl.allow("gone")
	clock.advance(2 * time.Minute) // "gone" fully expires
	rl.allow("fresh")              // a later request triggers the sweep
	if _, ok := rl.hits["gone"]; ok {
		t.Errorf("expired key should have been evicted, map=%v", rl.hits)
	}
}

func TestByUserID_KeysPerUserAndFallsBackToIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mk := func(setUser bool, uid uint) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.RemoteAddr = "5.5.5.5:1"
		if setUser {
			c.Set("user_id", uid)
		}
		return c
	}
	if k := ByUserID(mk(true, 7)); k != "u:7" {
		t.Errorf("expected u:7, got %q", k)
	}
	if ByUserID(mk(true, 7)) == ByUserID(mk(true, 8)) {
		t.Error("different users must get different keys")
	}
	if k := ByUserID(mk(false, 0)); k != "ip:5.5.5.5" {
		t.Errorf("missing user_id should fall back to IP, got %q", k)
	}
}

// mockTime is an injectable monotonic clock for the limiter tests.
type mockTime struct{ offset time.Duration }

func (m *mockTime) advance(d time.Duration) { m.offset += d }
func (m *mockTime) now() time.Time          { return time.Unix(0, 0).Add(m.offset) }

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
