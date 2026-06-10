package handler

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
)

func WechatPayNotify(c *gin.Context) {
	cfg := config.AppConfig.WeChat

	// WeChat Pay not configured — should never receive callbacks, but guard anyway
	if cfg.MchID == "" || config.WxPayClient == nil {
		log.Printf("[wechatpay] not configured, skipping notify")
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "not configured"})
		return
	}

	var handler *notify.Handler
	if cfg.WechatPayPublicKeyID != "" {
		var wechatPayPublicKey *rsa.PublicKey
		var err error
		if cfg.WechatPayPublicKeyContent != "" {
			wechatPayPublicKey, err = utils.LoadPublicKey(cfg.WechatPayPublicKeyContent)
		} else {
			wechatPayPublicKey, err = utils.LoadPublicKeyWithPath(cfg.WechatPayPublicKeyPath)
		}
		if err != nil {
			log.Printf("[wechatpay] FATAL: load public key for notify failed: %v", err)
			c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "internal config error"})
			return
		}
		handler = notify.NewNotifyHandler(cfg.MchAPIv3Key, verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.WechatPayPublicKeyID, *wechatPayPublicKey))
	} else {
		certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
		handler = notify.NewNotifyHandler(cfg.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	}

	transaction := new(payments.Transaction)
	_, err := handler.ParseNotifyRequest(context.Background(), c.Request, transaction)
	if err != nil {
		log.Printf("[wechatpay] notify parse failed: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "parse failed"})
		return
	}

	if transaction.OutTradeNo == nil {
		log.Printf("[wechatpay] missing out_trade_no in notify")
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "missing out_trade_no"})
		return
	}

	var order models.Order
	if err := config.DB.Where("order_no = ?", *transaction.OutTradeNo).First(&order).Error; err != nil {
		log.Printf("[wechatpay] order not found: %s", *transaction.OutTradeNo)
		c.JSON(http.StatusOK, gin.H{"code": "FAIL"})
		return
	}

	// Idempotent: skip if already paid or cancelled
	if order.Status != 1 {
		log.Printf("[wechatpay] order %s status=%d — skipping notify", order.OrderNo, order.Status)
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS"})
		return
	}

	// Only process SUCCESS trade state
	if transaction.TradeState == nil || *transaction.TradeState != "SUCCESS" {
		log.Printf("[wechatpay] trade state not success: %v for order %s", transaction.TradeState, order.OrderNo)
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS"})
		return
	}

	now := time.Now()
	if err := config.DB.Model(&order).Updates(map[string]interface{}{
		"status":  2,
		"paid_at": &now,
	}).Error; err != nil {
		log.Printf("[wechatpay] update order %s failed: %v", order.OrderNo, err)
		c.JSON(http.StatusOK, gin.H{"code": "FAIL"})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", order.UserID).
		Updates(map[string]interface{}{
			"last_consume_at":  time.Now(),
			"reward_paused_at": nil,
		}).Error; err != nil {
		log.Printf("[wechatpay] update user consumption tracking failed: userID=%d err=%v", order.UserID, err)
	}

	go services.DistributeReward(order.ID, order.UserID, order.ShopID, order.Amount)

	log.Printf("[wechatpay] order %s paid, transaction_id=%s", order.OrderNo, *transaction.TransactionId)
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS"})
}
