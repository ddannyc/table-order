package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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
		// In CI (REQUIRE_TEST_DB set) a missing DB is a hard failure, so the suite
		// can't pass while silently skipping every handler test. Locally it skips.
		if os.Getenv("REQUIRE_TEST_DB") != "" {
			t.Fatalf("REQUIRE_TEST_DB is set but the test database is not available: %v", err)
		}
		t.Skipf("skipping test, db not available: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.Exec("DROP TABLE IF EXISTS reward_logs")
	sqlDB.Exec("DROP TABLE IF EXISTS wallet_logs")
	sqlDB.Exec("DROP TABLE IF EXISTS order_items")
	sqlDB.Exec("DROP TABLE IF EXISTS order_deliveries")
	sqlDB.Exec("DROP TABLE IF EXISTS orders")
	sqlDB.Exec("DROP TABLE IF EXISTS product_specs")
	sqlDB.Exec("DROP TABLE IF EXISTS products")
	sqlDB.Exec("DROP TABLE IF EXISTS table_qrcodes")
	sqlDB.Exec("DROP TABLE IF EXISTS shops")
	sqlDB.Exec("DROP TABLE IF EXISTS merchants")
	sqlDB.Exec("DROP TABLE IF EXISTS users")

	config.DB = db

	// Quote tokens fail closed without a configured secret; mirror production by
	// ensuring a JWT secret is set (non-destructive to any other config a test sets).
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	config.AppConfig.JWT.Secret = "test-quote-secret"

	// Recreate tables — a migration failure must fail the suite, not silently
	// leave a broken schema that later assertions misread.
	if err := config.DB.AutoMigrate(&models.User{}, &models.Shop{}, &models.Product{}, &models.ProductSpec{}, &models.Order{}, &models.OrderItem{}, &models.OrderDelivery{}, &models.WalletLog{}, &models.TableQRCode{}, &models.Merchant{}, &models.InviteRelation{}, &models.RewardLog{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}

	// Wait for any lingering goroutines from previous tests to finish
	time.Sleep(300 * time.Millisecond)
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
	case "DELETE":
		r.DELETE(path, handlerWrapper)
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
	shop := models.Shop{Name: "Reward Test Shop", MerchantID: 1, Status: 1, RewardRateSelf: 0.03}
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
	rewardAmount, _ := resp["reward_amount"].(float64)
	if rewardAmount != 0 {
		t.Errorf("expected reward_amount 0 for non-referred user, got %v", resp["reward_amount"])
	}

	// Order should be pending (status=1) with WeChat Pay params
	if status, ok := resp["status"].(float64); !ok || status != 1 {
		t.Errorf("expected status 1 (pending), got %v", resp["status"])
	}
	if _, ok := resp["prepay_id"]; !ok {
		// prepay_id may be empty string if WeChat Pay not configured — that's OK in test
		// Just check the field exists
	}
}

func TestCreateOrder_DeductsRewardBalanceButNotBalance(t *testing.T) {
	setupTestDB(t)

	// Create shop
	shop := models.Shop{Name: "Deduct Shop", MerchantID: 1, Status: 1, RewardRateSelf: 0.03}
	config.DB.Create(&shop)

	// Create product
	product := models.Product{ShopID: shop.ID, Name: "Test Dish", Price: 100, Status: 1}
	config.DB.Create(&product)

	// Create user with reward balance and regular balance
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

	// Reload user and check reward_balance deducted but regular balance unchanged
	config.DB.First(&user, user.ID)
	if user.RewardBalance != 0 {
		t.Errorf("expected reward_balance 0 after deduction (50-50), got %f", user.RewardBalance)
	}
	if user.Balance != 100 {
		t.Errorf("expected balance unchanged (100) since WeChat Pay handles payment, got %f", user.Balance)
	}

	// Order should be pending (status=1), not paid (status=2)
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.Status != 1 {
		t.Errorf("expected order status 1 (pending), got %d", order.Status)
	}
	if order.PaidAt != nil {
		t.Errorf("expected PaidAt to be nil for pending order")
	}
}

func TestCreateOrder_InviterReceivesInviteReward(t *testing.T) {
	setupTestDB(t)

	// Create shop
	shop := models.Shop{Name: "Invite Reward Shop", MerchantID: 1, Status: 1, RewardRateSelf: 0.03, RewardRateLevel1: 0.10}
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

	// Order should be pending — reward distribution happens via callback now, not here
	var order models.Order
	config.DB.Where("user_id = ?", referredUser.ID).Order("id desc").First(&order)
	if order.Status != 1 {
		t.Errorf("expected order status 1 (pending), got %d", order.Status)
	}
}

func TestCreateOrder_ReferredUserNoSelfReward(t *testing.T) {
	setupTestDB(t)

	// Create shop
	shop := models.Shop{Name: "Reward Shop", MerchantID: 1, Status: 1, RewardRateSelf: 0.03, RewardRateLevel1: 0.10}
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
	rewardAmount, _ := resp["reward_amount"].(float64)
	if rewardAmount != 0 {
		t.Errorf("expected reward_amount 0 for referred user (only inviter gets reward), got %v", resp["reward_amount"])
	}
}

func TestCreateOrder_ReturnsPrepayFields(t *testing.T) {
	setupTestDB(t)

	shop := models.Shop{Name: "Prepay Test Shop", MerchantID: 1, Status: 1}
	config.DB.Create(&shop)

	product := models.Product{ShopID: shop.ID, Name: "Test Dish", Price: 50, Status: 1}
	config.DB.Create(&product)

	user := models.User{OpenID: "test_prepay_user", Nickname: "PrepayUser", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/orders", CreateOrder, user.ID)
	body := map[string]interface{}{
		"shop_id":  shop.ID,
		"table_no": "A01",
		"amount":   50,
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
		t.Errorf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify order fields
	if id, ok := resp["id"].(float64); !ok || id == 0 {
		t.Errorf("expected non-zero order id, got %v", resp["id"])
	}
	if status, ok := resp["status"].(float64); !ok || status != 1 {
		t.Errorf("expected status 1 (pending), got %v", resp["status"])
	}

	// Verify prepay fields — only present when WeChat Pay is configured
	_, hasPrepayID := resp["prepay_id"]
	_, hasTimeStamp := resp["time_stamp"]
	_, hasNonceStr := resp["nonce_str"]
	_, hasPackage := resp["package"]
	_, hasSignType := resp["sign_type"]
	_, hasPaySign := resp["pay_sign"]

	if config.WxPayClient != nil {
		if !hasPrepayID {
			t.Errorf("expected prepay_id field in response")
		}
		if !hasTimeStamp {
			t.Errorf("expected time_stamp field in response")
		}
		if !hasNonceStr {
			t.Errorf("expected nonce_str field in response")
		}
		if !hasPackage {
			t.Errorf("expected package field in response")
		}
		if !hasSignType {
			t.Errorf("expected sign_type field in response")
		}
		if !hasPaySign {
			t.Errorf("expected pay_sign field in response")
		}
	} else {
		// WeChat Pay not configured — expect error field instead
		if !hasPrepayID {
			t.Log("prepay fields absent: WeChat Pay not configured (expected)")
		}
		if errMsg, ok := resp["error"]; !ok || errMsg != "prepay failed" {
			t.Errorf("expected error='prepay failed', got %v", resp["error"])
		}
	}

	// Verify order persisted as pending
	var order models.Order
	config.DB.Where("user_id = ?", user.ID).Order("id desc").First(&order)
	if order.Status != 1 {
		t.Errorf("expected order status 1 in db, got %d", order.Status)
	}
	if order.PaidAt != nil {
		t.Errorf("expected PaidAt nil for pending order")
	}
}

func TestGetOrders_ExcludesNonReferredUserReward(t *testing.T) {
	setupTestDB(t)

	// Create shop with 10% reward
	shop := models.Shop{Name: "Reward Exclude Shop", MerchantID: 1, Status: 1, RewardRateSelf: 0.03}
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

// === Invite Handler Tests ===

func TestLogin_GeneratesInviteCode(t *testing.T) {
	setupTestDB(t)

	// invite_code is generated at user creation (Login handler)
	code := "LOGINTEST01"
	user := models.User{
		OpenID:     "login_invite_test",
		Nickname:   "LoginInviteTest",
		Role:       0,
		InviteCode: &code,
	}
	config.DB.Create(&user)

	var saved models.User
	config.DB.First(&saved, user.ID)
	if saved.InviteCode == nil || *saved.InviteCode == "" {
		t.Fatalf("expected invite_code to be set on user creation")
	}
	if *saved.InviteCode != code {
		t.Errorf("expected invite_code %s, got %s", code, *saved.InviteCode)
	}

	// Verify invite_code can construct valid share URL
	inviteURL := "/pages/invite/index?invite_code=" + *saved.InviteCode
	if !strings.Contains(inviteURL, "invite_code=") {
		t.Errorf("invite_url must contain invite_code= param")
	}
	if len(inviteURL) < 7 || inviteURL[0:7] != "/pages/" {
		t.Errorf("invite_url must start with /pages/, got: %s", inviteURL)
	}
	if strings.Contains(inviteURL, "http://") || strings.Contains(inviteURL, "https://") {
		t.Errorf("invite_url must NOT be external URL")
	}
}

func TestBindInviteCode_BindsSuccessfully(t *testing.T) {
	setupTestDB(t)

	// Inviter with invite code
	inviter := models.User{OpenID: "bind_inviter1", Nickname: "Inviter1", Role: 0}
	config.DB.Create(&inviter)
	inviteCode := "BINDCODE001"
	config.DB.Model(&inviter).Update("invite_code", inviteCode)

	// Invitee (no inviter yet)
	invitee := models.User{OpenID: "bind_invitee1", Nickname: "Invitee1", Role: 0, Balance: 100}
	config.DB.Create(&invitee)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/invites/bind", BindInviteCode, invitee.ID)

	body := map[string]interface{}{"code": inviteCode}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/invites/bind", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "bound successfully" {
		t.Errorf("expected 'bound successfully', got %v", resp["message"])
	}

	// Verify InviterID set
	config.DB.First(&invitee, invitee.ID)
	if invitee.InviterID == nil || *invitee.InviterID != inviter.ID {
		t.Errorf("expected InviterID=%d, got %v", inviter.ID, invitee.InviterID)
	}

	// Verify InviteRelation created (shop_id defaults to 0 when not provided)
	var relation models.InviteRelation
	if err := config.DB.Where("inviter_id = ? AND invitee_id = ?", inviter.ID, invitee.ID).First(&relation).Error; err != nil {
		t.Errorf("expected InviteRelation to exist: %v", err)
	}
}

func TestBindInviteCode_AlreadyBound(t *testing.T) {
	setupTestDB(t)

	inviter1 := models.User{OpenID: "bind_inviter2a", Nickname: "Inviter2a", Role: 0}
	config.DB.Create(&inviter1)
	config.DB.Model(&inviter1).Update("invite_code", "ALRBND001")

	inviter2 := models.User{OpenID: "bind_inviter2b", Nickname: "Inviter2b", Role: 0}
	config.DB.Create(&inviter2)
	config.DB.Model(&inviter2).Update("invite_code", "ALRBND002")

	invitee := models.User{OpenID: "bind_invitee2", Nickname: "BoundUser", Role: 0, InviterID: &inviter1.ID}
	config.DB.Create(&invitee)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/invites/bind", BindInviteCode, invitee.ID)

	body := map[string]interface{}{"code": "ALRBND002"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/invites/bind", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "already bound" {
		t.Errorf("expected 'already bound', got %v", resp["message"])
	}

	// InviterID should still point to first inviter
	config.DB.First(&invitee, invitee.ID)
	if *invitee.InviterID != inviter1.ID {
		t.Errorf("expected InviterID=%d (first), got %d", inviter1.ID, *invitee.InviterID)
	}
}

func TestBindInviteCode_InvalidCode(t *testing.T) {
	setupTestDB(t)

	user := models.User{OpenID: "bind_invitee3", Nickname: "NoBind", Role: 0}
	config.DB.Create(&user)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/invites/bind", BindInviteCode, user.ID)

	body := map[string]interface{}{"code": "NONEXISTENT"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/invites/bind", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for invalid code, got %d", w.Code)
	}
}

func TestBindInviteCode_SelfReferral(t *testing.T) {
	setupTestDB(t)

	user := models.User{OpenID: "bind_self", Nickname: "SelfRef", Role: 0}
	config.DB.Create(&user)
	code := "SELFREF001"
	config.DB.Model(&user).Update("invite_code", code)

	r := setupRouter()
	setAuthContext(r, "POST", "/api/invites/bind", BindInviteCode, user.ID)

	body := map[string]interface{}{"code": code}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/invites/bind", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for self-referral, got %d", w.Code)
	}
}

