package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

// doMerchantReq runs a single authenticated merchant request through a router
// that maps routePath → handler. body is JSON-encoded when non-nil.
func doMerchantReq(t *testing.T, merchantID uint, method, routePath, url string, handler gin.HandlerFunc, body any) *httptest.ResponseRecorder {
	t.Helper()
	r := setupRouter()
	setAuthContext(r, method, routePath, handler, merchantID)
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// merchantOrdersResp is the decoded shape the order board relies on.
type merchantOrdersResp struct {
	Orders []struct {
		OrderNo   string `json:"order_no"`
		OrderType string `json:"order_type"`
		Status    int    `json:"status"`
		Delivery  *struct {
			RecipientName  string `json:"recipient_name"`
			ShansongStatus int    `json:"shansong_status"`
		} `json:"delivery"`
	} `json:"orders"`
	Total    int64   `json:"total"`
	Revenue  float64 `json:"revenue"`
	Rewarded float64 `json:"rewarded"`
}

// getMerchantOrders issues GET /api/merchant/orders<query> as the given merchant.
func getMerchantOrders(t *testing.T, merchantID uint, query string) merchantOrdersResp {
	t.Helper()
	r := setupRouter()
	setAuthContext(r, "GET", "/api/merchant/orders", GetMerchantOrders, merchantID)
	req, _ := http.NewRequest("GET", "/api/merchant/orders"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var resp merchantOrdersResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body: %s", err, w.Body.String())
	}
	return resp
}

// T1: Order.PreparedAt (出餐时间) persists as a nullable column — nil by default,
// settable, and round-trips. 已出餐 = PreparedAt != nil.
func TestOrder_PreparedAtPersists(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Prepared Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)

	order := models.Order{OrderNo: "PREP_PERSIST_001", ShopID: shop.ID, Amount: 10, Status: 2}
	config.DB.Create(&order)

	var fresh models.Order
	config.DB.First(&fresh, order.ID)
	if fresh.PreparedAt != nil {
		t.Fatalf("expected PreparedAt nil by default, got %v", fresh.PreparedAt)
	}

	now := time.Now()
	config.DB.Model(&fresh).Update("prepared_at", now)

	var reloaded models.Order
	config.DB.First(&reloaded, order.ID)
	if reloaded.PreparedAt == nil {
		t.Fatalf("expected PreparedAt non-nil after update")
	}
}

// T2: list embeds delivery detail for delivery orders (null for dine-in) and
// includes unpaid (status=1) orders.
func TestGetMerchantOrders_EmbedsDeliveryAndIncludesUnpaid(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9201)
	shop := models.Shop{Name: "Board Shop A", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)

	dineIn := models.Order{OrderNo: "BD_DINE_1", ShopID: shop.ID, OrderType: "dine_in", Amount: 80, Status: 2}
	config.DB.Create(&dineIn)
	unpaid := models.Order{OrderNo: "BD_UNPAID_1", ShopID: shop.ID, OrderType: "dine_in", Amount: 50, Status: 1}
	config.DB.Create(&unpaid)
	deliv := models.Order{OrderNo: "BD_DELIV_1", ShopID: shop.ID, OrderType: "delivery", Amount: 120, Status: 2}
	config.DB.Create(&deliv)
	config.DB.Create(&models.OrderDelivery{OrderID: deliv.ID, RecipientName: "张三", ShansongStatus: -1})

	resp := getMerchantOrders(t, merchantID, "")
	if resp.Total != 3 {
		t.Fatalf("expected total 3 (incl. unpaid), got %d", resp.Total)
	}
	byNo := map[string]int{}
	for i, o := range resp.Orders {
		byNo[o.OrderNo] = i
	}
	if _, ok := byNo["BD_UNPAID_1"]; !ok {
		t.Errorf("unpaid order must be returned")
	}
	di := resp.Orders[byNo["BD_DINE_1"]]
	if di.Delivery != nil {
		t.Errorf("dine_in order must have null delivery, got %+v", di.Delivery)
	}
	dv := resp.Orders[byNo["BD_DELIV_1"]]
	if dv.Delivery == nil {
		t.Fatalf("delivery order must embed delivery detail")
	}
	if dv.Delivery.ShansongStatus != -1 || dv.Delivery.RecipientName != "张三" {
		t.Errorf("unexpected delivery detail: %+v", dv.Delivery)
	}
}

