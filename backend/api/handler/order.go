package handler

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
	"github.com/example/table-order/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderItemRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	ShopID  uint              `json:"shop_id" binding:"required"`
	TableNo string            `json:"table_no" binding:"required"`
	Amount  float64           `json:"amount" binding:"required,gt=0"`
	Items   []OrderItemRequest `json:"items"`
	UseReward bool            `json:"use_reward"`
}

func CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var shop models.Shop
	if err := config.DB.First(&shop, req.ShopID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	var calculatedAmount float64
	var productMap map[uint]models.Product

	if len(req.Items) > 0 {
		productIDs := make([]uint, len(req.Items))
		quantityMap := map[uint]int{}
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
			quantityMap[item.ProductID] = item.Quantity
		}

		var products []models.Product
		config.DB.Where("id IN ? AND shop_id = ?", productIDs, req.ShopID).Find(&products)
		if len(products) != len(req.Items) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid products"})
			return
		}

		productMap = make(map[uint]models.Product, len(products))
		for _, p := range products {
			if p.Status != 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product %s not available", p.Name)})
				return
			}
			productMap[p.ID] = p
			calculatedAmount += p.Price * float64(quantityMap[p.ID])
		}

		req.Amount = calculatedAmount
	}

	// Get user to check if referred
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Order placer does NOT receive self-reward (only inviter receives invite reward via processInviteReward)
	// Handle reward balance deduction for payment
	var deductAmount float64
	maxDeductAmount := req.Amount * shop.RewardCeiling
	if req.UseReward && user.RewardBalance > 0 {
		deductAmount = math.Min(user.RewardBalance, maxDeductAmount)
		req.Amount = req.Amount - deductAmount
	}

	orderNo := fmt.Sprintf("%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8])

	now := time.Now()
	order := models.Order{
		OrderNo:      orderNo,
		UserID:       userID,
		ShopID:       req.ShopID,
		TableNo:      req.TableNo,
		Amount:       req.Amount,
		RewardAmount: 0,
		Status:       2,
		PaidAt:       &now,
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction start failed"})
		return
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create order failed"})
		return
	}

	for _, item := range req.Items {
		product := productMap[item.ProductID]
		orderItem := models.OrderItem{
			OrderID:     order.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    item.Quantity,
			Subtotal:    product.Price * float64(item.Quantity),
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create order item failed"})
			return
		}
	}

	// Deduct reward balance if used for payment
	if req.UseReward && deductAmount > 0 {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&models.User{}).Where("id = ?", userID).
			UpdateColumn("reward_balance", gorm.Expr("reward_balance - ?", deductAmount)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "deduct reward failed"})
			return
		}
		deductLog := models.WalletLog{
			UserID:  userID,
			Type:    "deduct",
			Amount:  -deductAmount,
			OrderID: &order.ID,
			Remark:  fmt.Sprintf("使用福利金抵扣 %.2f 元", deductAmount),
		}
		if err := tx.Create(&deductLog).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create deduct log failed"})
			return
		}
	}

	// Deduct remaining amount from regular balance
	if req.Amount > 0 {
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&models.User{}).
			Where("id = ? AND balance >= ?", userID, req.Amount).
			UpdateColumn("balance", gorm.Expr("balance - ?", req.Amount))
		if result.Error != nil || result.RowsAffected == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
			return
		}
		payLog := models.WalletLog{
			UserID:  userID,
			Type:    "pay",
			Amount:  -req.Amount,
			OrderID: &order.ID,
			Remark:  fmt.Sprintf("订单支付 %.2f 元", req.Amount),
		}
		if err := tx.Create(&payLog).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create pay log failed"})
			return
		}
	}

	tx.Commit()

	// Update last_consume_at and resume reward
	config.DB.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_consume_at":  time.Now(),
			"reward_paused_at": nil,
		})

	// Distribute 3-tier reward asynchronously
	go services.DistributeReward(order.ID, userID, req.ShopID, req.Amount)

	c.JSON(http.StatusOK, gin.H{
		"id":            order.ID,
		"order_no":      order.OrderNo,
		"amount":        order.Amount,
		"reward_amount": order.RewardAmount,
		"status":        order.Status,
	})
}

func GetOrders(c *gin.Context) {
	userID := c.GetUint("user_id")

	page := uint(1)
	pageSize := uint(20)
	if p := c.Query("page"); p != "" {
		if parsed := utils.ParseUint(p); parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed := utils.ParseUint(ps); parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	var orders []models.Order
	var total int64

	config.DB.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total)
	config.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(int(offset)).Limit(int(pageSize)).Find(&orders)

	// Fetch order items for each order
	type orderWithItems struct {
		models.Order
		Items []models.OrderItem `json:"items"`
	}
	result := make([]orderWithItems, len(orders))
	for i, o := range orders {
		var items []models.OrderItem
		config.DB.Where("order_id = ?", o.ID).Find(&items)
		result[i] = orderWithItems{Order: o, Items: items}
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":    result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID := c.Param("id")

	var order models.Order
	if err := config.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	var items []models.OrderItem
	config.DB.Where("order_id = ?", orderID).Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"order": order,
		"items": items,
	})
}

func processInviteReward(userID uint, shopID uint, amount float64) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		log.Printf("[invite_reward] user not found: userID=%d err=%v", userID, err)
		return
	}
	if user.InviterID == nil {
		return
	}

	inviterID := *user.InviterID

	var shop models.Shop
	if err := config.DB.First(&shop, shopID).Error; err != nil {
		log.Printf("[invite_reward] shop not found: shopID=%d err=%v", shopID, err)
		return
	}

	inviteReward := amount * shop.RewardRateLevel1
	if inviteReward <= 0 {
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		log.Printf("[invite_reward] tx begin failed: err=%v", tx.Error)
		return
	}

	if err := tx.Model(&models.User{}).Where("id = ?", inviterID).
		UpdateColumn("reward_balance", gorm.Expr("reward_balance + ?", inviteReward)).Error; err != nil {
		tx.Rollback()
		log.Printf("[invite_reward] update reward_balance failed: inviterID=%d amount=%.2f err=%v", inviterID, inviteReward, err)
		return
	}

	walletLog := models.WalletLog{
		UserID:  inviterID,
		Type:    "invite_reward",
		Amount:  inviteReward,
		Remark:  fmt.Sprintf("邀请奖励 %.2f 元", inviteReward),
	}
	if err := tx.Create(&walletLog).Error; err != nil {
		tx.Rollback()
		log.Printf("[invite_reward] create wallet_log failed: inviterID=%d err=%v", inviterID, err)
		return
	}

	tx.Commit()
	log.Printf("[invite_reward] rewarded: inviterID=%d inviteeID=%d amount=%.2f", inviterID, userID, inviteReward)
}