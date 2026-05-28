package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"gorm.io/gorm"
)

const (
	rewardInactiveDays = 90
	rewardExpireDays   = 180
)

var rewardTypeLabels = map[string]string{
	"self":          "自购返利",
	"invite_level1": "直推奖励",
	"invite_level2": "间推奖励",
}

// DistributeReward handles 3-tier reward distribution after order payment.
// Called async — does not block order flow.
func DistributeReward(orderID uint, userID uint, shopID uint, amount float64) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		log.Printf("[reward] user not found: userID=%d err=%v", userID, err)
		return
	}

	if user.RewardPausedAt != nil && !user.RewardPausedAt.IsZero() {
		log.Printf("[reward] user paused, skipping: userID=%d", userID)
		return
	}

	var shop models.Shop
	if err := config.DB.First(&shop, shopID).Error; err != nil {
		log.Printf("[reward] shop not found: shopID=%d err=%v", shopID, err)
		return
	}

	if shop.RewardExcludeCategories != "[]" && shop.RewardExcludeCategories != "" {
		var orderItems []models.OrderItem
		config.DB.Where("order_id = ?", orderID).Find(&orderItems)
		if hasExcludedCategory(orderItems, shop) {
			log.Printf("[reward] order has excluded categories, skipping reward: orderID=%d", orderID)
			return
		}
	}

	expiresAt := time.Now().AddDate(0, 0, rewardExpireDays)

	issueReward(userID, orderID, "self", amount*shop.RewardRateSelf, &userID, expiresAt)

	if user.InviterID == nil {
		return
	}

	issueReward(*user.InviterID, orderID, "invite_level1", amount*shop.RewardRateLevel1, &userID, expiresAt)

	var inviter models.User
	if err := config.DB.First(&inviter, *user.InviterID).Error; err != nil || inviter.InviterID == nil {
		return
	}

	issueReward(*inviter.InviterID, orderID, "invite_level2", amount*shop.RewardRateLevel2, &userID, expiresAt)
}

func issueReward(userID uint, orderID uint, rewardType string, amount float64, fromUserID *uint, expiresAt time.Time) {
	if amount <= 0 {
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		log.Printf("[reward] tx begin failed: err=%v", tx.Error)
		return
	}

	if err := tx.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("reward_balance", gorm.Expr("reward_balance + ?", amount)).Error; err != nil {
		tx.Rollback()
		log.Printf("[reward] update reward_balance failed: userID=%d err=%v", userID, err)
		return
	}

	rewardLog := models.RewardLog{
		UserID:     userID,
		OrderID:    orderID,
		Type:       rewardType,
		Amount:     amount,
		FromUserID: fromUserID,
		ExpiresAt:  expiresAt,
	}
	if err := tx.Create(&rewardLog).Error; err != nil {
		tx.Rollback()
		log.Printf("[reward] create reward_log failed: err=%v", err)
		return
	}

	walletLog := models.WalletLog{
		UserID:  userID,
		Type:    "reward",
		Amount:  amount,
		OrderID: &orderID,
		Remark:  fmt.Sprintf("%s %.2f 元", rewardTypeLabels[rewardType], amount),
	}
	if err := tx.Create(&walletLog).Error; err != nil {
		tx.Rollback()
		log.Printf("[reward] create wallet_log failed: err=%v", err)
		return
	}

	tx.Commit()
	log.Printf("[reward] issued: userID=%d type=%s amount=%.2f orderID=%d", userID, rewardType, amount, orderID)
}

// CheckAndPauseInactivity checks if user has been inactive for 90+ days.
// If so, sets reward_paused_at. Returns true if paused.
func CheckAndPauseInactivity(userID uint) bool {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return false
	}

	if user.LastConsumeAt == nil {
		return false
	}

	if time.Since(*user.LastConsumeAt).Hours() >= float64(rewardInactiveDays*24) {
		now := time.Now()
		config.DB.Model(&user).UpdateColumn("reward_paused_at", &now)
		log.Printf("[reward] user paused due to inactivity: userID=%d lastConsume=%v", userID, *user.LastConsumeAt)
		return true
	}

	return false
}

// ResumeReward un-pauses user if they have consumed recently (called on order).
func ResumeReward(userID uint) {
	config.DB.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("reward_paused_at", nil)
}

// SweepExpiredRewards marks expired reward_logs and deducts from reward_balance.
// Called on demand or periodically.
func SweepExpiredRewards() {
	now := time.Now()

	var expiredLogs []models.RewardLog
	config.DB.Where("expires_at < ? AND expired = ?", now, false).Find(&expiredLogs)

	for _, rl := range expiredLogs {
		tx := config.DB.Begin()
		if tx.Error != nil {
			continue
		}

		// Deduct from reward_balance
		if err := tx.Model(&models.User{}).Where("id = ?", rl.UserID).
			UpdateColumn("reward_balance", gorm.Expr("GREATEST(reward_balance - ?, 0)", rl.Amount)).Error; err != nil {
			tx.Rollback()
			continue
		}

		// Mark expired
		if err := tx.Model(&rl).UpdateColumn("expired", true).Error; err != nil {
			tx.Rollback()
			continue
		}

		tx.Commit()
	}

	if len(expiredLogs) > 0 {
		log.Printf("[reward] swept %d expired rewards", len(expiredLogs))
	}
}

func hasExcludedCategory(items []models.OrderItem, shop models.Shop) bool {
	if shop.RewardExcludeCategories == "" || shop.RewardExcludeCategories == "[]" {
		return false
	}

	var categories []string
	if err := json.Unmarshal([]byte(shop.RewardExcludeCategories), &categories); err != nil {
		return false
	}
	excludeSet := make(map[string]bool, len(categories))
	for _, c := range categories {
		excludeSet[c] = true
	}

	var productIDs []uint
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}
	var products []models.Product
	config.DB.Where("id IN ?", productIDs).Find(&products)
	for _, p := range products {
		if excludeSet[p.Category] {
			return true
		}
	}
	return false
}

func parseJSONArray(jsonStr string, target *[]string) error {
	return json.Unmarshal([]byte(jsonStr), target)
}
