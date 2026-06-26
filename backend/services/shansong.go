package services

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ShansongClient talks to the Shansong (闪送) open platform merchants/v5 API for
// instant-courier delivery: realtime price quote (orderCalculate), order create
// (orderPlace), and cancel (abortOrder).
//
// Protocol (per official docs): POST application/x-www-form-urlencoded with
// system params clientId/shopId/timestamp/sign and a compact-JSON business param
// `data`. CALIBRATION remaining: the orderCalculate response fee field name and
// the inbound callback payload format (confirmed at 联调).
type ShansongClient struct {
	ClientID  string
	AppSecret string
	ShopID    string
	BaseURL   string // open.s.bingex.com（测试）/ open.ishansong.com（生产）
	HTTP      httpDoer
	Now       func() time.Time
}

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Shansong is the process-wide client, initialised from config at startup.
// Nil when credentials are not configured (delivery quoting then 503s).
var Shansong *ShansongClient

// InitShansongClient builds the process-wide client. No-op (leaves Shansong nil)
// when credentials are absent, so a deployment without delivery just disables it.
func InitShansongClient(clientID, appSecret, shopID, baseURL string) {
	if clientID == "" || appSecret == "" {
		return
	}
	Shansong = &ShansongClient{ClientID: clientID, AppSecret: appSecret, ShopID: shopID, BaseURL: baseURL}
}

// Shansong merchants/v5 API paths.
const (
	shansongPathQuote  = "/openapi/merchants/v5/orderCalculate"
	shansongPathCreate = "/openapi/merchants/v5/orderPlace"
	shansongPathCancel = "/openapi/merchants/v5/abortOrder"
)

// QuoteRequest is the business input for a delivery price quote.
type QuoteRequest struct {
	CityName         string // 寄件城市（orderCalculate 必填）
	SenderAddress    string
	SenderName       string // 寄件方名称（门店名）
	SenderMobile     string // 寄件方电话（门店电话）
	SenderLat        float64
	SenderLng        float64
	ThirdOrderNo     string // 第三方单号（= 我方 order_no），闪送 receiverList[].orderNo
	RecipientName    string
	RecipientPhone   string
	RecipientAddress string
	RecipientLat     float64
	RecipientLng     float64
}

// QuoteResult is the parsed quote: the fee plus Shansong's quote token, which is
// passed back at order create time so Shansong prices the same quote.
type QuoteResult struct {
	DeliveryFee float64
	QuoteToken  string
}

// CreateDeliveryRequest dispatches against a prior quote. orderPlace needs only
// the shansong issOrderNo (QuoteToken); OrderNo is kept for logging/cross-ref.
type CreateDeliveryRequest struct {
	QuoteToken string // shansong issOrderNo from the quote
	OrderNo    string // our order_no (for logging / cross-reference)
}

// shansongResp is the common Shansong response envelope: status==200 means OK.
type shansongResp struct {
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

// signShansong builds the request signature per the official rule:
// MD5(appSecret + "clientId" + clientId + "data" + data + "shopId" + shopId +
// "timestamp" + timestamp), uppercase. The "data" segment is omitted when data
// is empty.
func signShansong(clientID, appSecret, shopID, timestamp, data string) string {
	var b strings.Builder
	b.WriteString(appSecret)
	b.WriteString("clientId")
	b.WriteString(clientID)
	if data != "" {
		b.WriteString("data")
		b.WriteString(data)
	}
	b.WriteString("shopId")
	b.WriteString(shopID)
	b.WriteString("timestamp")
	b.WriteString(timestamp)
	sum := md5.Sum([]byte(b.String()))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// CallbackSign computes the signature for a callback payload via the same recipe.
func (c *ShansongClient) CallbackSign(timestamp, data string) string {
	return signShansong(c.ClientID, c.AppSecret, c.ShopID, timestamp, data)
}

// VerifyCallback reports whether sign matches the expected signature for the
// given callback payload (constant-time compare).
func (c *ShansongClient) VerifyCallback(timestamp, data, sign string) bool {
	expected := c.CallbackSign(timestamp, data)
	return hmac.Equal([]byte(expected), []byte(strings.ToUpper(sign)))
}

func (c *ShansongClient) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *ShansongClient) httpClient() httpDoer {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 10 * time.Second}
}

