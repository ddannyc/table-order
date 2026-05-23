package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

type StatsResponse struct {
	NewUsers   int64   `json:"new_users"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
	Rewarded   float64 `json:"rewarded"`
}

func GetMerchantStats(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	dateStr := c.Query("date")
	shopIDStr := c.Query("shop_id")

	var date time.Time
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	} else {
		date = time.Now()
	}

	// Get merchant's shops
	query := config.DB.Where("merchant_id = ?", merchantID)
	if shopIDStr != "" {
		query = query.Where("id = ?", shopIDStr)
	}

	var shops []models.Shop
	query.Find(&shops)
	var shopIDs []uint
	for _, s := range shops {
		shopIDs = append(shopIDs, s.ID)
	}

	// Count new users who placed orders at merchant's shops today
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var newUsers int64
	config.DB.Model(&models.User{}).
		Joins("JOIN orders ON orders.user_id = users.id").
		Where("orders.shop_id IN ? AND users.created_at >= ? AND users.created_at < ?", shopIDs, startOfDay, endOfDay).
		Distinct("users.id").Count(&newUsers)

	// Count orders
	var orders int64
	config.DB.Model(&models.Order{}).
		Where("shop_id IN ? AND created_at >= ? AND created_at < ?", shopIDs, startOfDay, endOfDay).
		Count(&orders)

	// Sum revenue
	var revenue, rewarded float64
	config.DB.Model(&models.Order{}).
		Where("shop_id IN ? AND created_at >= ? AND created_at < ?", shopIDs, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0), COALESCE(SUM(reward_amount), 0)").
		Row().Scan(&revenue, &rewarded)

	c.JSON(http.StatusOK, StatsResponse{
		NewUsers: newUsers,
		Orders:   orders,
		Revenue:  revenue,
		Rewarded: rewarded,
	})
}

func GetMerchantDashboard(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	// Get all shops
	var shops []models.Shop
	config.DB.Where("merchant_id = ?", merchantID).Find(&shops)

	// Total stats
	var totalUsers, totalOrders int64
	var totalRevenue float64

	config.DB.Model(&models.User{}).Joins("JOIN orders ON orders.user_id = users.id").
		Where("orders.shop_id IN ?", shopIDs(shops)).Distinct("users.id").Count(&totalUsers)

	config.DB.Model(&models.Order{}).Where("shop_id IN ?", shopIDs(shops)).Count(&totalOrders)
	config.DB.Model(&models.Order{}).Where("shop_id IN ?", shopIDs(shops)).Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)

	c.JSON(http.StatusOK, gin.H{
		"shops":        shops,
		"total_users":  totalUsers,
		"total_orders": totalOrders,
		"total_revenue": totalRevenue,
	})
}

func shopIDs(shops []models.Shop) []uint {
	var ids []uint
	for _, s := range shops {
		ids = append(ids, s.ID)
	}
	return ids
}

func GetAllMerchants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var merchants []models.Merchant
	var total int64

	config.DB.Model(&models.Merchant{}).Count(&total)
	config.DB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&merchants)

	c.JSON(http.StatusOK, gin.H{
		"merchants": merchants,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}