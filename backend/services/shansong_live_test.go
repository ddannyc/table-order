package services

import (
	"context"
	"os"
	"testing"
)

// TestLive_CalculatePrice exercises the real Shansong test environment with the
// actual ShansongClient (our sign + transport + body + parse). Skipped unless
// SHANSONG_LIVE=1, so it never runs in normal CI. Run it during 联调:
//
//	SHANSONG_LIVE=1 SHANSONG_CLIENT_ID=.. SHANSONG_APP_SECRET=.. SHANSONG_SHOP_ID=.. \
//	SHANSONG_BASE_URL=http://open.s.bingex.com go test ./services/ -run TestLive_CalculatePrice -v
func TestLive_CalculatePrice(t *testing.T) {
	if os.Getenv("SHANSONG_LIVE") != "1" {
		t.Skip("set SHANSONG_LIVE=1 (with creds) to run the live shansong test")
	}
	c := &ShansongClient{
		ClientID:  os.Getenv("SHANSONG_CLIENT_ID"),
		AppSecret: os.Getenv("SHANSONG_APP_SECRET"),
		ShopID:    os.Getenv("SHANSONG_SHOP_ID"),
		BaseURL:   os.Getenv("SHANSONG_BASE_URL"),
	}
	res, err := c.CalculatePrice(context.Background(), QuoteRequest{
		CityName:      "北京市",
		SenderAddress: "北京市朝阳区望京SOHO", SenderName: "测试餐饮店", SenderMobile: "19020243001",
		SenderLat: 39.996525, SenderLng: 116.481488,
		ThirdOrderNo:  "LIVE_VERIFY_1",
		RecipientName: "收件专员", RecipientPhone: "13800138001",
		RecipientAddress: "北京市朝阳区阜通东大街6号", RecipientLat: 39.992405, RecipientLng: 116.473500,
	})
	if err != nil {
		t.Fatalf("live CalculatePrice failed: %v", err)
	}
	t.Logf("live quote: fee=%.2f 元, issOrderNo=%s", res.DeliveryFee, res.QuoteToken)
	if res.DeliveryFee <= 0 {
		t.Errorf("expected a positive delivery fee, got %v", res.DeliveryFee)
	}
	if res.QuoteToken == "" {
		t.Errorf("expected an orderNumber (issOrderNo) from the quote")
	}
}
