package middleware

import (
	"net/http"
	"net/http/httptest"
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
