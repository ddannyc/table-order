package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// Single-shop reality: resolving a delivery shop returns the active shop.
// Forward stub for nearest-shop ranking when multiple shops exist.
func TestResolveDeliveryShop_ReturnsActiveShop(t *testing.T) {
	setupTestDB(t)

	inactive := models.Shop{Name: "closed", Status: 1}
	config.DB.Create(&inactive)
	config.DB.Model(&inactive).Update("status", 0) // force 0 (default tag overrides zero-value on create)
	shop := models.Shop{Name: "世纪广场店", Status: 1, Latitude: 31.23, Longitude: 121.47}
	config.DB.Create(&shop)

	r := setupRouter()
	r.GET("/api/delivery/shop", ResolveDeliveryShop)
	req, _ := http.NewRequest("GET", "/api/delivery/shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if id, _ := resp["id"].(float64); uint(id) != shop.ID {
		t.Errorf("expected active shop id %d, got %v", shop.ID, resp["id"])
	}
	if _, ok := resp["latitude"]; !ok {
		t.Errorf("expected latitude field in resolved shop")
	}
}

func TestResolveDeliveryShop_NoneActive(t *testing.T) {
	setupTestDB(t)
	closed := models.Shop{Name: "closed", Status: 1}
	config.DB.Create(&closed)
	config.DB.Model(&closed).Update("status", 0) // force 0 (default tag overrides zero-value on create)

	r := setupRouter()
	r.GET("/api/delivery/shop", ResolveDeliveryShop)
	req, _ := http.NewRequest("GET", "/api/delivery/shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when no active shop, got %d", w.Code)
	}
}
