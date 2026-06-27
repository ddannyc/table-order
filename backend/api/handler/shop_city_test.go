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

// The merchant must be able to set the shop's city (needed for Shansong
// orderCalculate), and the public DTO should expose it.
func TestUpdateShop_PersistsCity(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "City Shop", MerchantID: 7, Status: 1}
	config.DB.Create(&shop)

	r := setupRouter()
	setAuthContext(r, "PUT", "/api/merchant/shops/:id", UpdateShop, 7)
	body, _ := json.Marshal(map[string]any{"city": "北京市"})
	req, _ := http.NewRequest("PUT", "/api/merchant/shops/"+itoa(shop.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	config.DB.First(&shop, shop.ID)
	if shop.City != "北京市" {
		t.Errorf("expected city persisted, got %q", shop.City)
	}
}

// A shop must be createable with its city and coordinates in one step, so a new
// delivery shop isn't born with city="" / coords=0 (the "门店未配置坐标/城市" gates).
func TestCreateShop_PersistsCityAndCoords(t *testing.T) {
	setupTestDB(t)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/merchant/shops", CreateShop, 9)
	body, _ := json.Marshal(map[string]any{
		"name":      "New Delivery Shop",
		"address":   "世纪广场1楼111号",
		"city":      "上海市",
		"latitude":  31.230416,
		"longitude": 121.473701,
	})
	req, _ := http.NewRequest("POST", "/api/merchant/shops", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var shop models.Shop
	if err := config.DB.Where("merchant_id = ?", 9).First(&shop).Error; err != nil {
		t.Fatalf("created shop not found: %v", err)
	}
	if shop.City != "上海市" {
		t.Errorf("expected city persisted, got %q", shop.City)
	}
	if shop.Latitude == 0 || shop.Longitude == 0 {
		t.Errorf("expected coords persisted, got lat=%v lng=%v", shop.Latitude, shop.Longitude)
	}
}

func TestGetShop_DTOIncludesCity(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "City DTO Shop", MerchantID: 1, Status: 1, City: "上海市"}
	config.DB.Create(&shop)

	r := setupRouter()
	r.GET("/api/shops/:id", GetShop)
	req, _ := http.NewRequest("GET", "/api/shops/"+itoa(shop.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["city"] != "上海市" {
		t.Errorf("expected city in public DTO, got %v", resp["city"])
	}
}
