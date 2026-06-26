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

// ShansongCallback receives Shansong delivery status updates. Per the platform
// protocol the callback is form-urlencoded with clientId/shopId/timestamp/sign
// and a `data` JSON business param. It verifies the signature, updates the
// delivery status, and MUST return {"status":200} on success (else Shansong
// retries). Idempotent: re-applying the same status is harmless.
//
// CALIBRATION: confirm the inbound payload (form vs body, exact data field
// names) against the official callback doc during 联调.
func ShansongCallback(c *gin.Context) {
	timestamp := c.PostForm("timestamp")
	sign := c.PostForm("sign")
	data := c.PostForm("data")

	if services.Shansong == nil {
		// Not configured — can't verify. Ack to avoid a retry storm; log loudly.
		log.Printf("[shansong] callback received but client not configured")
		c.JSON(http.StatusOK, gin.H{"status": 200})
		return
	}
	if !services.Shansong.VerifyCallback(timestamp, data, sign) {
		log.Printf("[shansong] callback signature verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "invalid sign"})
		return
	}

	var biz struct {
		IssOrderNo  string `json:"issOrderNo"`
		OrderStatus int    `json:"orderStatus"`
	}
	if err := json.Unmarshal([]byte(data), &biz); err != nil || biz.IssOrderNo == "" {
		log.Printf("[shansong] callback data unparseable: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": 400, "msg": "bad data"})
		return
	}

	// Match on either the dispatched order no or the quote no (both hold the
	// shansong issOrderNo depending on what orderPlace echoed). Never regress a
	// terminal status (50 已完成 / 60 已取消): a late/out-of-order callback that
	// would move it backward is ignored (RowsAffected==0) but still acked.
	res := config.DB.Model(&models.OrderDelivery{}).
		Where("(shansong_order_no = ? OR shansong_quote_no = ?) AND shansong_status NOT IN (?)",
			biz.IssOrderNo, biz.IssOrderNo, []int{50, 60}).
		Update("shansong_status", biz.OrderStatus)
	if res.Error != nil {
		log.Printf("[shansong] callback update failed issOrderNo=%s: %v", biz.IssOrderNo, res.Error)
		c.JSON(http.StatusOK, gin.H{"status": 500, "msg": "update failed"})
		return
	}
	if res.RowsAffected == 0 {
		log.Printf("[shansong] callback for unknown issOrderNo=%s", biz.IssOrderNo)
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
}
