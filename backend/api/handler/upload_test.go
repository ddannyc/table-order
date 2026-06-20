package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
)

// TestUploadImage_NotConfigured verifies the handler returns 503 when R2 is not
// configured. Needs no database or R2 — runs anywhere.
func TestUploadImage_NotConfigured(t *testing.T) {
	config.R2Client = nil // ensure unconfigured

	r := setupRouter()
	r.POST("/api/merchant/upload", UploadImage)

	req, _ := http.NewRequest("POST", "/api/merchant/upload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when R2 unconfigured, got %d body: %s", w.Code, w.Body.String())
	}
}
