package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter is a simple in-memory sliding-window counter keyed by an arbitrary
// string (IP or user id). Adequate for a single instance; a multi-instance
// deployment would move this to a shared store (e.g. Redis).
type rateLimiter struct {
	mu        sync.Mutex
	hits      map[string][]time.Time
	limit     int
	window    time.Duration
	now       func() time.Time
	lastSweep time.Time
}

// sweep evicts keys whose newest hit is older than the window, so abandoned keys
// don't grow the map without bound. Throttled to once per window. Caller holds mu.
func (rl *rateLimiter) sweep(now time.Time) {
	if now.Sub(rl.lastSweep) < rl.window {
		return
	}
	rl.lastSweep = now
	cutoff := now.Add(-rl.window)
	for k, ts := range rl.hits {
		if len(ts) == 0 || !ts[len(ts)-1].After(cutoff) {
			delete(rl.hits, k)
		}
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	rl.sweep(now)
	cutoff := now.Add(-rl.window)
	kept := rl.hits[key][:0]
	for _, t := range rl.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.limit {
		rl.hits[key] = kept
		return false
	}
	rl.hits[key] = append(kept, now)
	return true
}

// RateLimit returns middleware that allows at most `limit` requests per `window`
// per key (derived by keyFn), responding 429 when exceeded.
func RateLimit(limit int, window time.Duration, keyFn func(*gin.Context) string) gin.HandlerFunc {
	rl := &rateLimiter{hits: map[string][]time.Time{}, limit: limit, window: window, now: time.Now}
	return func(c *gin.Context) {
		if !rl.allow(keyFn(c)) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ByIP keys the limiter on the client IP (for unauthenticated endpoints).
func ByIP(c *gin.Context) string { return "ip:" + c.ClientIP() }

// ByUserID keys on the authenticated user, falling back to IP.
func ByUserID(c *gin.Context) string {
	if id, ok := c.Get("user_id"); ok {
		if uid, ok := id.(uint); ok {
			return "u:" + strconv.FormatUint(uint64(uid), 10)
		}
	}
	return "ip:" + c.ClientIP()
}
