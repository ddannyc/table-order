package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ShansongClient talks to the Shansong (闪送) open platform for instant-courier
// delivery: realtime price quote, order create, and cancel.
//
// CALIBRATION: the exact endpoint paths, the signing recipe, and the response
// field names below follow the documented Shansong OpenAPI conventions but MUST
// be verified against the official docs during 联调. They are deliberately
// isolated (signShansong + the path/field constants) so calibration is a
// localized edit; the request-signing and envelope-parsing mechanics — which the
// tests lock down — do not change.
type ShansongClient struct {
	ClientID  string
	AppSecret string
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
func InitShansongClient(clientID, appSecret, baseURL string) {
	if clientID == "" || appSecret == "" {
		return
	}
	Shansong = &ShansongClient{ClientID: clientID, AppSecret: appSecret, BaseURL: baseURL}
}

// Shansong OpenAPI paths (calibration point).
const (
	shansongPathQuote  = "/openapi/v3/order/precompute"
	shansongPathCreate = "/openapi/v3/order/create"
	shansongPathCancel = "/openapi/v3/order/cancel"
)

// QuoteRequest is the business input for a delivery price quote.
type QuoteRequest struct {
	SenderAddress    string
	SenderLat        float64
	SenderLng        float64
	RecipientName    string
	RecipientPhone   string
	RecipientAddress string
	RecipientLat     float64
	RecipientLng     float64
	WeightGram       int
}

// QuoteResult is the parsed quote: the fee plus Shansong's quote token, which is
// passed back at order create time so Shansong prices the same quote.
type QuoteResult struct {
	DeliveryFee float64
	QuoteToken  string
}

// CreateDeliveryRequest creates the actual dispatch against a prior quote.
type CreateDeliveryRequest struct {
	QuoteToken       string
	OrderNo          string // our order_no, for cross-reference
	SenderAddress    string
	SenderLat        float64
	SenderLng        float64
	RecipientName    string
	RecipientPhone   string
	RecipientAddress string
	RecipientLat     float64
	RecipientLng     float64
}

// shansongResp is the common Shansong response envelope: status==200 means OK.
type shansongResp struct {
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
	Data   json.RawMessage `json:"data"`
}

// signShansong builds the request signature. CALIBRATION: confirm the exact
// concatenation order and hash against the official docs. Documented recipe:
// uppercase(MD5(clientId + appSecret + timestamp + bizContent)).
func signShansong(clientID, appSecret, timestamp, bizContent string) string {
	raw := clientID + appSecret + timestamp + bizContent
	sum := md5.Sum([]byte(raw))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
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
	bizContent, err := json.Marshal(biz)
	if err != nil {
		return nil, fmt.Errorf("shansong: marshal biz: %w", err)
	}
	timestamp := strconv.FormatInt(c.now().UnixMilli(), 10)
	sign := signShansong(c.ClientID, c.AppSecret, timestamp, string(bizContent))

	payload := map[string]string{
		"clientId":   c.ClientID,
		"timestamp":  timestamp,
		"sign":       sign,
		"bizContent": string(bizContent),
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("shansong: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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

// CalculatePrice returns the realtime delivery fee + quote token.
func (c *ShansongClient) CalculatePrice(ctx context.Context, q QuoteRequest) (*QuoteResult, error) {
	biz := map[string]any{
		"senderAddress":    q.SenderAddress,
		"senderLat":        q.SenderLat,
		"senderLng":        q.SenderLng,
		"recipientName":    q.RecipientName,
		"recipientPhone":   q.RecipientPhone,
		"recipientAddress": q.RecipientAddress,
		"recipientLat":     q.RecipientLat,
		"recipientLng":     q.RecipientLng,
		"weight":           q.WeightGram,
	}
	env, err := c.post(ctx, shansongPathQuote, biz)
	if err != nil {
		return nil, err
	}
	var data struct {
		TotalFeeYuan float64 `json:"totalFeeYuan"` // CALIBRATION: confirm unit (元 vs 分) + field name
		OrderNumber  string  `json:"orderNumber"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("shansong: parse quote data: %w", err)
	}
	return &QuoteResult{DeliveryFee: data.TotalFeeYuan, QuoteToken: data.OrderNumber}, nil
}

// CreateOrder dispatches a courier against a prior quote, returning the Shansong order no.
func (c *ShansongClient) CreateOrder(ctx context.Context, r CreateDeliveryRequest) (string, error) {
	biz := map[string]any{
		"orderNumber":      r.QuoteToken, // the quote token from CalculatePrice
		"merchantOrderNo":  r.OrderNo,
		"senderAddress":    r.SenderAddress,
		"senderLat":        r.SenderLat,
		"senderLng":        r.SenderLng,
		"recipientName":    r.RecipientName,
		"recipientPhone":   r.RecipientPhone,
		"recipientAddress": r.RecipientAddress,
		"recipientLat":     r.RecipientLat,
		"recipientLng":     r.RecipientLng,
	}
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
	return data.OrderNumber, nil
}

// CancelOrder cancels a dispatched delivery. Wired but not exposed via a route
// this iteration (cancel/refund is out of scope per plan); kept for completeness.
func (c *ShansongClient) CancelOrder(ctx context.Context, shansongOrderNo string) error {
	biz := map[string]any{"orderNumber": shansongOrderNo}
	_, err := c.post(ctx, shansongPathCancel, biz)
	return err
}

// shansongStatusLabels maps Shansong delivery status codes to Chinese display
// text. CALIBRATION: confirm the actual code values against the callback docs.
var shansongStatusLabels = map[int]string{
	0:   "待派单",
	60:  "待接单",
	70:  "已接单",
	80:  "取货中",
	90:  "配送中",
	100: "已送达",
	1:   "已取消",
}

// ShansongStatusLabel returns a human label for a delivery status code, falling
// back to a generic label for unknown codes (never empty, never panics).
func ShansongStatusLabel(code int) string {
	if label, ok := shansongStatusLabels[code]; ok {
		return label
	}
	return "配送中"
}
