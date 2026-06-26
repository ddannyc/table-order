package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
)

// postCallback builds a Shansong callback with a (optionally valid) signature.
func postCallback(t *testing.T, orderNo string, status int, validSign bool) *httptest.ResponseRecorder {
	t.Helper()
	biz, _ := json.Marshal(map[string]any{"orderNumber": orderNo, "status": status})
	ts := "1700000000000"
	sign := services.Shansong.CallbackSign(ts, string(biz))
	if !validSign {
		sign = "DEADBEEF"
	}
	body, _ := json.Marshal(map[string]any{
		"clientId": "c", "timestamp": ts, "sign": sign, "bizContent": string(biz),
	})
	r := setupRouter()
	r.POST("/api/shansong/callback", ShansongCallback)
	req, _ := http.NewRequest("POST", "/api/shansong/callback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func withCallbackClient(t *testing.T) func() {
	t.Helper()
	prev := services.Shansong
	services.Shansong = &services.ShansongClient{ClientID: "c", AppSecret: "s"}
	return func() { services.Shansong = prev }
}

func seedDelivery(t *testing.T, shansongNo string, status int) models.OrderDelivery {
	t.Helper()
	od := models.OrderDelivery{OrderID: 9000 + uint(status), ShansongOrderNo: shansongNo, ShansongStatus: status}
	if err := config.DB.Create(&od).Error; err != nil {
		t.Fatalf("seed delivery: %v", err)
	}
	return od
}

func TestShansongCallback_UpdatesStatusOnValidSign(t *testing.T) {
	setupTestDB(t)
	cleanup := withCallbackClient(t)
	defer cleanup()
	od := seedDelivery(t, "SS-CB-1", 60)

	w := postCallback(t, "SS-CB-1", 90, true)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if s, _ := resp["status"].(float64); s != 200 {
		t.Errorf(`expected body {"status":200}, got %v`, resp["status"])
	}
	var got models.OrderDelivery
	config.DB.First(&got, od.ID)
	if got.ShansongStatus != 90 {
		t.Errorf("expected status updated to 90, got %d", got.ShansongStatus)
	}
}

func TestShansongCallback_RejectsInvalidSign(t *testing.T) {
	setupTestDB(t)
	cleanup := withCallbackClient(t)
	defer cleanup()
	od := seedDelivery(t, "SS-CB-2", 60)

	w := postCallback(t, "SS-CB-2", 90, false)
	if w.Code == http.StatusOK {
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		if s, _ := resp["status"].(float64); s == 200 {
			t.Errorf("invalid sign must not return success status 200")
		}
	}
	var got models.OrderDelivery
	config.DB.First(&got, od.ID)
	if got.ShansongStatus != 60 {
		t.Errorf("status must NOT change on invalid sign, got %d", got.ShansongStatus)
	}
}

func TestShansongCallback_Idempotent(t *testing.T) {
	setupTestDB(t)
	cleanup := withCallbackClient(t)
	defer cleanup()
	od := seedDelivery(t, "SS-CB-3", 60)

	w1 := postCallback(t, "SS-CB-3", 90, true)
	w2 := postCallback(t, "SS-CB-3", 90, true)
	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Fatalf("expected both 200, got %d / %d", w1.Code, w2.Code)
	}
	var got models.OrderDelivery
	config.DB.First(&got, od.ID)
	if got.ShansongStatus != 90 {
		t.Errorf("expected stable status 90 after duplicate callback, got %d", got.ShansongStatus)
	}
}
