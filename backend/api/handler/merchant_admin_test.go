package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// ① UpdateShop persists reward config, including 0 values (pointer fields).
func TestUpdateShop_PersistsRewardRates(t *testing.T) {
	setupTestDB(t)

	const merchantID = uint(901)
	shop := models.Shop{Name: "Reward Cfg Shop", MerchantID: merchantID, Status: 1,
		RewardRateSelf: 0.03, RewardCeiling: 0.50}
	config.DB.Create(&shop)

	r := setupRouter()
	setAuthContext(r, "PUT", "/api/merchant/shops/:id", UpdateShop, merchantID)

	body := map[string]interface{}{
		"reward_rate_self":   0.05,
		"reward_ceiling":     0, // 0 must persist (pointer field)
		"reward_rate_level1": 0.12,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", "/api/merchant/shops/"+itoa(shop.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var saved models.Shop
	config.DB.First(&saved, shop.ID)
	if saved.RewardRateSelf != 0.05 {
		t.Errorf("expected reward_rate_self 0.05, got %v", saved.RewardRateSelf)
	}
	if saved.RewardRateLevel1 != 0.12 {
		t.Errorf("expected reward_rate_level1 0.12, got %v", saved.RewardRateLevel1)
	}
	if saved.RewardCeiling != 0 {
		t.Errorf("expected reward_ceiling 0 (0 must persist), got %v", saved.RewardCeiling)
	}
}

// ① UpdateShop rejects a merchant editing a shop they don't own.
func TestUpdateShop_RejectsNonOwner(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Owned Shop", MerchantID: 910, Status: 1}
	config.DB.Create(&shop)

	r := setupRouter()
	setAuthContext(r, "PUT", "/api/merchant/shops/:id", UpdateShop, 911) // different merchant

	body, _ := json.Marshal(map[string]interface{}{"name": "Hijacked"})
	req, _ := http.NewRequest("PUT", "/api/merchant/shops/"+itoa(shop.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d", w.Code)
	}
}

// ② UpdateProduct can set status to 0 (下架) — previously blocked by `if status > 0`.
func TestUpdateProduct_CanSetStatusZero(t *testing.T) {
	setupTestDB(t)

	const merchantID = uint(902)
	shop := models.Shop{Name: "Product Status Shop", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 10, Status: 1}
	config.DB.Create(&product)

	r := setupRouter()
	setAuthContext(r, "PUT", "/api/merchant/products/:id", UpdateProduct, merchantID)

	body, _ := json.Marshal(map[string]interface{}{"status": 0})
	req, _ := http.NewRequest("PUT", "/api/merchant/products/"+itoa(product.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var saved models.Product
	config.DB.First(&saved, product.ID)
	if saved.Status != 0 {
		t.Errorf("expected status 0 (下架), got %d", saved.Status)
	}
}

// ② UpdateProduct rejects a merchant editing a product in a shop they don't own.
func TestUpdateProduct_RejectsNonOwner(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Owner Shop", MerchantID: 920, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 10, Status: 1}
	config.DB.Create(&product)

	r := setupRouter()
	setAuthContext(r, "PUT", "/api/merchant/products/:id", UpdateProduct, 921) // different merchant

	body, _ := json.Marshal(map[string]interface{}{"name": "Hijacked"})
	req, _ := http.NewRequest("PUT", "/api/merchant/products/"+itoa(product.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d", w.Code)
	}

	// And the name must be unchanged.
	var saved models.Product
	config.DB.First(&saved, product.ID)
	if saved.Name != "Dish" {
		t.Errorf("expected name unchanged, got %q", saved.Name)
	}
}

// ③ GenerateMerchantQR issues a QR for an owned shop and rejects others.
func TestGenerateMerchantQR_OwnershipEnforced(t *testing.T) {
	setupTestDB(t)
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{} // handler reads Server.BaseURL
	}

	const ownerID = uint(903)
	shop := models.Shop{Name: "QR Shop", MerchantID: ownerID, Status: 1}
	config.DB.Create(&shop)

	// Owner can generate
	r := setupRouter()
	setAuthContext(r, "POST", "/api/merchant/shops/:id/qrcodes", GenerateMerchantQR, ownerID)
	body, _ := json.Marshal(map[string]interface{}{"table_no": "A01"})
	req, _ := http.NewRequest("POST", "/api/merchant/shops/"+itoa(shop.ID)+"/qrcodes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for owner, got %d body: %s", w.Code, w.Body.String())
	}

	// Non-owner is rejected
	r2 := setupRouter()
	setAuthContext(r2, "POST", "/api/merchant/shops/:id/qrcodes", GenerateMerchantQR, 999)
	req2, _ := http.NewRequest("POST", "/api/merchant/shops/"+itoa(shop.ID)+"/qrcodes", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-owner, got %d", w2.Code)
	}
}

func itoa(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
