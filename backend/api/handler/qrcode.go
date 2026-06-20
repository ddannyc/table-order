package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/utils"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type GenerateQRRequest struct {
	TableNo string `json:"table_no" binding:"required"`
}

// GenerateMerchantQR issues a table QR code for a shop owned by the calling merchant.
func GenerateMerchantQR(c *gin.Context) {
	shopID := c.Param("id")
	if !merchantOwnsShop(shopID, c.GetUint("user_id")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req GenerateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_no required"})
		return
	}

	// Generate unique token for this QR
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	qr := models.TableQRCode{
		ShopID:  utils.ParseUint(shopID),
		TableNo: req.TableNo,
		Token:   token,
		Status:  1,
	}
	if err := config.DB.Create(&qr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create qr failed"})
		return
	}

	// Generate QR code image
	baseURL := config.AppConfig.Server.BaseURL
	if baseURL == "" {
		baseURL = "https://domain.com"
	}
	qrURL := fmt.Sprintf("%s/scan?shop_id=%s&table_no=%s&token=%s", baseURL, shopID, req.TableNo, token)
	png, err := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate qr image failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       qr.ID,
		"shop_id":  qr.ShopID,
		"table_no": qr.TableNo,
		"token":    qr.Token,
		"qr_image": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	})
}

// ListMerchantQRCodes lists QR codes for a shop owned by the calling merchant.
func ListMerchantQRCodes(c *gin.Context) {
	shopID := c.Param("id")
	if !merchantOwnsShop(shopID, c.GetUint("user_id")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var qrcodes []models.TableQRCode
	config.DB.Where("shop_id = ?", shopID).Find(&qrcodes)
	c.JSON(http.StatusOK, qrcodes)
}
