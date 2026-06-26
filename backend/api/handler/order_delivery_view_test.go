package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

func TestGetOrders_IncludesDeliveryInfo(t *testing.T) {
	setupTestDB(t)
	user := models.User{OpenID: "ov_user_1", Nickname: "OV1", Role: 0}
	config.DB.Create(&user)
	order := models.Order{OrderNo: "OV1", UserID: user.ID, ShopID: 1, OrderType: "delivery", Amount: 100, Status: 2}
	config.DB.Create(&order)
	config.DB.Create(&models.OrderDelivery{
		OrderID: order.ID, DeliveryFee: 8.5, ShansongStatus: 90,
		RecipientName: "张三", RecipientPhone: "13800000000",
		Province: "北京市", City: "北京市", County: "朝阳区", DetailAddress: "某路1号",
	})

	r := setupRouter()
	setAuthContext(r, "GET", "/api/orders", GetOrders, user.ID)
	req, _ := http.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	orders := resp["orders"].([]any)
	if len(orders) == 0 {
		t.Fatalf("expected at least one order")
	}
	first := orders[0].(map[string]any)
	d, ok := first["delivery"].(map[string]any)
	if !ok {
		t.Fatalf("expected delivery block in delivery order, got %v", first["delivery"])
	}
	if fee, _ := d["delivery_fee"].(float64); fee != 8.5 {
		t.Errorf("expected delivery_fee 8.5, got %v", d["delivery_fee"])
	}
	if d["status_label"] != "配送中" {
		t.Errorf("expected status_label 配送中 for status 90, got %v", d["status_label"])
	}
}

func TestGetOrders_DineInOmitsDelivery(t *testing.T) {
	setupTestDB(t)
	user := models.User{OpenID: "ov_user_2", Nickname: "OV2", Role: 0}
	config.DB.Create(&user)
	order := models.Order{OrderNo: "OV2", UserID: user.ID, ShopID: 1, OrderType: "dine_in", TableNo: "A01", Amount: 50, Status: 2}
	config.DB.Create(&order)

	r := setupRouter()
	setAuthContext(r, "GET", "/api/orders", GetOrders, user.ID)
	req, _ := http.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	orders := resp["orders"].([]any)
	first := orders[0].(map[string]any)
	if _, present := first["delivery"]; present {
		t.Errorf("dine_in order must omit delivery block, got %v", first["delivery"])
	}
}

func TestGetOrder_IncludesDelivery(t *testing.T) {
	setupTestDB(t)
	user := models.User{OpenID: "ov_user_3", Nickname: "OV3", Role: 0}
	config.DB.Create(&user)
	order := models.Order{OrderNo: "OV3", UserID: user.ID, ShopID: 1, OrderType: "delivery", Amount: 100, Status: 2}
	config.DB.Create(&order)
	config.DB.Create(&models.OrderDelivery{OrderID: order.ID, DeliveryFee: 6.0, ShansongStatus: 70})

	r := setupRouter()
	setAuthContext(r, "GET", "/api/orders/:id", GetOrder, user.ID)
	req, _ := http.NewRequest("GET", "/api/orders/"+itoa(order.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["delivery"].(map[string]any); !ok {
		t.Errorf("expected delivery block in GetOrder response, got %v", resp["delivery"])
	}
}
