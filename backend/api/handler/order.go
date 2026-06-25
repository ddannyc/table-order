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
	ProductID uint `json:"product_id" binding:"required"`
	SpecID    uint `json:"spec_id"` // 0 = no spec (single default spec at product price)
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	ShopID    uint               `json:"shop_id" binding:"required"`
	OrderType string             `json:"order_type"` // dine_in (default) | delivery
	TableNo   string             `json:"table_no"`   // required for dine_in; empty allowed for delivery
	Amount    float64            `json:"amount" binding:"required,gt=0"`
	Items     []OrderItemRequest `json:"items"`
	UseReward bool               `json:"use_reward"`
}

func CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	orderType := req.OrderType
	if orderType == "" {
		orderType = "dine_in"
	}
	if orderType != "delivery" && req.TableNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_no required for dine-in"})
		return
	}

	var shop models.Shop
	if err := config.DB.First(&shop, req.ShopID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	// Resolve each line to a product (+ optional spec) and price it server-side.
	type resolvedLine struct {
		product  models.Product
		specID   uint
		specName string
		price    float64
		quantity int
	}
	var lines []resolvedLine
	var calculatedAmount float64

	if len(req.Items) > 0 {
		productCache := map[uint]models.Product{}
		for _, item := range req.Items {
			product, cached := productCache[item.ProductID]
			if !cached {
				if err := config.DB.Where("id = ? AND shop_id = ?", item.ProductID, req.ShopID).First(&product).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid products"})
					return
				}
				productCache[item.ProductID] = product
			}
			if product.Status != 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product %s not available", product.Name)})
				return
			}

			price := product.Price
			var specID uint
			var specName string
			if item.SpecID != 0 {
				var spec models.ProductSpec
				if err := config.DB.Where("id = ? AND product_id = ?", item.SpecID, item.ProductID).First(&spec).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid spec"})
					return
				}
				if spec.Status != 1 {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("spec %s not available", spec.Name)})
					return
				}
				price = spec.Price
				specID = spec.ID
				specName = spec.Name
			}

			calculatedAmount += price * float64(item.Quantity)
			lines = append(lines, resolvedLine{product: product, specID: specID, specName: specName, price: price, quantity: item.Quantity})
		}

		req.Amount = calculatedAmount
	}

	// Get user (for openid and reward balance)
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Handle reward balance deduction for discount
	var deductAmount float64
	maxDeductAmount := req.Amount * shop.RewardCeiling
	if req.UseReward && user.RewardBalance > 0 {
		deductAmount = math.Min(user.RewardBalance, maxDeductAmount)
		req.Amount = req.Amount - deductAmount
	}

	orderNo := fmt.Sprintf("%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8])

	order := models.Order{
		OrderNo:      orderNo,
		UserID:       userID,
		ShopID:       req.ShopID,
		OrderType:    orderType,
		TableNo:      req.TableNo,
		Amount:       req.Amount,
		RewardAmount: 0,
		Status:       1, // pending — WeChat Pay will confirm
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

	for _, ln := range lines {
		orderItem := models.OrderItem{
			OrderID:     order.ID,
			ProductID:   ln.product.ID,
			ProductName: ln.product.Name,
			SpecID:      ln.specID,
			SpecName:    ln.specName,
			Price:       ln.price,
			Quantity:    ln.quantity,
			Subtotal:    ln.price * float64(ln.quantity),
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create order item failed"})
			return
		}
	}

	// Deduct reward balance if used for discount (happens at order creation)
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

	tx.Commit()

	// Zero-amount order (reward covered everything): mark paid directly, skip WeChat Pay
	if req.Amount <= 0 {
		now := time.Now()
		config.DB.Model(&order).Updates(map[string]interface{}{
			"status":  2,
			"paid_at": &now,
		})
		config.DB.Model(&models.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{
				"last_consume_at":  now,
				"reward_paused_at": nil,
			})
		go services.DistributeReward(order.ID, userID, req.ShopID, req.Amount)
		c.JSON(http.StatusOK, gin.H{
			"id":       order.ID,
			"order_no": order.OrderNo,
			"amount":   order.Amount,
			"status":   2,
		})
		return
	}

	// Call WeChat Pay JSAPI prepay
	amountCents := int64(math.Round(req.Amount * 100))
	description := fmt.Sprintf("%s-订单%s", shop.Name, orderNo)
	prepay, err := services.CreateJSAPIPrepay(c.Request.Context(), user.OpenID, orderNo, description, amountCents)
	if err != nil {
		log.Printf("[order] wechat prepay failed for order %s: %v", orderNo, err)
		c.JSON(http.StatusOK, gin.H{
			"id":       order.ID,
			"order_no": order.OrderNo,
			"amount":   order.Amount,
			"status":   order.Status,
			"error":    "prepay failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         order.ID,
		"order_no":   order.OrderNo,
		"amount":     order.Amount,
		"status":     order.Status,
		"prepay_id":  prepay.PrepayID,
		"time_stamp": prepay.TimeStamp,
		"nonce_str":  prepay.NonceStr,
		"package":    prepay.Package,
		"sign_type":  prepay.SignType,
		"pay_sign":   prepay.PaySign,
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