// T2: status and type filters narrow the result set.
func TestGetMerchantOrders_FiltersByStatusAndType(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9202)
	shop := models.Shop{Name: "Board Shop B", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)

	config.DB.Create(&models.Order{OrderNo: "FL_DINE_PAID", ShopID: shop.ID, OrderType: "dine_in", Amount: 10, Status: 2})
	config.DB.Create(&models.Order{OrderNo: "FL_DINE_UNPAID", ShopID: shop.ID, OrderType: "dine_in", Amount: 10, Status: 1})
	deliv := models.Order{OrderNo: "FL_DELIV_PAID", ShopID: shop.ID, OrderType: "delivery", Amount: 10, Status: 2}
	config.DB.Create(&deliv)
	config.DB.Create(&models.OrderDelivery{OrderID: deliv.ID, ShansongStatus: 20})

	byStatus := getMerchantOrders(t, merchantID, "?status=1")
	if byStatus.Total != 1 || len(byStatus.Orders) != 1 || byStatus.Orders[0].OrderNo != "FL_DINE_UNPAID" {
		t.Errorf("status=1 filter failed: total=%d orders=%+v", byStatus.Total, byStatus.Orders)
	}
	byType := getMerchantOrders(t, merchantID, "?type=delivery")
	if byType.Total != 1 || len(byType.Orders) != 1 || byType.Orders[0].OrderNo != "FL_DELIV_PAID" {
		t.Errorf("type=delivery filter failed: total=%d orders=%+v", byType.Total, byType.Orders)
	}
}

// T2: pagination caps the page, but total + revenue/rewarded reflect the full
// filtered set, not just the returned page.
func TestGetMerchantOrders_PaginatesAndAggregatesOverFullSet(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9203)
	shop := models.Shop{Name: "Board Shop C", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)
	for i := 0; i < 3; i++ {
		config.DB.Create(&models.Order{
			OrderNo: "PG_" + string(rune('A'+i)), ShopID: shop.ID, OrderType: "dine_in",
			Amount: 100, RewardAmount: 10, Status: 2,
		})
	}

	resp := getMerchantOrders(t, merchantID, "?page=1&page_size=2")
	if len(resp.Orders) != 2 {
		t.Errorf("expected 2 orders on page, got %d", len(resp.Orders))
	}
	if resp.Total != 3 {
		t.Errorf("expected total 3 across full set, got %d", resp.Total)
	}
	if resp.Revenue != 300 {
		t.Errorf("expected revenue 300 over full set, got %v", resp.Revenue)
	}
	if resp.Rewarded != 30 {
		t.Errorf("expected rewarded 30 over full set, got %v", resp.Rewarded)
	}
}

// --- T3: 出餐 (PrepareOrder) ---

func TestPrepareOrder_SetsPreparedAt(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9301)
	shop := models.Shop{Name: "Prep Shop", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "PR_OK", ShopID: shop.ID, OrderType: "dine_in", Amount: 30, Status: 2}
	config.DB.Create(&order)

	w := doMerchantReq(t, merchantID, "POST", "/api/merchant/orders/:id/prepare",
		"/api/merchant/orders/"+itoa(order.ID)+"/prepare", PrepareOrder, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var got models.Order
	config.DB.First(&got, order.ID)
	if got.PreparedAt == nil {
		t.Fatalf("expected PreparedAt set after prepare")
	}
}

func TestPrepareOrder_Idempotent(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9302)
	shop := models.Shop{Name: "Prep Shop 2", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "PR_IDEM", ShopID: shop.ID, OrderType: "dine_in", Amount: 30, Status: 2}
	config.DB.Create(&order)

	doMerchantReq(t, merchantID, "POST", "/api/merchant/orders/:id/prepare",
		"/api/merchant/orders/"+itoa(order.ID)+"/prepare", PrepareOrder, nil)
	var first models.Order
	config.DB.First(&first, order.ID)

	w := doMerchantReq(t, merchantID, "POST", "/api/merchant/orders/:id/prepare",
		"/api/merchant/orders/"+itoa(order.ID)+"/prepare", PrepareOrder, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on repeat, got %d", w.Code)
	}
	var second models.Order
	config.DB.First(&second, order.ID)
	if !first.PreparedAt.Equal(*second.PreparedAt) {
		t.Errorf("prepare must be idempotent (time unchanged): %v vs %v", first.PreparedAt, second.PreparedAt)
	}
}

