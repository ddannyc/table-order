package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
	"github.com/gin-gonic/gin"
)

// quoteTokenTTL bounds how long a delivery quote stays valid for order placement.
const quoteTokenTTL = 15 * time.Minute

// quoteClaims is the trusted payload the quote endpoint signs and the order
// endpoint verifies, so the client can never forge the delivery fee.
type quoteClaims struct {
	ShopID        uint    `json:"s"`
	Fee           float64 `json:"f"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	ShansongQuote string  `json:"q"` // shansong issOrderNo (from orderCalculate), replayed at orderPlace
	OrderNo       string  `json:"o"` // our order_no, minted at quote time (= shansong thirdOrderNo)
	Exp           int64   `json:"e"` // unix seconds
}

// quoteSecret returns the HMAC key for quote tokens. It fails closed: with no
// configured JWT secret it returns ok=false so signing/verification refuse,
// rather than falling back to a predictable constant that would let a client
// forge the delivery fee.
func quoteSecret() ([]byte, bool) {
	if config.AppConfig != nil && config.AppConfig.JWT.Secret != "" {
		return []byte(config.AppConfig.JWT.Secret), true
	}
	return nil, false
}

// signQuoteToken returns "<payload>.<hmac>" (both base64url). Returns "" when
// no quote secret is configured (DeliveryQuote guards this before calling).
func signQuoteToken(claims quoteClaims) string {
	secret, ok := quoteSecret()
	if !ok {
		return ""
	}
	payload, _ := json.Marshal(claims)
	b64 := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(b64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return b64 + "." + sig
}

// verifyQuoteToken checks the HMAC and expiry, returning the trusted claims.
func verifyQuoteToken(token string) (*quoteClaims, error) {
	secret, ok := quoteSecret()
	if !ok {
		return nil, errors.New("quote signing not configured")
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, errors.New("malformed quote token")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, errors.New("bad quote signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("bad quote payload")
	}
	var claims quoteClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.New("bad quote payload")
	}
	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("quote expired")
	}
	if claims.Fee < 0 {
		return nil, errors.New("invalid quote fee")
	}
	return &claims, nil
}

type DeliveryQuoteRequest struct {
	ShopID           uint    `json:"shop_id" binding:"required"`
	RecipientName    string  `json:"recipient_name"`
	RecipientPhone   string  `json:"recipient_phone"`
	RecipientAddress string  `json:"recipient_address"`
	RecipientLat     float64 `json:"recipient_lat"`
	RecipientLng     float64 `json:"recipient_lng"`
}

// DeliveryQuote returns a realtime Shansong delivery fee plus a signed quote
// token. The fee is trusted only via the token at order time — never from the
// client. Destination coords come from wx.getLocation on the client.
func DeliveryQuote(c *gin.Context) {
	var req DeliveryQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.RecipientLat == 0 || req.RecipientLng == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "收件坐标缺失，请重新定位"})
		return
	}

	var shop models.Shop
	if err := config.DB.First(&shop, req.ShopID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}
	if shop.Latitude == 0 || shop.Longitude == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "门店未配置坐标，暂不支持外卖"})
		return
	}
	if shop.City == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "门店未配置城市，暂不支持外卖"})
		return
	}

	if services.Shansong == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "外卖配送暂未开通"})
		return
	}
	if _, ok := quoteSecret(); !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "外卖配送暂未开通"})
		return
	}

	// Mint the order_no now so it is the third-party orderNo Shansong records at
	// quote time; CreateOrder reuses it as the order's order_no.
	orderNo := generateOrderNo()

	res, err := services.Shansong.CalculatePrice(c.Request.Context(), services.QuoteRequest{
		CityName:         shop.City,
		SenderAddress:    shop.Address,
		SenderName:       shop.Name,
		SenderMobile:     shop.Phone,
		SenderLat:        shop.Latitude,
		SenderLng:        shop.Longitude,
		ThirdOrderNo:     orderNo,
		RecipientName:    req.RecipientName,
		RecipientPhone:   req.RecipientPhone,
		RecipientAddress: req.RecipientAddress,
		RecipientLat:     req.RecipientLat,
		RecipientLng:     req.RecipientLng,
	})
	if err != nil {
		// Surface the real Shansong failure (transport/range/sign/balance) — the
		// client message stays generic, but ops needs the cause in the logs.
		log.Printf("[shansong] quote failed shopID=%d city=%q sender=(%f,%f) recipient=(%f,%f): %v",
			req.ShopID, shop.City, shop.Latitude, shop.Longitude, req.RecipientLat, req.RecipientLng, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "配送报价失败，可能超出配送范围"})
		return
	}

	token := signQuoteToken(quoteClaims{
		ShopID:        req.ShopID,
		Fee:           res.DeliveryFee,
		Lat:           req.RecipientLat,
		Lng:           req.RecipientLng,
		ShansongQuote: res.QuoteToken,
		OrderNo:       orderNo,
		Exp:           time.Now().Add(quoteTokenTTL).Unix(),
	})

	c.JSON(http.StatusOK, gin.H{
		"delivery_fee": res.DeliveryFee,
		"quote_token":  token,
	})
}