// post signs and sends a business payload, returning the parsed envelope.
// It errors if transport fails, the envelope is unparseable, or status != 200.
func (c *ShansongClient) post(ctx context.Context, path string, biz map[string]any) (*shansongResp, error) {
	var data string
	if len(biz) > 0 {
		b, err := json.Marshal(biz)
		if err != nil {
			return nil, fmt.Errorf("shansong: marshal biz: %w", err)
		}
		data = string(b)
	}
	timestamp := strconv.FormatInt(c.now().UnixMilli(), 10)
	sign := signShansong(c.ClientID, c.AppSecret, c.ShopID, timestamp, data)

	form := url.Values{}
	form.Set("clientId", c.ClientID)
	form.Set("shopId", c.ShopID)
	form.Set("timestamp", timestamp)
	form.Set("sign", sign)
	if data != "" {
		form.Set("data", data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("shansong: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("shansong: transport: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("shansong: read body: %w", err)
	}
	var env shansongResp
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("shansong: parse envelope: %w (body=%s)", err, string(raw))
	}
	if env.Status != 200 {
		return nil, fmt.Errorf("shansong: status %d: %s", env.Status, env.Msg)
	}
	return &env, nil
}

// fmtCoord renders a coordinate as Shansong expects — a fixed-precision string.
func fmtCoord(v float64) string {
	return strconv.FormatFloat(v, 'f', 6, 64)
}

// CalculatePrice returns the realtime delivery fee + quote token (orderCalculate).
func (c *ShansongClient) CalculatePrice(ctx context.Context, q QuoteRequest) (*QuoteResult, error) {
	biz := map[string]any{
		"cityName":    q.CityName,
		"appointType": 0, // 实时单
		"sender": map[string]any{
			"fromAddress":    q.SenderAddress,
			"fromSenderName": q.SenderName,
			"fromMobile":     q.SenderMobile,
			"fromLatitude":   fmtCoord(q.SenderLat),
			"fromLongitude":  fmtCoord(q.SenderLng),
		},
		"receiverList": []map[string]any{
			{
				"orderNo":        q.ThirdOrderNo,
				"toAddress":      q.RecipientAddress,
				"toReceiverName": q.RecipientName,
				"toMobile":       q.RecipientPhone,
				"toLatitude":     fmtCoord(q.RecipientLat),
				"toLongitude":    fmtCoord(q.RecipientLng),
				"goodType":       6, // 餐饮
				"weight":         1, // kg
			},
		},
	}
	env, err := c.post(ctx, shansongPathQuote, biz)
	if err != nil {
		return nil, err
	}
	// orderCalculate returns fees in 分 (cents). totalFeeAfterSave is the
	// merchant-payable total after coupon; totalAmount is the pre-coupon total.
	// (Verified against a live test orderCalculate response.)
	var data struct {
		TotalFeeAfterSave int64  `json:"totalFeeAfterSave"`
		TotalAmount       int64  `json:"totalAmount"`
		OrderNumber       string `json:"orderNumber"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("shansong: parse quote data: %w", err)
	}
	feeCents := data.TotalFeeAfterSave
	if feeCents == 0 {
		feeCents = data.TotalAmount
	}
	return &QuoteResult{DeliveryFee: float64(feeCents) / 100.0, QuoteToken: data.OrderNumber}, nil
}

// CreateOrder places the courier order against a prior quote (orderPlace). The
// business body is only {issOrderNo} — the quote's orderNumber. Returns the
// Shansong order number (echoed, or the issOrderNo if not returned).
func (c *ShansongClient) CreateOrder(ctx context.Context, r CreateDeliveryRequest) (string, error) {
	biz := map[string]any{"issOrderNo": r.QuoteToken}
	env, err := c.post(ctx, shansongPathCreate, biz)
	if err != nil {
		return "", err
	}
	var data struct {
		OrderNumber string `json:"orderNumber"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return "", fmt.Errorf("shansong: parse create data: %w", err)
	}
	if data.OrderNumber != "" {
		return data.OrderNumber, nil
	}
	return r.QuoteToken, nil // orderPlace confirms the same number
}

// CancelOrder cancels a dispatched delivery. Wired but not exposed via a route
// this iteration (cancel/refund is out of scope per plan); kept for completeness.
func (c *ShansongClient) CancelOrder(ctx context.Context, shansongOrderNo string) error {
	biz := map[string]any{"orderNumber": shansongOrderNo}
	_, err := c.post(ctx, shansongPathCancel, biz)
	return err
}

// shansongStatusLabels maps Shansong orderStatus codes to Chinese display text
// (per the merchants/v5 order status enum).
var shansongStatusLabels = map[int]string{
	-1: "派单失败", // 派单调用失败（已付款但未成功下单，需运营介入）
	0:  "待派单", // 默认：已创建但尚未派单（不可显示为「配送中」掩盖真实状态）
	20: "派单中",
	30: "待取货",
	40: "闪送中",
	50: "已完成",
	60: "已取消",
}

// ShansongStatusLabel returns a human label for a delivery status code, falling
// back to a generic label for unknown codes (never empty, never panics).
func ShansongStatusLabel(code int) string {
	if label, ok := shansongStatusLabels[code]; ok {
		return label
	}
	return "配送中"
}
