package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// The public GET /shops/:id must not leak the merchant's commission economics
// (reward_rate_*), but must still expose reward_ceiling (the client needs it to
// compute the on-screen discount cap) and the display fields.
func TestGetShop_PublicDTOHidesCommissionRates(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{
		Name: "Public Shop", MerchantID: 1, Status: 1,
		RewardRateSelf: 0.03, RewardRateLevel1: 0.10, RewardRateLevel2: 0.04,
		RewardCeiling: 0.6,
	}
	config.DB.Create(&shop)

	r := setupRouter()
	r.GET("/api/shops/:id", GetShop)
	req, _ := http.NewRequest("GET", "/api/shops/"+itoa(shop.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	for _, leaked := range []string{"reward_rate_self", "reward_rate_level1", "reward_rate_level2", "reward_exclude_categories"} {
		if _, present := resp[leaked]; present {
			t.Errorf("public shop DTO leaks %q", leaked)
		}
	}
	if _, ok := resp["reward_ceiling"]; !ok {
		t.Error("public shop DTO should expose reward_ceiling")
	}
	if resp["name"] != "Public Shop" {
		t.Errorf("expected name in DTO, got %v", resp["name"])
	}
}

// The public delivery-shop resolver must not leak commission rates either —
// /delivery/shop is an unauthenticated route.
func TestResolveDeliveryShop_PublicDTOHidesCommissionRates(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{
		Name: "Delivery DTO Shop", MerchantID: 1, Status: 1,
		RewardRateSelf: 0.03, RewardRateLevel1: 0.10, RewardRateLevel2: 0.04,
		RewardCeiling: 0.5,
	}
	config.DB.Create(&shop)
	config.DB.Create(&models.Product{ShopID: shop.ID, Name: "Dish", Price: 20, Status: 1})

	r := setupRouter()
	r.GET("/api/delivery/shop", ResolveDeliveryShop)
	req, _ := http.NewRequest("GET", "/api/delivery/shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	for _, leaked := range []string{"reward_rate_self", "reward_rate_level1", "reward_rate_level2", "reward_exclude_categories"} {
		if _, present := resp[leaked]; present {
			t.Errorf("delivery shop DTO leaks %q", leaked)
		}
	}
	if _, ok := resp["reward_ceiling"]; !ok {
		t.Error("delivery shop DTO should expose reward_ceiling")
	}
}

// The fallback branch (an active shop with NO active products) must also return
// the sanitized DTO — it's an unauthenticated route, so a re-leak here would be
// just as bad as the primary branch.
func TestResolveDeliveryShop_FallbackShopAlsoHidesCommissionRates(t *testing.T) {
	setupTestDB(t)

	// Active shop, but no active products → forces the fallback branch.
	shop := models.Shop{
		Name: "Fallback Shop", MerchantID: 1, Status: 1,
		RewardRateSelf: 0.03, RewardRateLevel1: 0.10, RewardRateLevel2: 0.04,
		RewardCeiling: 0.5,
	}
	config.DB.Create(&shop)

	r := setupRouter()
	r.GET("/api/delivery/shop", ResolveDeliveryShop)
	req, _ := http.NewRequest("GET", "/api/delivery/shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from fallback branch, got %d body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	for _, leaked := range []string{"reward_rate_self", "reward_rate_level1", "reward_rate_level2", "reward_exclude_categories"} {
		if _, present := resp[leaked]; present {
			t.Errorf("fallback delivery shop DTO leaks %q", leaked)
		}
	}
	if resp["name"] != "Fallback Shop" {
		t.Errorf("expected fallback shop in response, got %v", resp["name"])
	}
}