func TestPrepareOrder_RejectsOtherMerchant(t *testing.T) {
	setupTestDB(t)
	owner := uint(9303)
	other := uint(9304)
	shop := models.Shop{Name: "Owner Shop", MerchantID: owner, Status: 1}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "PR_FORBID", ShopID: shop.ID, OrderType: "dine_in", Amount: 30, Status: 2}
	config.DB.Create(&order)

	w := doMerchantReq(t, other, "POST", "/api/merchant/orders/:id/prepare",
		"/api/merchant/orders/"+itoa(order.ID)+"/prepare", PrepareOrder, nil)
	if w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Fatalf("expected 403/404 for other merchant, got %d", w.Code)
	}
	var got models.Order
	config.DB.First(&got, order.ID)
	if got.PreparedAt != nil {
		t.Errorf("must not prepare another merchant's order")
	}
}

func TestPrepareOrder_NotFound(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9305)
	w := doMerchantReq(t, merchantID, "POST", "/api/merchant/orders/:id/prepare",
		"/api/merchant/orders/999999/prepare", PrepareOrder, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing order, got %d", w.Code)
	}
}

// --- T4: 改状态 (UpdateMerchantOrderStatus) ---

func TestUpdateOrderStatus_UpdatesValid(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9401)
	shop := models.Shop{Name: "Status Shop", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "ST_OK", ShopID: shop.ID, OrderType: "dine_in", Amount: 30, Status: 2}
	config.DB.Create(&order)

	w := doMerchantReq(t, merchantID, "PUT", "/api/merchant/orders/:id/status",
		"/api/merchant/orders/"+itoa(order.ID)+"/status", UpdateMerchantOrderStatus, map[string]any{"status": 3})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var got models.Order
	config.DB.First(&got, order.ID)
	if got.Status != 3 {
		t.Errorf("expected status 3, got %d", got.Status)
	}
}

func TestUpdateOrderStatus_RejectsInvalid(t *testing.T) {
	setupTestDB(t)
	const merchantID = uint(9402)
	shop := models.Shop{Name: "Status Shop 2", MerchantID: merchantID, Status: 1}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "ST_BAD", ShopID: shop.ID, OrderType: "dine_in", Amount: 30, Status: 2}
	config.DB.Create(&order)

	w := doMerchantReq(t, merchantID, "PUT", "/api/merchant/orders/:id/status",
		"/api/merchant/orders/"+itoa(order.ID)+"/status", UpdateMerchantOrderStatus, map[string]any{"status": 9})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", w.Code)
	}
	var got models.Order
	config.DB.First(&got, order.ID)
	if got.Status != 2 {
		t.Errorf("status must stay 2 after rejected update, got %d", got.Status)
	}
}

func TestUpdateOrderStatus_RejectsOtherMerchant(t *testing.T) {
	setupTestDB(t)
	owner := uint(9403)
	other := uint(9404)
	shop := models.Shop{Name: "Owner Status Shop", MerchantID: owner, Status: 1}
	config.DB.Create(&shop)
	order := models.Order{OrderNo: "ST_FORBID", ShopID: shop.ID, OrderType: "dine_in", Amount: 30, Status: 2}
	config.DB.Create(&order)

	w := doMerchantReq(t, other, "PUT", "/api/merchant/orders/:id/status",
		"/api/merchant/orders/"+itoa(order.ID)+"/status", UpdateMerchantOrderStatus, map[string]any{"status": 4})
	if w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
		t.Fatalf("expected 403/404, got %d", w.Code)
	}
	var got models.Order
	config.DB.First(&got, order.ID)
	if got.Status != 2 {
		t.Errorf("must not change another merchant's order status, got %d", got.Status)
	}
}
