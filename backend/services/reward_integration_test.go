package services

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := "host=localhost port=5432 user=postgres dbname=table_order_test sslmode=disable"
	var err error
	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect test DB: %v", err)
	}

	// Migrate test schema
	testDB.AutoMigrate(
		&models.User{},
		&models.Shop{},
		&models.Order{},
		&models.OrderItem{},
		&models.Product{},
		&models.InviteRelation{},
		&models.WalletLog{},
		&models.RewardLog{},
	)

	// Wire global DB for services
	config.DB = testDB

	code := m.Run()

	// Cleanup: drop all tables
	testDB.Migrator().DropTable(
		&models.RewardLog{},
		&models.WalletLog{},
		&models.InviteRelation{},
		&models.OrderItem{},
		&models.Order{},
		&models.Product{},
		&models.Shop{},
		&models.User{},
	)

	// Restore nil so handler tests don't use stale DB
	config.DB = nil

	os.Exit(code)
}

func setupUser(t *testing.T, overrides map[string]interface{}) models.User {
	t.Helper()
	suffix := time.Now().UnixNano()
	inviteCode := fmt.Sprintf("T%d", suffix%1000000000)
	if len(inviteCode) > 12 {
		inviteCode = inviteCode[:12]
	}
	openid := fmt.Sprintf("open_%d", suffix)
	user := models.User{
		OpenID:        openid,
		Nickname:      "test_user",
		PhoneVerified: true,
		InviteCode:    &inviteCode,
		RewardBalance: 0,
	}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func setupShop(t *testing.T) models.Shop {
	t.Helper()
	shop := models.Shop{
		Name:             "test_shop",
		RewardRateSelf:   0.03,
		RewardRateLevel1: 0.10,
		RewardRateLevel2: 0.04,
		RewardCeiling:    0.50,
	}
	if err := testDB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}
	return shop
}

func setupOrder(t *testing.T, userID uint, shopID uint, amount float64) models.Order {
	t.Helper()
	order := models.Order{
		OrderNo: "TEST" + time.Now().Format("150405000000"),
		UserID:  userID,
		ShopID:  shopID,
		Amount:  amount,
		Status:  2,
	}
	if err := testDB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	return order
}

func cleanTables(t *testing.T) {
	t.Helper()
	testDB.Exec("DELETE FROM reward_logs")
	testDB.Exec("DELETE FROM wallet_logs")
	testDB.Exec("DELETE FROM invite_relations")
	testDB.Exec("DELETE FROM order_items")
	testDB.Exec("DELETE FROM orders")
	testDB.Exec("DELETE FROM products")
	testDB.Exec("UPDATE users SET reward_balance = 0, reward_paused_at = NULL, last_consume_at = NULL, inviter_id = NULL")
}

// --- Integration Tests ---

func TestIntegration_SelfRewardOnly(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)
	shop := setupShop(t)
	order := setupOrder(t, user.ID, shop.ID, 100.00)

	DistributeReward(order.ID, user.ID, shop.ID, 100.00)

	// Allow goroutine to complete
	time.Sleep(100 * time.Millisecond)

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardBalance != 3.00 {
		t.Errorf("expected reward_balance 3.00, got %.2f", dbUser.RewardBalance)
	}

	var logs []models.RewardLog
	testDB.Where("user_id = ?", user.ID).Find(&logs)
	if len(logs) != 1 {
		t.Fatalf("expected 1 reward_log, got %d", len(logs))
	}
	if logs[0].Type != "self" || logs[0].Amount != 3.00 {
		t.Errorf("unexpected reward_log: type=%s amount=%.2f", logs[0].Type, logs[0].Amount)
	}
}

