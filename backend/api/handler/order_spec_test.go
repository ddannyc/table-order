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

// An order line referencing a spec is priced at the spec's price and records
// the spec on the order item.
func TestCreateOrder_UsesSpecPrice(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Spec Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "酸奶青提", Price: 15, Status: 1}
	config.DB.Create(&product)
	spec := models.ProductSpec{ProductID: product.ID, Name: "800ml", Price: 18, Status: 1}
	config.DB.Create(&spec)
	user := models.User{OpenID: "spec_price_user", Nickname: "SpecPrice", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   1, // server recomputes
		"items":    []map[string]interface{}{{"product_id": product.ID, "spec_id": spec.ID, "quantity": 1}},
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
	if order.Amount != 18 {
		t.Errorf("expected amount 18 (spec price), got %v", order.Amount)
	}
	var item models.OrderItem
	config.DB.Where("order_id = ?", order.ID).First(&item)
	if item.SpecID != spec.ID {
		t.Errorf("expected order item spec_id %d, got %d", spec.ID, item.SpecID)
	}
	if item.SpecName != "800ml" {
		t.Errorf("expected spec_name 800ml, got %q", item.SpecName)
	}
	if item.Price != 18 {
		t.Errorf("expected item price 18, got %v", item.Price)
	}
}

// A spec that belongs to a different product is rejected.
func TestCreateOrder_RejectsSpecFromOtherProduct(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Spec Guard Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	p1 := models.Product{ShopID: shop.ID, Name: "A", Price: 10, Status: 1}
	config.DB.Create(&p1)
	p2 := models.Product{ShopID: shop.ID, Name: "B", Price: 20, Status: 1}
	config.DB.Create(&p2)
	otherSpec := models.ProductSpec{ProductID: p2.ID, Name: "大杯", Price: 25, Status: 1}
	config.DB.Create(&otherSpec)
	user := models.User{OpenID: "spec_guard_user", Nickname: "SpecGuard", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   1,
		"items":    []map[string]interface{}{{"product_id": p1.ID, "spec_id": otherSpec.ID, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for spec from other product, got %d body: %s", w.Code, w.Body.String())
	}
}

// A product that HAS active specs must not be ordered with spec_id:0 — that
// would let a client pay the base product price instead of the spec price.
func TestCreateOrder_RejectsMissingSpecWhenProductHasSpecs(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Spec Required Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "奶茶", Price: 10, Status: 1}
	config.DB.Create(&product)
	// Only spec is pricier than the base price — the underpay gap.
	config.DB.Create(&models.ProductSpec{ProductID: product.ID, Name: "大杯", Price: 18, Status: 1})
	user := models.User{OpenID: "spec_required_user", Nickname: "SpecRequired", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   1,
		"items":    []map[string]interface{}{{"product_id": product.ID, "spec_id": 0, "quantity": 1}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when spec omitted for a spec'd product, got %d body: %s", w.Code, w.Body.String())
	}
	var count int64
	config.DB.Model(&models.Order{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected no order created, got %d", count)
	}
}

// A product with NO specs is still orderable with spec_id:0 at its base price.
func TestCreateOrder_AllowsMissingSpecWhenProductHasNoSpecs(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "No Spec Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "矿泉水", Price: 3, Status: 1}
	config.DB.Create(&product)
	user := models.User{OpenID: "no_spec_user", Nickname: "NoSpec", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   1,
		"items":    []map[string]interface{}{{"product_id": product.ID, "spec_id": 0, "quantity": 2}},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for spec-less product, got %d body: %s", w.Code, w.Body.String())
	}
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.Amount != 6 {
		t.Errorf("expected amount 6 (base price x2), got %v", order.Amount)
	}
}

// GetShopProducts includes each product's specs.
func TestGetShopProducts_IncludesSpecs(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "List Spec Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)
	product := models.Product{ShopID: shop.ID, Name: "酸奶芒果", Price: 15, Status: 1}
	config.DB.Create(&product)
	config.DB.Create(&models.ProductSpec{ProductID: product.ID, Name: "600ml", Price: 15, Status: 1})
	config.DB.Create(&models.ProductSpec{ProductID: product.ID, Name: "800ml", Price: 18, Status: 1})

	r := setupRouter()
	r.GET("/api/shops/:id/products", GetShopProducts)
	req, _ := http.NewRequest("GET", "/api/shops/"+itoa(shop.ID)+"/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var products []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &products)
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	specs, ok := products[0]["specs"].([]interface{})
	if !ok || len(specs) != 2 {
		t.Errorf("expected 2 specs in product, got %v", products[0]["specs"])
	}
}
