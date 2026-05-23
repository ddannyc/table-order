package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

type MerchantOrdersResponse struct {
	Orders   []models.Order `json:"orders"`
	Total    int64          `json:"total"`
	Revenue  float64        `json:"revenue"`
	Rewarded float64        `json:"rewarded"`
}

func GetMerchantOrders(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	// Get shops for this merchant
	var shops []models.Shop
	config.DB.Where("merchant_id = ?", merchantID).Find(&shops)
	var shopIDs []uint
	for _, s := range shops {
		shopIDs = append(shopIDs, s.ID)
	}

	// Filters
	shopID := c.Query("shop_id")
	date := c.Query("date")

	query := config.DB.Model(&models.Order{}).Where("shop_id IN ?", shopIDs)

	if shopID != "" {
		query = query.Where("shop_id = ?", shopID)
	}
	if date != "" {
		t, _ := time.Parse("2006-01-02", date)
		query = query.Where("DATE(created_at) = ?", t)
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	query.Order("created_at desc").Limit(50).Find(&orders)

	// Calculate totals
	var revenue, rewarded float64
	for _, o := range orders {
		revenue += o.Amount
		rewarded += o.RewardAmount
	}

	c.JSON(http.StatusOK, MerchantOrdersResponse{
		Orders:   orders,
		Total:    total,
		Revenue:  revenue,
		Rewarded: rewarded,
	})
}

type CreateMerchantOrderRequest struct {
	UserID   uint    `json:"user_id" binding:"required"`
	ShopID   uint    `json:"shop_id" binding:"required"`
	TableNo  string  `json:"table_no"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

func CreateMerchantOrder(c *gin.Context) {
	merchantID := c.GetUint("user_id")
	shopIDStr := c.Query("shop_id")

	shopID, _ := strconv.ParseUint(shopIDStr, 10, 32)
	if shopID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop_id required"})
		return
	}

	// Verify shop belongs to merchant
	var shop models.Shop
	if err := config.DB.Where("id = ? AND merchant_id = ?", shopID, merchantID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "shop not found"})
		return
	}

	var req CreateMerchantOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	order := models.Order{
		OrderNo:      generateOrderNo(),
		UserID:       req.UserID,
		ShopID:       uint(shopID),
		TableNo:      req.TableNo,
		Amount:       req.Amount,
		RewardAmount: 0,
		Status:       2,
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create order failed"})
		return
	}

	// Credit reward to inviter (not to order placer per spec)
	processInviteReward(req.UserID, uint(shopID), req.Amount)

	c.JSON(http.StatusOK, order)
}

func generateOrderNo() string {
	return time.Now().Format("20060102150405") + strconv.FormatInt(time.Now().UnixNano()%100000000, 10)
}