func TestIntegration_Level1Reward(t *testing.T) {
	cleanTables(t)
	inviter := setupUser(t, nil)
	invitee := setupUser(t, nil)

	// Bind invitee to inviter
	testDB.Model(&invitee).Update("inviter_id", inviter.ID)
	testDB.Create(&models.InviteRelation{InviterID: inviter.ID, InviteeID: invitee.ID})

	shop := setupShop(t)
	order := setupOrder(t, invitee.ID, shop.ID, 100.00)

	DistributeReward(order.ID, invitee.ID, shop.ID, 100.00)
	time.Sleep(100 * time.Millisecond)

	// Invitee gets self-reward (3%)
	var inviteeDB models.User
	testDB.First(&inviteeDB, invitee.ID)
	if inviteeDB.RewardBalance != 3.00 {
		t.Errorf("invitee expected 3.00, got %.2f", inviteeDB.RewardBalance)
	}

	// Inviter gets level-1 reward (10%)
	var inviterDB models.User
	testDB.First(&inviterDB, inviter.ID)
	if inviterDB.RewardBalance != 10.00 {
		t.Errorf("inviter expected 10.00, got %.2f", inviterDB.RewardBalance)
	}

	var logs []models.RewardLog
	testDB.Where("user_id = ?", inviter.ID).Find(&logs)
	if len(logs) != 1 || logs[0].Type != "invite_level1" {
		t.Errorf("expected 1 invite_level1 log, got %d type=%s", len(logs), logs[0].Type)
	}
}

func TestIntegration_Level2Reward(t *testing.T) {
	cleanTables(t)
	level2 := setupUser(t, nil) // top-level inviter
	level1 := setupUser(t, nil)
	consumer := setupUser(t, nil)

	// Bind chain: level2 -> level1 -> consumer
	testDB.Model(&level1).Update("inviter_id", level2.ID)
	testDB.Create(&models.InviteRelation{InviterID: level2.ID, InviteeID: level1.ID})
	testDB.Model(&consumer).Update("inviter_id", level1.ID)
	testDB.Create(&models.InviteRelation{InviterID: level1.ID, InviteeID: consumer.ID})

	shop := setupShop(t)
	order := setupOrder(t, consumer.ID, shop.ID, 100.00)

	DistributeReward(order.ID, consumer.ID, shop.ID, 100.00)
	time.Sleep(100 * time.Millisecond)

	var level2DB models.User
	testDB.First(&level2DB, level2.ID)
	if level2DB.RewardBalance != 4.00 {
		t.Errorf("level2 expected 4.00, got %.2f", level2DB.RewardBalance)
	}

	var level1DB models.User
	testDB.First(&level1DB, level1.ID)
	if level1DB.RewardBalance != 10.00 {
		t.Errorf("level1 expected 10.00, got %.2f", level1DB.RewardBalance)
	}

	var consumerDB models.User
	testDB.First(&consumerDB, consumer.ID)
	if consumerDB.RewardBalance != 3.00 {
		t.Errorf("consumer expected 3.00, got %.2f", consumerDB.RewardBalance)
	}
}

func TestIntegration_PausedUserSkipsReward(t *testing.T) {
	cleanTables(t)
	now := time.Now()
	user := setupUser(t, map[string]interface{}{"reward_paused_at": now})
	shop := setupShop(t)
	order := setupOrder(t, user.ID, shop.ID, 100.00)

	// Manually set paused
	testDB.Model(&user).Update("reward_paused_at", now)

	DistributeReward(order.ID, user.ID, shop.ID, 100.00)
	time.Sleep(100 * time.Millisecond)

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardBalance != 0 {
		t.Errorf("paused user should get 0 reward, got %.2f", dbUser.RewardBalance)
	}
}

func TestIntegration_CheckAndPauseInactivity(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)

	// Set last consume to 91 days ago
	past := time.Now().AddDate(0, 0, -91)
	testDB.Model(&user).Update("last_consume_at", past)

	paused := CheckAndPauseInactivity(user.ID)
	if !paused {
		t.Error("expected user to be paused after 91 days inactivity")
	}

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardPausedAt == nil {
		t.Error("expected reward_paused_at to be set")
	}
}

func TestIntegration_ResumeReward(t *testing.T) {
	cleanTables(t)
	now := time.Now()
	user := setupUser(t, nil)
	testDB.Model(&user).Updates(map[string]interface{}{
		"reward_paused_at": now,
		"last_consume_at":  now.AddDate(0, 0, -1),
	})

	// Simulate order — resume
	ResumeReward(user.ID)

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardPausedAt != nil {
		t.Error("expected reward_paused_at to be nil after resume")
	}
}

