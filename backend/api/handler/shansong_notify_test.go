package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
)

// postCallback builds a form-urlencoded Shansong callback with a (optionally
// valid) signature, carrying data {issOrderNo, orderStatus}.
func postCallback(t *testing.T, issOrderNo string, status int, validSign bool) *httptest.ResponseRecorder {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"issOrderNo": issOrderNo, "orderStatus": status})
	ts := "1700000000000"
	sign := services.Shansong.CallbackSign(ts, string(data))
	if !validSign {
		sign = "DEADBEEF"
	}
	form := url.Values{}
	form.Set("clientId", "c")
	form.Set("timestamp", ts)
	form.Set("sign", sign)
	form.Set("data", string(data))

	r := setupRouter()
	r.POST("/api/shansong/callback", ShansongCallback)
	req, _ := http.NewRequest("POST", "/api/shansong/callback", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
	od := seedDelivery(t, "SS-CB-1", 20)

	w := postCallback(t, "SS-CB-1", 40, true)
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
	if got.ShansongStatus != 40 {
		t.Errorf("expected status updated to 40 (闪送中), got %d", got.ShansongStatus)
	}
}

func TestShansongCallback_RejectsInvalidSign(t *testing.T) {
	setupTestDB(t)
	cleanup := withCallbackClient(t)
	defer cleanup()
	od := seedDelivery(t, "SS-CB-2", 20)

	w := postCallback(t, "SS-CB-2", 40, false)
	if w.Code == http.StatusOK {
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		if s, _ := resp["status"].(float64); s == 200 {
			t.Errorf("invalid sign must not return success status 200")
		}
	}
	var got models.OrderDelivery
	config.DB.First(&got, od.ID)
	if got.ShansongStatus != 20 {
		t.Errorf("status must NOT change on invalid sign, got %d", got.ShansongStatus)
	}
}

func TestShansongCallback_Idempotent(t *testing.T) {
	setupTestDB(t)
	cleanup := withCallbackClient(t)
	defer cleanup()
	od := seedDelivery(t, "SS-CB-3", 20)

	w1 := postCallback(t, "SS-CB-3", 40, true)
	w2 := postCallback(t, "SS-CB-3", 40, true)
	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Fatalf("expected both 200, got %d / %d", w1.Code, w2.Code)
	}
	var got models.OrderDelivery
	config.DB.First(&got, od.ID)
	if got.ShansongStatus != 40 {
		t.Errorf("expected stable status 40 after duplicate callback, got %d", got.ShansongStatus)
	}
}

// Matching also works when only the quote no holds the issOrderNo.
func TestShansongCallback_MatchesByQuoteNo(t *testing.T) {
	setupTestDB(t)
	cleanup := withCallbackClient(t)
	defer cleanup()
	od := models.OrderDelivery{OrderID: 8123, ShansongQuoteNo: "SS-Q-CB", ShansongStatus: 20}
	config.DB.Create(&od)

	w := postCallback(t, "SS-Q-CB", 50, true)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got models.OrderDelivery
	config.DB.First(&got, od.ID)
	if got.ShansongStatus != 50 {
		t.Errorf("expected status 50 matched by quote no, got %d", got.ShansongStatus)
	}
}
