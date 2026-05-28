package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
	"github.com/example/table-order/utils"
)

type RewardBalanceResponse struct {
	RewardBalance float64 `json:"reward_balance"`
	RewardPaused  bool    `json:"reward_paused"`
}

func GetRewardBalance(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, RewardBalanceResponse{
		RewardBalance: user.RewardBalance,
		RewardPaused:  user.RewardPausedAt != nil && !user.RewardPausedAt.IsZero(),
	})
}

func GetRewardLogs(c *gin.Context) {
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

	var logs []models.RewardLog
	var total int64

	config.DB.Model(&models.RewardLog{}).Where("user_id = ? AND expired = ?", userID, false).Count(&total)
	config.DB.Where("user_id = ? AND expired = ?", userID, false).Order("created_at desc").Offset(int(offset)).Limit(int(pageSize)).Find(&logs)

	typeLabelMap := map[string]string{
		"self":          "自购返利",
		"invite_level1": "直推奖励",
		"invite_level2": "间推奖励",
	}
	type rewardLogVO struct {
		models.RewardLog
		TypeLabel string `json:"type_label"`
		ExpiresIn int    `json:"expires_in_days"`
	}
	result := make([]rewardLogVO, len(logs))
	for i, l := range logs {
		expiresIn := int(l.ExpiresAt.Sub(l.CreatedAt).Hours() / 24)
		result[i] = rewardLogVO{
			RewardLog: l,
			TypeLabel: typeLabelMap[l.Type],
			ExpiresIn: expiresIn,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetRewardExpiryInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	services.SweepExpiredRewards()

	var user models.User
	config.DB.First(&user, userID)

	// Count expiring soon (within 30 days)
	var expiringSoon int64
	config.DB.Model(&models.RewardLog{}).
		Where("user_id = ? AND expired = ? AND expires_at < NOW() + INTERVAL '30 days'", userID, false).
		Count(&expiringSoon)

	c.JSON(http.StatusOK, gin.H{
		"reward_balance":    user.RewardBalance,
		"reward_paused":     user.RewardPausedAt != nil && !user.RewardPausedAt.IsZero(),
		"expiring_soon_count": expiringSoon,
	})
}
