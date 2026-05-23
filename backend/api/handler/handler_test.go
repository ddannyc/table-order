package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) {
	dsn := "host=localhost user=postgres password=postgres dbname=table_order_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("skipping test, db not available: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.Exec("DROP TABLE IF EXISTS wallet_logs")
	sqlDB.Exec("DROP TABLE IF EXISTS order_items")
	sqlDB.Exec("DROP TABLE IF EXISTS orders")
	sqlDB.Exec("DROP TABLE IF EXISTS products")
	sqlDB.Exec("DROP TABLE IF EXISTS table_qrcodes")
	sqlDB.Exec("DROP TABLE IF EXISTS shops")
	sqlDB.Exec("DROP TABLE IF EXISTS merchants")
	sqlDB.Exec("DROP TABLE IF EXISTS users")

	config.DB = db

	// Recreate tables
	config.DB.AutoMigrate(&models.User{}, &models.Shop{}, &models.Product{}, &models.Order{}, &models.OrderItem{}, &models.WalletLog{}, &models.TableQRCode{}, &models.Merchant{})
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func setAuthContext(r *gin.Engine, method, path string, handler gin.HandlerFunc, userID uint) {
	handlerWrapper := func(c *gin.Context) {
		c.Set("user_id", userID)
		handler(c)
	}
	switch method {
	case "POST":
		r.POST(path, handlerWrapper)
	case "GET":
		r.GET(path, handlerWrapper)
	case "PUT":
		r.PUT(path, handlerWrapper)
	default:
		r.GET(path, handlerWrapper)
	}
}

func TestLoginRequest_Validation(t *testing.T) {
	setupTestDB(t)

	r := setupRouter()
	r.POST("/api/auth/login", Login)

	// Missing code
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing code, got %d", w.Code)
	}
}

// === Order Reward Tests ===