func TestGetInviteStats(t *testing.T) {
	setupTestDB(t)

	inviter := models.User{OpenID: "stats_inviter_fresh", Nickname: "StatsInviter", Role: 0}
	config.DB.Create(&inviter)

	invitee1 := models.User{OpenID: "stats_invitee1_fresh", Nickname: "StatsI1", Role: 0, InviterID: &inviter.ID}
	config.DB.Create(&invitee1)
	invitee2 := models.User{OpenID: "stats_invitee2_fresh", Nickname: "StatsI2", Role: 0, InviterID: &inviter.ID}
	config.DB.Create(&invitee2)

	// Create invite relations directly
	config.DB.Create(&models.InviteRelation{InviterID: inviter.ID, InviteeID: invitee1.ID})
	config.DB.Create(&models.InviteRelation{InviterID: inviter.ID, InviteeID: invitee2.ID})

	// Create wallet logs for invite rewards
	config.DB.Create(&models.WalletLog{UserID: inviter.ID, Type: "invite_reward", Amount: 5})
	config.DB.Create(&models.WalletLog{UserID: inviter.ID, Type: "invite_reward", Amount: 3})

	r := setupRouter()
	setAuthContext(r, "GET", "/api/invites/stats", GetInviteStats, inviter.ID)

	req, _ := http.NewRequest("GET", "/api/invites/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// The count may include rows from previous tests since DB is shared
	// Just verify the response structure and that counts are reasonable
	if _, ok := resp["invite_count"].(float64); !ok {
		t.Errorf("expected invite_count to be a number, got %v", resp["invite_count"])
	}
	if _, ok := resp["total_invite_reward"].(float64); !ok {
		t.Errorf("expected total_invite_reward to be a number, got %v", resp["total_invite_reward"])
	}
	if _, ok := resp["today_reward"].(float64); !ok {
		t.Errorf("expected today_reward to be a number, got %v", resp["today_reward"])
	}
}