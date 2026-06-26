package handler

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
	"github.com/example/table-order/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	SpecID    uint `json:"spec_id"` // 0 = no spec (single default spec at product price)
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type DeliveryInfo struct {
	RecipientName  string  `json:"recipient_name"`
	RecipientPhone string  `json:"recipient_phone"`
	Province       string  `json:"province"`
	City           string  `json:"city"`
	County         string  `json:"county"`
	DetailAddress  string  `json:"detail_address"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
}

type CreateOrderRequest struct {
	ShopID     uint               `json:"shop_id" binding:"required"`
	OrderType  string             `json:"order_type"` // dine_in (default) | delivery
	TableNo    string             `json:"table_no"`   // required for dine_in; empty allowed for delivery
	Amount     float64            `json:"amount" binding:"required,gt=0"`
	Items      []OrderItemRequest `json:"items"`
	UseReward  bool               `json:"use_reward"`
	Delivery   *DeliveryInfo      `json:"delivery"`    // required for delivery orders
	QuoteToken string             `json:"quote_token"` // signed delivery fee token from /delivery/quote
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
	if orderType != "dine_in" && orderType != "delivery" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order_type"})
		return
	}
	if orderType != "delivery" && req.TableNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_no required for dine-in"})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items required"})
		return
	}

	var shop models.Shop
	if err := config.DB.First(&shop, req.ShopID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	// Delivery orders carry a signed quote token; the delivery fee is trusted
	// only from the token, never from the client (mirrors server-side pricing).
	var deliveryFee float64
	var deliveryClaims *quoteClaims
	if orderType == "delivery" {
		if req.Delivery == nil || req.QuoteToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "delivery info required"})
			return
		}
		claims, err := verifyQuoteToken(req.QuoteToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired quote"})
			return
		}
		if claims.ShopID != req.ShopID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quote shop mismatch"})
			return
		}
		deliveryFee = claims.Fee
		deliveryClaims = claims
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

	// items is guaranteed non-empty (rejected above); always price server-side.
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
		} else {
			// No spec chosen — only allowed for plain products. A product is
			// spec-priced if it has any non-下架 spec (上架=1 or 售罄=2), so a
			// bare spec_id:0 would underpay at the base price; reject it.
			// 下架=0 specs are excluded: the merchant removed those variants, so
			// the product reverts to base-price ordering.
			// Fail closed on a DB error rather than skipping the guard.
			var specCount int64
			if err := config.DB.Model(&models.ProductSpec{}).
				Where("product_id = ? AND status <> 0", item.ProductID).Count(&specCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "spec lookup failed"})
				return
			}
			if specCount > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("spec required for %s", product.Name)})
				return
			}
		}

		calculatedAmount += price * float64(item.Quantity)
		lines = append(lines, resolvedLine{product: product, specID: specID, specName: specName, price: price, quantity: item.Quantity})
	}

	req.Amount = calculatedAmount

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

	// Persist delivery detail (1:1) for delivery orders. Coords come from the
	// signed token (authoritative), address text from the request.
	if orderType == "delivery" {
		od := models.OrderDelivery{
			OrderID:         order.ID,
			RecipientName:   req.Delivery.RecipientName,
			RecipientPhone:  req.Delivery.RecipientPhone,
			Province:        req.Delivery.Province,
			City:            req.Delivery.City,
			County:          req.Delivery.County,
			DetailAddress:   req.Delivery.DetailAddress,
			RecipientLat:    deliveryClaims.Lat,
			RecipientLng:    deliveryClaims.Lng,
			DeliveryFee:     deliveryFee,
			ShansongQuoteNo: deliveryClaims.ShansongQuote,
		}
		if err := tx.Create(&od).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create delivery failed"})
			return
		}
	}

	// Deduct reward balance if used for discount (happens at order creation).
	// Guarded conditional update: the WHERE rejects the write if the balance was
	// drained by a concurrent order between the read above and here, so the
	// balance can never go negative (no double-spend).
	if req.UseReward && deductAmount > 0 {
		res := tx.Model(&models.User{}).Where("id = ? AND reward_balance >= ?", userID, deductAmount).
			UpdateColumn("reward_balance", gorm.Expr("reward_balance - ?", deductAmount))
		if res.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "deduct reward failed"})
			return
		}
		if res.RowsAffected == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "reward balance insufficient"})
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

	// Payable = item net (after reward) + delivery fee. The delivery fee is
	// excluded from order.Amount so it never enters the reward/返利 base.
	payAmount := req.Amount + deliveryFee

	// Zero-amount order (reward covered items AND no delivery fee owed): mark paid
	// directly, skip WeChat Pay. A delivery fee keeps payAmount > 0 so the fee is
	// still collected via WeChat Pay.
	if payAmount <= 0 {
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
		// Fee-free delivery edge: still dispatch the courier on auto-paid orders.
		if orderType == "delivery" {
			go services.DispatchShansong(order.ID)
		}
		c.JSON(http.StatusOK, gin.H{
			"id":           order.ID,
			"order_no":     order.OrderNo,
			"amount":       order.Amount,
			"delivery_fee": deliveryFee,
			"pay_amount":   payAmount,
			"status":       2,
		})
		return
	}

	// Call WeChat Pay JSAPI prepay
	amountCents := int64(math.Round(payAmount * 100))
	description := fmt.Sprintf("%s-订单%s", shop.Name, orderNo)
	prepay, err := services.CreateJSAPIPrepay(c.Request.Context(), user.OpenID, orderNo, description, amountCents)
	if err != nil {
		log.Printf("[order] wechat prepay failed for order %s: %v", orderNo, err)
		c.JSON(http.StatusOK, gin.H{
			"id":           order.ID,
			"order_no":     order.OrderNo,
			"amount":       order.Amount,
			"delivery_fee": deliveryFee,
			"pay_amount":   payAmount,
			"status":       order.Status,
			"error":        "prepay failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           order.ID,
		"order_no":     order.OrderNo,
		"amount":       order.Amount,
		"delivery_fee": deliveryFee,
		"pay_amount":   payAmount,
		"status":       order.Status,
		"prepay_id":    prepay.PrepayID,
		"time_stamp":   prepay.TimeStamp,
		"nonce_str":    prepay.NonceStr,
		"package":      prepay.Package,
		"sign_type":    prepay.SignType,
		"pay_sign":     prepay.PaySign,
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
		UserID: inviterID,
		Type:   "invite_reward",
		Amount: inviteReward,
		Remark: fmt.Sprintf("邀请奖励 %.2f 元", inviteReward),
	}
	if err := tx.Create(&walletLog).Error; err != nil {
		tx.Rollback()
		log.Printf("[invite_reward] create wallet_log failed: inviterID=%d err=%v", inviterID, err)
		return
	}

	tx.Commit()
	log.Printf("[invite_reward] rewarded: inviterID=%d inviteeID=%d amount=%.2f", inviterID, userID, inviteReward)
}