func TestIntegration_SweepExpiredRewards(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)
	shop := setupShop(t)
	order := setupOrder(t, user.ID, shop.ID, 100.00)

	// Issue reward that's already expired
	expiredTime := time.Now().AddDate(0, 0, -181)
	rl := models.RewardLog{
		UserID:    user.ID,
		OrderID:   order.ID,
		Type:      "self",
		Amount:    3.00,
		ExpiresAt: expiredTime,
		Expired:   false,
	}
	testDB.Create(&rl)

	// Credit balance
	testDB.Model(&user).Update("reward_balance", 3.00)

	SweepExpiredRewards()

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardBalance != 0 {
		t.Errorf("expected reward_balance 0 after sweep, got %.2f", dbUser.RewardBalance)
	}

	var rlDB models.RewardLog
	testDB.First(&rlDB, rl.ID)
	if !rlDB.Expired {
		t.Error("expected reward_log to be marked expired")
	}
}

func TestIntegration_NoRewardForExcludedCategory(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)
	shop := setupShop(t)
	shop.RewardExcludeCategories = `["特价"]`
	testDB.Save(&shop)

	order := setupOrder(t, user.ID, shop.ID, 50.00)

	// Create order item with excluded category
	product := models.Product{
		ShopID:   shop.ID,
		Name:     "特价商品",
		Price:    50.00,
		Category: "特价",
		Status:   1,
	}
	testDB.Create(&product)
	item := models.OrderItem{
		OrderID:     order.ID,
		ProductID:   product.ID,
		ProductName: product.Name,
		Price:       50.00,
		Quantity:    1,
		Subtotal:    50.00,
	}
	testDB.Create(&item)

	DistributeReward(order.ID, user.ID, shop.ID, 50.00)
	time.Sleep(100 * time.Millisecond)

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardBalance != 0 {
		t.Errorf("expected 0 reward for excluded category, got %.2f", dbUser.RewardBalance)
	}
}

func TestIntegration_WalletLogCreated(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)
	shop := setupShop(t)
	order := setupOrder(t, user.ID, shop.ID, 200.00)

	DistributeReward(order.ID, user.ID, shop.ID, 200.00)
	time.Sleep(100 * time.Millisecond)

	var logs []models.WalletLog
	testDB.Where("user_id = ? AND type = ?", user.ID, "reward").Find(&logs)
	if len(logs) != 1 {
		t.Fatalf("expected 1 wallet_log, got %d", len(logs))
	}
	if logs[0].Amount != 6.00 {
		t.Errorf("expected wallet_log amount 6.00, got %.2f", logs[0].Amount)
	}
}

func TestIntegration_NoSelfRewardWhenRateIsZero(t *testing.T) {
	cleanTables(t)
	user := setupUser(t, nil)
	shop := setupShop(t)
	shop.RewardRateSelf = 0
	testDB.Save(&shop)
	order := setupOrder(t, user.ID, shop.ID, 100.00)

	DistributeReward(order.ID, user.ID, shop.ID, 100.00)
	time.Sleep(100 * time.Millisecond)

	var dbUser models.User
	testDB.First(&dbUser, user.ID)
	if dbUser.RewardBalance != 0 {
		t.Errorf("expected 0 reward when self rate is 0, got %.2f", dbUser.RewardBalance)
	}
}

func TestIntegration_InviteRewardNoLevel2WhenInviterHasNoInviter(t *testing.T) {
	cleanTables(t)
	inviter := setupUser(t, nil)
	invitee := setupUser(t, nil)
	testDB.Model(&invitee).Update("inviter_id", inviter.ID)
	testDB.Create(&models.InviteRelation{InviterID: inviter.ID, InviteeID: invitee.ID})

	shop := setupShop(t)
	order := setupOrder(t, invitee.ID, shop.ID, 100.00)

	DistributeReward(order.ID, invitee.ID, shop.ID, 100.00)
	time.Sleep(100 * time.Millisecond)

	// Inviter gets level-1, but no level-2 (inviter has no inviter)
	var inviterDB models.User
	testDB.First(&inviterDB, inviter.ID)
	if inviterDB.RewardBalance != 10.00 {
		t.Errorf("inviter expected 10.00, got %.2f", inviterDB.RewardBalance)
	}

	// No level-2 user exists, so only 2 reward_logs: self + invite_level1
	var count int64
	testDB.Model(&models.RewardLog{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 reward_logs (self + level1), got %d", count)
	}
}