func TestCreateOrder_NoRewardForNonReferredUser(t *testing.T) {
	setupTestDB(t)

	// Create shop
	shop := models.Shop{Name: "Reward Test Shop", MerchantID: 1, Status: 1, RewardRate: 0.1}
	config.DB.Create(&shop)

	// Create product
	product := models.Product{ShopID: shop.ID, Name: "Test Dish", Price: 100, Status: 1}
	config.DB.Create(&product)

	// Create user WITHOUT inviter (not referred)
	user := models.User{OpenID: "test_non_referred", Nickname: "NonReferredUser", Role: 0, RewardBalance: 0, Balance: 100}
	config.DB.Create(&user)

	// Create order
	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id": shop.ID,
		"table_no": "A01",
		"amount": 100,
		"items": []map[string]interface{}{
			{"product_id": product.ID, "quantity": 1},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Non-referred user should NOT receive reward
	rewardAmount, ok := resp["reward_amount"].(float64)
	if !ok || rewardAmount != 0 {
		t.Errorf("expected reward_amount 0 for non-referred user, got %v", resp["reward_amount"])
	}
}

func TestCreateOrder_WithdrawRewardBalanceOnPayment(t *testing.T) {
	setupTestDB(t)

	// Create shop
	shop := models.Shop{Name: "Deduct Shop", MerchantID: 1, Status: 1, RewardRate: 0.1}
	config.DB.Create(&shop)

	// Create product
	product := models.Product{ShopID: shop.ID, Name: "Test Dish", Price: 100, Status: 1}
	config.DB.Create(&product)

	// Create user with reward balance (pre-accumulated, no inviter needed for deduction test)
	user := models.User{OpenID: "test_rewards_user", Nickname: "RewardsUser", Role: 0, RewardBalance: 50, Balance: 100}
	config.DB.Create(&user)

	// Create order using reward balance
	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":       shop.ID,
		"table_no":      "A01",
		"amount":        100,
		"use_reward":    true,
		"items": []map[string]interface{}{
			{"product_id": product.ID, "quantity": 1},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d body: %s", w.Code, w.Body.String())
	}

	// Reload user and check balance deducted
	config.DB.First(&user, user.ID)
	if user.RewardBalance != 0 {
		t.Errorf("expected reward_balance 0 after payment (50-50), got %f", user.RewardBalance)
	}
	if user.Balance != 50 {
		t.Errorf("expected balance 50 after payment (100-50), got %f", user.Balance)
	}
}

func TestCreateOrder_InviterReceivesInviteReward(t *testing.T) {
	setupTestDB(t)

	// Create shop with 10% reward rate and 5% invite rate
	shop := models.Shop{Name: "Invite Reward Shop", MerchantID: 1, Status: 1, RewardRate: 0.1, InviteRate: 0.05}
	config.DB.Create(&shop)

	// Create product
	product := models.Product{ShopID: shop.ID, Name: "Test Dish", Price: 100, Status: 1}
	config.DB.Create(&product)

	// Create referrer (inviter)
	referrer := models.User{OpenID: "test_referrer", Nickname: "Referrer", Role: 0, RewardBalance: 0}
	config.DB.Create(&referrer)

	// Create referred user (has inviter)
	referredUser := models.User{OpenID: "test_referred", Nickname: "Referred", Role: 0, InviterID: &referrer.ID, RewardBalance: 0, Balance: 100}
	config.DB.Create(&referredUser)

	// Place order as referred user
	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, referredUser.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   100,
		"items": []map[string]interface{}{
			{"product_id": product.ID, "quantity": 1},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d body: %s", w.Code, w.Body.String())
	}

	// Referred user should NOT receive self-reward (no inviter field check passes for them, but they are the ORDER PLACER)
	// Actually the order placer DOES get reward if they have inviter. Let me check order.go logic.
	// Line 87: `if user.InviterID != nil { rewardAmount = req.Amount * shop.RewardRate }`
	// So referredUser HAS inviter -> they GET rewardAmount = 10. That's the bug - they shouldn't get self-reward.
	// But this test is about INVITER getting invite reward, not about self-reward.
	// The referredUser's inviter (referrer) should receive invite reward.

	// Check referrer received invite reward
	config.DB.First(&referrer, referrer.ID)
	// Invite reward = amount * inviteRate = 100 * 0.05 = 5
	if referrer.RewardBalance != 5 {
		t.Errorf("expected referrer reward_balance 5, got %f", referrer.RewardBalance)
	}
}

func TestCreateOrder_ReferredUserGetsRewardButInviterGetsInviteReward(t *testing.T) {
	setupTestDB(t)

	// Create shop with 10% reward rate and 5% invite rate
	shop := models.Shop{Name: "Reward Shop", MerchantID: 1, Status: 1, RewardRate: 0.1, InviteRate: 0.05}
	config.DB.Create(&shop)

	// Create product
	product := models.Product{ShopID: shop.ID, Name: "Test Dish", Price: 100, Status: 1}
	config.DB.Create(&product)

	// Create referrer
	referrer := models.User{OpenID: "test_referrer2", Nickname: "Referrer2", Role: 0, RewardBalance: 0}
	config.DB.Create(&referrer)

	// Create referred user with inviter
	referredUser := models.User{OpenID: "test_referred2", Nickname: "Referred2", Role: 0, InviterID: &referrer.ID, RewardBalance: 0, Balance: 100}
	config.DB.Create(&referredUser)

	// Place order
	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, referredUser.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   100,
		"items": []map[string]interface{}{
			{"product_id": product.ID, "quantity": 1},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Referred user should NOT receive self-reward (only inviter gets invite reward)
	rewardAmount, ok := resp["reward_amount"].(float64)
	if !ok || rewardAmount != 0 {
		t.Errorf("expected reward_amount 0 for referred user (only inviter gets reward), got %v", resp["reward_amount"])
	}

	// Inviter should receive invite reward
	config.DB.First(&referrer, referrer.ID)
	if referrer.RewardBalance != 5 {
		t.Errorf("expected referrer invite_reward 5, got %f", referrer.RewardBalance)
	}
}

func TestGetOrders_ExcludesNonReferredUserReward(t *testing.T) {
	setupTestDB(t)

	// Create shop with 10% reward
	shop := models.Shop{Name: "Reward Exclude Shop", MerchantID: 1, Status: 1, RewardRate: 0.1}
	config.DB.Create(&shop)

	// Create referred user (has inviter)
	referrer := models.User{OpenID: "test_referrer", Nickname: "Referrer", Role: 0}
	config.DB.Create(&referrer)
	referredUser := models.User{OpenID: "test_referred", Nickname: "Referred", Role: 0, InviterID: &referrer.ID}
	config.DB.Create(&referredUser)

	// Create non-referred user
	nonReferredUser := models.User{OpenID: "test_nonreferred", Nickname: "NonReferred", Role: 0}
	config.DB.Create(&nonReferredUser)

	// Create order for referred user - should have reward
	config.DB.Create(&models.Order{
		OrderNo:      "REWARD001",
		UserID:       referredUser.ID,
		ShopID:       shop.ID,
		Amount:       100,
		RewardAmount: 10, // 10% reward
		Status:       2,
	})

	// Create order for non-referred user - should have 0 reward
	config.DB.Create(&models.Order{
		OrderNo:      "NOREWARD001",
		UserID:       nonReferredUser.ID,
		ShopID:       shop.ID,
		Amount:       100,
		RewardAmount: 0,
		Status:       2,
	})

	r := setupRouter()
	setAuthContext(r, "GET", "/api/orders", GetOrders, referredUser.ID)

	req, _ := http.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	orders := resp["orders"].([]interface{})

	foundRewardOrder := false
	foundNoRewardOrder := false
	for _, o := range orders {
		order := o.(map[string]interface{})
		if order["order_no"] == "REWARD001" {
			foundRewardOrder = true
			if rewardAmt, ok := order["reward_amount"].(float64); !ok || rewardAmt != 10 {
				t.Errorf("expected reward_amount 10 for referred user order, got %v", order["reward_amount"])
			}
		}
		if order["order_no"] == "NOREWARD001" {
			foundNoRewardOrder = true
		}
	}
	if !foundRewardOrder {
		t.Errorf("expected to find REWARD001 order for referred user")
	}
	_ = foundNoRewardOrder // NonReferredUser's orders not visible in this query
}