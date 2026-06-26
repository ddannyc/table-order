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

// stubShansongServer returns an httptest server that answers the Shansong quote
// endpoint with the given envelope, plus a client pointed at it.
func withStubShansong(t *testing.T, envelope string) func() {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(envelope))
	}))
	prev := services.Shansong
	services.Shansong = &services.ShansongClient{ClientID: "c", AppSecret: "s", BaseURL: ts.URL}
	return func() {
		services.Shansong = prev
		ts.Close()
	}
}

func TestDeliveryQuote_ReturnsFeeAndVerifiableToken(t *testing.T) {
	setupTestDB(t)
	cleanup := withStubShansong(t, `{"status":200,"msg":"ok","data":{"totalFeeAfterCommission":8.5,"orderNumber":"SS-Q-1"}}`)
	defer cleanup()

	shop := models.Shop{Name: "Geo Shop", MerchantID: 1, Status: 1, City: "北京市", Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)

	r := setupRouter()
	r.POST("/api/delivery/quote", DeliveryQuote)
	body, _ := json.Marshal(map[string]any{
		"shop_id":           shop.ID,
		"recipient_name":    "张三",
		"recipient_phone":   "13800000000",
		"recipient_address": "收件地址",
		"recipient_lat":     39.91,
		"recipient_lng":     116.41,
	})
	req, _ := http.NewRequest("POST", "/api/delivery/quote", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if fee, _ := resp["delivery_fee"].(float64); fee != 8.5 {
		t.Errorf("expected delivery_fee 8.5, got %v", resp["delivery_fee"])
	}
	token, _ := resp["quote_token"].(string)
	if token == "" {
		t.Fatalf("expected quote_token in response")
	}
	// The token must verify and carry the trusted fee + coords.
	claims, err := verifyQuoteToken(token)
	if err != nil {
		t.Fatalf("quote token should verify: %v", err)
	}
	if claims.Fee != 8.5 {
		t.Errorf("token fee mismatch: %v", claims.Fee)
	}
	if claims.ShopID != shop.ID {
		t.Errorf("token shop mismatch: %v", claims.ShopID)
	}
}

func TestDeliveryQuote_RejectsMissingShopCoords(t *testing.T) {
	setupTestDB(t)
	cleanup := withStubShansong(t, `{"status":200,"data":{"totalFeeAfterCommission":8.5,"orderNumber":"X"}}`)
	defer cleanup()

	shop := models.Shop{Name: "No Geo Shop", MerchantID: 1, Status: 1} // lat/lng 0
	config.DB.Create(&shop)

	r := setupRouter()
	r.POST("/api/delivery/quote", DeliveryQuote)
	body, _ := json.Marshal(map[string]any{
		"shop_id": shop.ID, "recipient_lat": 39.91, "recipient_lng": 116.41,
	})
	req, _ := http.NewRequest("POST", "/api/delivery/quote", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when shop has no coords, got %d", w.Code)
	}
}

func TestDeliveryQuote_RejectsMissingCity(t *testing.T) {
	setupTestDB(t)
	cleanup := withStubShansong(t, `{"status":200,"data":{"totalFeeAfterCommission":8.5,"orderNumber":"X"}}`)
	defer cleanup()

	// Coords set but no city → orderCalculate can't be built.
	shop := models.Shop{Name: "No City Shop", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)

	r := setupRouter()
	r.POST("/api/delivery/quote", DeliveryQuote)
	body, _ := json.Marshal(map[string]any{
		"shop_id": shop.ID, "recipient_lat": 39.91, "recipient_lng": 116.41,
	})
	req, _ := http.NewRequest("POST", "/api/delivery/quote", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when shop has no city, got %d", w.Code)
	}
}

func TestDeliveryQuote_RejectsMissingRecipientCoords(t *testing.T) {
	setupTestDB(t)
	cleanup := withStubShansong(t, `{"status":200,"data":{"totalFeeAfterCommission":8.5,"orderNumber":"X"}}`)
	defer cleanup()

	shop := models.Shop{Name: "Geo Shop2", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)

	r := setupRouter()
	r.POST("/api/delivery/quote", DeliveryQuote)
	body, _ := json.Marshal(map[string]any{"shop_id": shop.ID}) // no recipient coords
	req, _ := http.NewRequest("POST", "/api/delivery/quote", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when recipient coords missing, got %d", w.Code)
	}
}

func TestDeliveryQuote_QuoteFailureSurfaces(t *testing.T) {
	setupTestDB(t)
	cleanup := withStubShansong(t, `{"status":500,"msg":"out of range","data":null}`)
	defer cleanup()

	shop := models.Shop{Name: "Geo Shop3", MerchantID: 1, Status: 1, Latitude: 39.9, Longitude: 116.4}
	config.DB.Create(&shop)

	r := setupRouter()
	r.POST("/api/delivery/quote", DeliveryQuote)
	body, _ := json.Marshal(map[string]any{
		"shop_id": shop.ID, "recipient_lat": 39.91, "recipient_lng": 116.41,
	})
	req, _ := http.NewRequest("POST", "/api/delivery/quote", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code < 400 {
		t.Errorf("expected 4xx/5xx on quote failure, got %d body: %s", w.Code, w.Body.String())
	}
}

func TestVerifyQuoteToken_RejectsTampered(t *testing.T) {
	token := signQuoteToken(quoteClaims{ShopID: 1, Fee: 8.5, Lat: 39.9, Lng: 116.4, ShansongQuote: "Q", Exp: 1 << 62})
	if _, err := verifyQuoteToken(token); err != nil {
		t.Fatalf("valid token should verify: %v", err)
	}
	if _, err := verifyQuoteToken(token + "x"); err == nil {
		t.Errorf("tampered token must be rejected")
	}
}

func TestVerifyQuoteToken_RejectsExpired(t *testing.T) {
	token := signQuoteToken(quoteClaims{ShopID: 1, Fee: 8.5, Exp: 1}) // far past
	if _, err := verifyQuoteToken(token); err == nil {
		t.Errorf("expired token must be rejected")
	}
}
