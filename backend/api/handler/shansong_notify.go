package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
	"github.com/gin-gonic/gin"
)

type shansongCallbackRequest struct {
	ClientID   string `json:"clientId"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
	BizContent string `json:"bizContent"`
}

// ShansongCallback receives Shansong delivery status updates. It verifies the
// signature, updates the delivery status, and MUST return {"status":200} on
// success (otherwise Shansong retries). Idempotent: re-applying the same status
// is harmless.
func ShansongCallback(c *gin.Context) {
	var cb shansongCallbackRequest
	if err := c.ShouldBindJSON(&cb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "bad request"})
		return
	}

	if services.Shansong == nil {
		// Not configured — can't verify. Ack to avoid a retry storm; log loudly.
		log.Printf("[shansong] callback received but client not configured")
		c.JSON(http.StatusOK, gin.H{"status": 200})
		return
	}
	if !services.Shansong.VerifyCallback(cb.Timestamp, cb.BizContent, cb.Sign) {
		log.Printf("[shansong] callback signature verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "invalid sign"})
		return
	}

	var biz struct {
		OrderNumber string `json:"orderNumber"`
		Status      int    `json:"status"`
	}
	if err := json.Unmarshal([]byte(cb.BizContent), &biz); err != nil || biz.OrderNumber == "" {
		log.Printf("[shansong] callback bizContent unparseable: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "bad bizContent"})
		return
	}

	res := config.DB.Model(&models.OrderDelivery{}).
		Where("shansong_order_no = ?", biz.OrderNumber).
		Update("shansong_status", biz.Status)
	if res.Error != nil {
		log.Printf("[shansong] callback update failed shansongNo=%s: %v", biz.OrderNumber, res.Error)
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": "update failed"})
		return
	}
	if res.RowsAffected == 0 {
		// Unknown order number — ack anyway so Shansong stops retrying; log it.
		log.Printf("[shansong] callback for unknown shansongNo=%s", biz.OrderNumber)
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}
