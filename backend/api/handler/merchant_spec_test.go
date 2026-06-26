package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

func TestCreateProductSpec_CreatesForOwnedProduct(t *testing.T) {
	setupTestDB(t)
	merchantID := uint(1)
	shop := models.Shop{MerchantID: merchantID, Name: "S", Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "酸奶青提", Price: 15, Status: 1}
	config.DB.Create(&product)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/merchant/products/:id/specs", CreateProductSpec, merchantID)
	body := map[string]interface{}{"name": "600ml", "price": 15, "status": 1}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/merchant/products/"+itoa(product.ID)+"/specs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var spec models.ProductSpec
	if err := config.DB.Where("product_id = ?", product.ID).First(&spec).Error; err != nil {
		t.Fatalf("expected spec persisted: %v", err)
	}
	if spec.Name != "600ml" || spec.Price != 15 {
		t.Errorf("unexpected spec: %+v", spec)
	}
}

// A merchant can create a spec that starts out 下架 (status 0); omitting status
// still defaults to 上架 (status 1).
func TestCreateProductSpec_AllowsExplicitInactive(t *testing.T) {
	setupTestDB(t)
	merchantID := uint(1)
	shop := models.Shop{MerchantID: merchantID, Name: "S", Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "酸奶青提", Price: 15, Status: 1}
	config.DB.Create(&product)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/merchant/products/:id/specs", CreateProductSpec, merchantID)

	// Explicit status:0 must persist as 下架.
	body := map[string]interface{}{"name": "限量大杯", "price": 25, "status": 0}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/merchant/products/"+itoa(product.ID)+"/specs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var inactive models.ProductSpec
	config.DB.Where("product_id = ? AND name = ?", product.ID, "限量大杯").First(&inactive)
	if inactive.Status != 0 {
		t.Errorf("expected status 0 (下架), got %d", inactive.Status)
	}

	// Omitting status defaults to 上架 (1).
	body2 := map[string]interface{}{"name": "常规中杯", "price": 18}
	jsonBody2, _ := json.Marshal(body2)
	req2, _ := http.NewRequest("POST", "/api/merchant/products/"+itoa(product.ID)+"/specs", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w2.Code, w2.Body.String())
	}
	var active models.ProductSpec
	config.DB.Where("product_id = ? AND name = ?", product.ID, "常规中杯").First(&active)
	if active.Status != 1 {
		t.Errorf("expected default status 1 (上架), got %d", active.Status)
	}
}

func TestCreateProductSpec_RejectsNonOwner(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{MerchantID: 1, Name: "S", Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "x", Price: 15, Status: 1}
	config.DB.Create(&product)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/merchant/products/:id/specs", CreateProductSpec, 999) // not owner
	body := map[string]interface{}{"name": "600ml", "price": 15}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/merchant/products/"+itoa(product.ID)+"/specs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("expected non-owner to be rejected, got 200")
	}
}

func TestUpdateProductSpec_UpdatesOwned(t *testing.T) {
	setupTestDB(t)
	merchantID := uint(1)
	shop := models.Shop{MerchantID: merchantID, Name: "S", Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "x", Price: 15, Status: 1}
	config.DB.Create(&product)
	spec := models.ProductSpec{ProductID: product.ID, Name: "600ml", Price: 15, Status: 1}
	config.DB.Create(&spec)

	r := setupRouter()
	setAuthContext(r, "PUT", "/api/merchant/specs/:id", UpdateProductSpec, merchantID)
	body := map[string]interface{}{"price": 18}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", "/api/merchant/specs/"+itoa(spec.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	config.DB.First(&spec, spec.ID)
	if spec.Price != 18 {
		t.Errorf("expected price 18 after update, got %v", spec.Price)
	}
}

func TestDeleteProductSpec_DeletesOwned(t *testing.T) {
	setupTestDB(t)
	merchantID := uint(1)
	shop := models.Shop{MerchantID: merchantID, Name: "S", Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "x", Price: 15, Status: 1}
	config.DB.Create(&product)
	spec := models.ProductSpec{ProductID: product.ID, Name: "600ml", Price: 15, Status: 1}
	config.DB.Create(&spec)

	r := setupRouter()
	setAuthContext(r, "DELETE", "/api/merchant/specs/:id", DeleteProductSpec, merchantID)
	req, _ := http.NewRequest("DELETE", "/api/merchant/specs/"+itoa(spec.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var count int64
	config.DB.Model(&models.ProductSpec{}).Where("id = ?", spec.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected spec deleted, still present")
	}
}
