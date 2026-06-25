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

// dine_in is the default order type when none is supplied (regression).
func TestCreateOrder_DefaultsToDineIn(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "OrderType Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 30, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "ot_default", Nickname: "OTDefault", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   30,
		"items":    []map[string]interface{}{{"product_id": product.ID, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.OrderType != "dine_in" {
		t.Errorf("expected order_type dine_in, got %q", order.OrderType)
	}
}

// delivery orders may omit table_no and must persist order_type=delivery.
func TestCreateOrder_DeliveryAllowsEmptyTableNo(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Delivery Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 40, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "ot_delivery", Nickname: "OTDelivery", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":    shop.ID,
		"amount":     40,
		"order_type": "delivery",
		"items":      []map[string]interface{}{{"product_id": product.ID, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for delivery without table_no, got %d body: %s", w.Code, w.Body.String())
	}
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.OrderType != "delivery" {
		t.Errorf("expected order_type delivery, got %q", order.OrderType)
	}
	if order.TableNo != "" {
		t.Errorf("expected empty table_no for delivery, got %q", order.TableNo)
	}
}

// dine_in still requires a table_no.
func TestCreateOrder_DineInRequiresTableNo(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "DineIn Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 20, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "ot_dinein_req", Nickname: "OTDineInReq", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id": shop.ID,
		"amount":  20,
		"items":   []map[string]interface{}{{"product_id": product.ID, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for dine_in without table_no, got %d body: %s", w.Code, w.Body.String())
	}
}
