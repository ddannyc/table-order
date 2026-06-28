package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/example/table-order/config"
)

func authTestToken(method jwt.SigningMethod, secret string) string {
	tok := jwt.NewWithClaims(method, &Claims{UserID: 1, Role: 1})
	s, _ := tok.SignedString([]byte(secret))
	return s
}

func runAuth(token string) int {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req, _ := http.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestAuthMiddleware_AcceptsHS256(t *testing.T) {
	config.AppConfig = &config.Config{}
	config.AppConfig.JWT.Secret = "test-secret"
	if code := runAuth(authTestToken(jwt.SigningMethodHS256, "test-secret")); code != http.StatusOK {
		t.Fatalf("expected 200 for valid HS256 token, got %d", code)
	}
}

// Defense-in-depth: a token signed with a different (still HMAC) algorithm using
// the same secret must be rejected once the verifier pins HS256.
func TestAuthMiddleware_RejectsNonHS256Alg(t *testing.T) {
	config.AppConfig = &config.Config{}
	config.AppConfig.JWT.Secret = "test-secret"
	if code := runAuth(authTestToken(jwt.SigningMethodHS512, "test-secret")); code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for HS512 token (alg not pinned), got %d", code)
	}
}
