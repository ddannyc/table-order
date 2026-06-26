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
// (Delivery now requires a delivery block + signed quote token — T4.)
func TestCreateOrder_DeliveryAllowsEmptyTableNo(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Delivery Shop", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 40, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "ot_delivery", Nickname: "OTDelivery", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":     shop.ID,
		"amount":      40,
		"order_type":  "delivery",
		"quote_token": validToken(shop.ID, 5.0, 39.91, 116.41),
		"items":       []map[string]interface{}{{"product_id": product.ID, "quantity": 1}},
		"delivery": map[string]interface{}{
			"recipient_name": "张三", "recipient_phone": "13800000000",
			"detail_address": "某路1号", "lat": 39.91, "lng": 116.41,
		},
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

// An unknown order_type is rejected (only dine_in / delivery are allowed).
func TestCreateOrder_RejectsUnknownOrderType(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "BadType Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 20, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "ot_bad", Nickname: "OTBad", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":    shop.ID,
		"table_no":   "A01",
		"amount":     20,
		"order_type": "takeaway", // not allowlisted
		"items":      []map[string]interface{}{{"product_id": product.ID, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown order_type, got %d body: %s", w.Code, w.Body.String())
	}
}

// An order with no line items is rejected — the server must price from items,
// never trust a bare client-supplied amount.
func TestCreateOrder_RejectsEmptyItems(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "NoItems Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	user := models.User{OpenID: "ot_noitems", Nickname: "OTNoItems", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   999, // arbitrary client amount must not be trusted
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty items, got %d body: %s", w.Code, w.Body.String())
	}
	var count int64
	config.DB.Model(&models.Order{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected no order created, got %d", count)
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
