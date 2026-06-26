package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// stubDoer captures the outgoing request and returns a canned response.
type stubDoer struct {
	gotReq  *http.Request
	gotBody string
	resp    string
	status  int
	err     error
}

func (s *stubDoer) Do(req *http.Request) (*http.Response, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.gotReq = req
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		s.gotBody = string(b)
	}
	code := s.status
	if code == 0 {
		code = 200
	}
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(s.resp)),
		Header:     make(http.Header),
	}, nil
}

func fixedClient(d httpDoer) *ShansongClient {
	return &ShansongClient{
		ClientID:  "cid-1",
		AppSecret: "secret-1",
		ShopID:    "shop-9",
		BaseURL:   "https://open.s.bingex.com",
		HTTP:      d,
		Now:       func() time.Time { return time.Unix(1700000000, 0) },
	}
}

// Real Shansong recipe: MD5(appSecret + "clientId" + clientId + "data" + data +
// "shopId" + shopId + "timestamp" + timestamp), uppercase; data segment omitted
// when data is empty. Golden value computed from the documented rule.
func TestShansongSign_RealRecipe(t *testing.T) {
	got := signShansong("cid-1", "secret-1", "shop-9", "1700000000000", `{"a":1}`)
	if got != "683042FE4702409D69A75ABB10D43409" {
		t.Fatalf("sign mismatch vs golden, got %s", got)
	}
	if len(got) != 32 || got != strings.ToUpper(got) {
		t.Errorf("expected 32-char uppercase hex, got %q", got)
	}
	// Must depend on shopId and appSecret.
	if signShansong("cid-1", "secret-1", "other-shop", "1700000000000", `{"a":1}`) == got {
		t.Errorf("sign must depend on shopId")
	}
	if signShansong("cid-1", "other", "shop-9", "1700000000000", `{"a":1}`) == got {
		t.Errorf("sign must depend on appSecret")
	}
	// Empty data omits the data segment, so the signature differs.
	if signShansong("cid-1", "secret-1", "shop-9", "1700000000000", "") == got {
		t.Errorf("empty data must omit the data segment and change the signature")
	}
}

func TestCalculatePrice_SignsRequestAndParsesFee(t *testing.T) {
	d := &stubDoer{resp: `{"status":200,"msg":"success","data":{"totalFeeYuan":8.5,"orderNumber":"SS-Q-1"}}`}
	c := fixedClient(d)

	res, err := c.CalculatePrice(context.Background(), QuoteRequest{
		SenderLat: 39.9, SenderLng: 116.4, SenderAddress: "门店",
		RecipientName: "张三", RecipientPhone: "13800000000",
		RecipientLat: 39.91, RecipientLng: 116.41, RecipientAddress: "收件",
	})
	if err != nil {
		t.Fatalf("CalculatePrice failed: %v", err)
	}
	if res.DeliveryFee != 8.5 {
		t.Errorf("expected fee 8.5, got %v", res.DeliveryFee)
	}
	if res.QuoteToken != "SS-Q-1" {
		t.Errorf("expected quote token SS-Q-1, got %q", res.QuoteToken)
	}
	// The outgoing request must be form-urlencoded with the system params.
	if ct := d.gotReq.Header.Get("Content-Type"); !strings.Contains(ct, "x-www-form-urlencoded") {
		t.Errorf("expected form-urlencoded content type, got %q", ct)
	}
	for _, want := range []string{"clientId=cid-1", "shopId=shop-9", "sign=", "data="} {
		if !strings.Contains(d.gotBody, want) {
			t.Errorf("request body missing %q: %s", want, d.gotBody)
		}
	}
}

func TestCalculatePrice_ErrorOnNonSuccessStatus(t *testing.T) {
	d := &stubDoer{resp: `{"status":500,"msg":"out of range","data":null}`}
	c := fixedClient(d)
	_, err := c.CalculatePrice(context.Background(), QuoteRequest{})
	if err == nil {
		t.Fatalf("expected error on non-200 shansong status, got nil")
	}
}

func TestCreateOrder_ParsesOrderNo(t *testing.T) {
	d := &stubDoer{resp: `{"status":200,"msg":"success","data":{"orderNumber":"SS-ORD-9"}}`}
	c := fixedClient(d)
	no, err := c.CreateOrder(context.Background(), CreateDeliveryRequest{QuoteToken: "SS-Q-1"})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if no != "SS-ORD-9" {
		t.Errorf("expected order no SS-ORD-9, got %q", no)
	}
}

func TestShansongStatusLabel(t *testing.T) {
	// Known codes map to Chinese labels; unknown codes fall back, never panic.
	if ShansongStatusLabel(0) == "" {
		t.Errorf("status 0 should have a label")
	}
	if ShansongStatusLabel(999999) == "" {
		t.Errorf("unknown status should fall back to a non-empty label")
	}
}

// Guard: a transport error surfaces, not a panic.
func TestCalculatePrice_TransportError(t *testing.T) {
	d := &stubDoer{err: io.ErrUnexpectedEOF}
	c := fixedClient(d)
	if _, err := c.CalculatePrice(context.Background(), QuoteRequest{}); err == nil {
		t.Fatalf("expected transport error to surface")
	}
}

var _ = json.Marshal
