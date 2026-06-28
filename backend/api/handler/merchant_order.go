package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"gorm.io/gorm"
)

const (
	defaultOrderPageSize = 20
	maxOrderPageSize     = 100
)

// MerchantOrderItem is an order plus its delivery detail (nil for dine-in).
type MerchantOrderItem struct {
	models.Order
	Delivery *models.OrderDelivery `json:"delivery"`
}

type MerchantOrdersResponse struct {
	Orders   []MerchantOrderItem `json:"orders"`
	Total    int64               `json:"total"`
	Revenue  float64             `json:"revenue"`
	Rewarded float64             `json:"rewarded"`
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
	if len(shopIDs) == 0 {
		c.JSON(http.StatusOK, MerchantOrdersResponse{Orders: []MerchantOrderItem{}})
		return
	}

	// filtered returns a fresh query with all filters applied, so Count/aggregate/
	// list can each run independently without leaking statement state into each other.
	filtered := func() *gorm.DB {
		q := config.DB.Model(&models.Order{}).Where("shop_id IN ?", shopIDs)
		if shopID := c.Query("shop_id"); shopID != "" {
			q = q.Where("shop_id = ?", shopID)
		}
		if date := c.Query("date"); date != "" {
			if t, err := time.Parse("2006-01-02", date); err == nil {
				q = q.Where("DATE(created_at) = ?", t)
			}
		}
		if status := c.Query("status"); status != "" {
			q = q.Where("status = ?", status)
		}
		if typ := c.Query("type"); typ != "" {
			q = q.Where("order_type = ?", typ)
		}
		return q
	}

	// Totals reflect the full filtered set, not just the current page.
	var total int64
	filtered().Count(&total)
	var agg struct {
		Revenue  float64
		Rewarded float64
	}
	filtered().Select("COALESCE(SUM(amount),0) AS revenue, COALESCE(SUM(reward_amount),0) AS rewarded").Scan(&agg)

	// Pagination
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 {
		pageSize = defaultOrderPageSize
	}
	if pageSize > maxOrderPageSize {
		pageSize = maxOrderPageSize
	}

	var orders []models.Order
	filtered().Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&orders)

	// Embed delivery detail for the page (batch-loaded, nil for dine-in).
	items := make([]MerchantOrderItem, len(orders))
	orderIDs := make([]uint, len(orders))
	for i, o := range orders {
		items[i] = MerchantOrderItem{Order: o}
		orderIDs[i] = o.ID
	}
	if len(orderIDs) > 0 {
		var deliveries []models.OrderDelivery
		config.DB.Where("order_id IN ?", orderIDs).Find(&deliveries)
		byOrderID := make(map[uint]*models.OrderDelivery, len(deliveries))
		for i := range deliveries {
			byOrderID[deliveries[i].OrderID] = &deliveries[i]
		}
		for i := range items {
			if d, ok := byOrderID[items[i].Order.ID]; ok {
				items[i].Delivery = d
			}
		}
	}

	c.JSON(http.StatusOK, MerchantOrdersResponse{
		Orders:   items,
		Total:    total,
		Revenue:  agg.Revenue,
		Rewarded: agg.Rewarded,
	})
}

// loadOwnedOrder loads the order at :id and verifies it belongs to one of the
// merchant's shops. On failure it writes the response (404/403) and returns ok=false.
func loadOwnedOrder(c *gin.Context, merchantID uint) (*models.Order, bool) {
	var order models.Order
	if err := config.DB.First(&order, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return nil, false
	}
	var shop models.Shop
	if err := config.DB.Where("id = ? AND merchant_id = ?", order.ShopID, merchantID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return nil, false
	}
	return &order, true
}

// PrepareOrder marks an order as 出餐 (food ready) by setting PreparedAt.
// Idempotent: re-calling on an already-prepared order succeeds without moving the time.
func PrepareOrder(c *gin.Context) {
	merchantID := c.GetUint("user_id")
	order, ok := loadOwnedOrder(c, merchantID)
	if !ok {
		return
	}
	if order.PreparedAt == nil {
		now := time.Now()
		if err := config.DB.Model(order).Update("prepared_at", now).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "prepare failed"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "prepared"})
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