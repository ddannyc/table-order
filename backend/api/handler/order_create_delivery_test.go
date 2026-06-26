package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// validToken signs a quote token the way the quote endpoint would.
func validToken(shopID uint, fee, lat, lng float64) string {
	return signQuoteToken(quoteClaims{
		ShopID: shopID, Fee: fee, Lat: lat, Lng: lng,
		ShansongQuote: "SS-Q-XYZ", Exp: time.Now().Add(10 * time.Minute).Unix(),
	})
}

func postDeliveryOrder(t *testing.T, userID uint, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, userID)
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateOrder_DeliveryPersistsOrderDeliveryWithTrustedFee(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "Delivery Shop", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "deliv_user_1", Nickname: "D1", Role: 0, Balance: 0}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id":     shop.ID,
		"order_type":  "delivery",
		"amount":      100,
		"quote_token": validToken(shop.ID, 8.5, 39.91, 116.41),
		"items":       []map[string]any{{"product_id": product.ID, "quantity": 1}},
		"delivery": map[string]any{
			"recipient_name": "张三", "recipient_phone": "13800000000",
			"province": "北京市", "city": "北京市", "county": "朝阳区",
			"detail_address": "某路1号", "lat": 39.91, "lng": 116.41,
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.OrderType != "delivery" {
		t.Errorf("expected order_type delivery, got %q", order.OrderType)
	}
	// order.Amount is the reward base = items - reward (no reward here) = 100, NOT incl. fee.
	if order.Amount != 100 {
		t.Errorf("expected order.Amount 100 (reward base, no fee), got %v", order.Amount)
	}

	var od models.OrderDelivery
	if err := config.DB.Where("order_id = ?", order.ID).First(&od).Error; err != nil {
		t.Fatalf("expected OrderDelivery row: %v", err)
	}
	if od.DeliveryFee != 8.5 {
		t.Errorf("expected delivery_fee 8.5 (trusted from token), got %v", od.DeliveryFee)
	}
	if od.RecipientLat == 0 || od.RecipientName != "张三" {
		t.Errorf("expected recipient persisted, got %+v", od)
	}
	if od.ShansongQuoteNo != "SS-Q-XYZ" {
		t.Errorf("expected shansong quote no persisted, got %q", od.ShansongQuoteNo)
	}

	// Response should expose delivery_fee + pay_amount = 100 + 8.5
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if fee, _ := resp["delivery_fee"].(float64); fee != 8.5 {
		t.Errorf("expected resp delivery_fee 8.5, got %v", resp["delivery_fee"])
	}
	if pay, _ := resp["pay_amount"].(float64); pay != 108.5 {
		t.Errorf("expected resp pay_amount 108.5, got %v", resp["pay_amount"])
	}
}

func TestCreateOrder_DeliveryRejectsInvalidToken(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "Delivery Shop2", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "deliv_user_2", Nickname: "D2", Role: 0}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id": shop.ID, "order_type": "delivery", "amount": 100,
		"quote_token": "garbage.token",
		"items":       []map[string]any{{"product_id": product.ID, "quantity": 1}},
		"delivery":    map[string]any{"recipient_name": "X", "lat": 39.91, "lng": 116.41},
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on invalid token, got %d", w.Code)
	}
	var count int64
	config.DB.Model(&models.Order{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected no order created on bad token, got %d", count)
	}
}

func TestCreateOrder_DeliveryRejectsTokenShopMismatch(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "Delivery Shop3", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "deliv_user_3", Nickname: "D3", Role: 0}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id": shop.ID, "order_type": "delivery", "amount": 100,
		"quote_token": validToken(shop.ID+999, 8.5, 39.91, 116.41), // token for a different shop
		"items":       []map[string]any{{"product_id": product.ID, "quantity": 1}},
		"delivery":    map[string]any{"recipient_name": "X", "lat": 39.91, "lng": 116.41},
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on shop mismatch, got %d", w.Code)
	}
}

func TestCreateOrder_DeliveryFeeExcludedFromRewardBase(t *testing.T) {
	setupTestDB(t)
	// ceiling 0.5: reward can cover at most 50% of items (100 -> 50 deductible)
	shop := models.Shop{Name: "Delivery Shop4", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4, RewardCeiling: 0.5}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "deliv_user_4", Nickname: "D4", Role: 0, RewardBalance: 50}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id": shop.ID, "order_type": "delivery", "amount": 100, "use_reward": true,
		"quote_token": validToken(shop.ID, 8.5, 39.91, 116.41),
		"items":       []map[string]any{{"product_id": product.ID, "quantity": 1}},
		"delivery":    map[string]any{"recipient_name": "X", "lat": 39.91, "lng": 116.41},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	// items 100 - reward 50 = 50 reward base. Fee (8.5) must NOT be in order.Amount.
	if order.Amount != 50 {
		t.Errorf("expected order.Amount 50 (reward base, fee excluded), got %v", order.Amount)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if pay, _ := resp["pay_amount"].(float64); pay != 58.5 {
		t.Errorf("expected pay_amount 58.5 (50 + 8.5 fee), got %v", resp["pay_amount"])
	}
}

// Items fully covered by reward but a delivery fee remains -> must NOT take the
// zero-amount auto-paid path; it goes to WeChat Pay to collect the fee.
func TestCreateOrder_DeliveryWithFeeNotZeroAutoPaid(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "Delivery Shop5", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4, RewardCeiling: 1.0}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "deliv_user_5", Nickname: "D5", Role: 0, RewardBalance: 100}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id": shop.ID, "order_type": "delivery", "amount": 100, "use_reward": true,
		"quote_token": validToken(shop.ID, 8.5, 39.91, 116.41),
		"items":       []map[string]any{{"product_id": product.ID, "quantity": 1}},
		"delivery":    map[string]any{"recipient_name": "X", "lat": 39.91, "lng": 116.41},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.Status != 1 {
		t.Errorf("expected status 1 (pending, must pay the fee), got %d", order.Status)
	}
	if order.PaidAt != nil {
		t.Errorf("expected not auto-paid when a delivery fee is owed")
	}
}

// A delivery order_type with no delivery block / token is rejected.
func TestCreateOrder_DeliveryRequiresDeliveryBlock(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "Delivery Shop6", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "deliv_user_6", Nickname: "D6", Role: 0}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id": shop.ID, "order_type": "delivery", "amount": 100,
		"items": []map[string]any{{"product_id": product.ID, "quantity": 1}},
		// no delivery, no quote_token
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when delivery block/token missing, got %d", w.Code)
	}
}

// Regression: a dine-in order creates no OrderDelivery row.
func TestCreateOrder_DineInCreatesNoDeliveryRow(t *testing.T) {
	setupTestDB(t)
	shop := models.Shop{Name: "Dinein Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "Dish", Price: 100, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "dinein_user_1", Nickname: "DI1", Role: 0}
	config.DB.Create(&user)

	w := postDeliveryOrder(t, user.ID, map[string]any{
		"shop_id": shop.ID, "table_no": "A01", "amount": 100,
		"items": []map[string]any{{"product_id": product.ID, "quantity": 1}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	var count int64
	config.DB.Model(&models.OrderDelivery{}).Where("order_id = ?", order.ID).Count(&count)
	if count != 0 {
		t.Errorf("dine-in order must have no OrderDelivery row, got %d", count)
	}
}
