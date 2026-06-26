package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// dataParam decodes the form-encoded request body and returns the parsed `data`
// business JSON object.
func dataParam(t *testing.T, body string) map[string]any {
	t.Helper()
	v, err := url.ParseQuery(body)
	if err != nil {
		t.Fatalf("parse form body: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(v.Get("data")), &m); err != nil {
		t.Fatalf("parse data json: %v (raw=%s)", err, v.Get("data"))
	}
	return m
}

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

func TestCalculatePrice_BuildsOrderCalculateRequestAndParsesFee(t *testing.T) {
	// Real orderCalculate response shape: fees in 分, totalFeeAfterSave is payable.
	d := &stubDoer{resp: `{"status":200,"msg":"success","data":{"totalAmount":900,"totalFeeAfterSave":850,"couponSaveFee":50,"orderNumber":"SS-Q-1"}}`}
	c := fixedClient(d)

	res, err := c.CalculatePrice(context.Background(), QuoteRequest{
		CityName:      "北京市",
		SenderAddress: "门店", SenderName: "测试餐饮店", SenderMobile: "19000000000",
		SenderLat: 39.9, SenderLng: 116.4,
		ThirdOrderNo:  "OUR-1",
		RecipientName: "张三", RecipientPhone: "13800000000",
		RecipientAddress: "收件", RecipientLat: 39.91, RecipientLng: 116.41,
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
	if ct := d.gotReq.Header.Get("Content-Type"); !strings.Contains(ct, "x-www-form-urlencoded") {
		t.Errorf("expected form-urlencoded, got %q", ct)
	}

	// Real orderCalculate structure: cityName + nested sender + receiverList,
	// string coords, goodType 6 (餐饮).
	data := dataParam(t, d.gotBody)
	if data["cityName"] != "北京市" {
		t.Errorf("missing cityName, got %v", data["cityName"])
	}
	sender, _ := data["sender"].(map[string]any)
	if sender == nil || sender["fromSenderName"] != "测试餐饮店" {
		t.Fatalf("sender block wrong: %v", data["sender"])
	}
	if sender["fromLatitude"] != "39.900000" {
		t.Errorf("expected string sender coord 39.900000, got %v", sender["fromLatitude"])
	}
	recv, _ := data["receiverList"].([]any)
	if len(recv) != 1 {
		t.Fatalf("expected one receiver, got %v", data["receiverList"])
	}
	r0 := recv[0].(map[string]any)
	if r0["orderNo"] != "OUR-1" {
		t.Errorf("receiver orderNo (thirdOrderNo) wrong: %v", r0["orderNo"])
	}
	if r0["toMobile"] != "13800000000" {
		t.Errorf("receiver mobile wrong: %v", r0["toMobile"])
	}
	if gt, _ := r0["goodType"].(float64); gt != 6 {
		t.Errorf("goodType should be 6 (餐饮), got %v", r0["goodType"])
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

func TestCreateOrder_SendsOnlyIssOrderNo(t *testing.T) {
	d := &stubDoer{resp: `{"status":200,"msg":"success","data":{"orderNumber":"SS-ORD-9"}}`}
	c := fixedClient(d)
	no, err := c.CreateOrder(context.Background(), CreateDeliveryRequest{QuoteToken: "SS-Q-1"})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if no != "SS-ORD-9" {
		t.Errorf("expected order no SS-ORD-9, got %q", no)
	}
	// orderPlace business body is ONLY {issOrderNo}.
	data := dataParam(t, d.gotBody)
	if data["issOrderNo"] != "SS-Q-1" {
		t.Errorf("orderPlace must send issOrderNo, got %v", data["issOrderNo"])
	}
	if len(data) != 1 {
		t.Errorf("orderPlace should send only issOrderNo, got %v", data)
	}
}

func TestShansongStatusLabel(t *testing.T) {
	cases := map[int]string{0: "待派单", -1: "派单失败", 20: "派单中", 30: "待取货", 40: "闪送中", 50: "已完成", 60: "已取消"}
	for code, want := range cases {
		if got := ShansongStatusLabel(code); got != want {
			t.Errorf("status %d: want %q got %q", code, want, got)
		}
	}
	if ShansongStatusLabel(999) != "配送中" {
		t.Errorf("unknown nonzero status should fall back to 配送中, got %q", ShansongStatusLabel(999))
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
