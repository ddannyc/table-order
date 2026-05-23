package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

type GenerateInviteRequest struct {
	ShopID uint `json:"shop_id"`
}

type InviteResponse struct {
	InviteCode string `json:"invite_code"`
	InviteURL  string `json:"invite_url"`
}

func GenerateInviteCode(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req GenerateInviteRequest
	c.ShouldBindJSON(&req)

	// Generate unique invite code
	inviteCode := uuid.New().String()[:12]

	// Store invite code mapping in Redis for quick lookup
	// For MVP, store in DB
	inviteURL := fmt.Sprintf("https://domain.com/invite?code=%s&u=%d", inviteCode, userID)

	c.JSON(http.StatusOK, InviteResponse{
		InviteCode: inviteCode,
		InviteURL:  inviteURL,
	})
}

// BindInviteCode called when user scans an invite QR or clicks invite link
func BindInviteCode(c *gin.Context) {
	userID := c.GetUint("user_id")
	code := c.Query("code")
	shopID := c.Query("shop_id")

	if code == "" {
		return
	}

	// TODO: Look up inviter from invite code (Redis/DB)
	// For MVP, just store the relation
	// inviterID := getInviterFromCode(code)

	// Only bind if user doesn't already have an inviter
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return
	}

	if user.InviterID != nil {
		return // Already has inviter
	}

	// TODO: Set inviter_id after looking up from code
	_ = shopID
}

func GetInviteStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Count invitees
	var inviteCount int64
	config.DB.Model(&models.InviteRelation{}).Where("inviter_id = ?", userID).Count(&inviteCount)

	// Sum invite rewards
	var totalInviteReward float64
	config.DB.Model(&models.WalletLog{}).
		Where("user_id = ? AND type = ?", userID, "invite_reward").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalInviteReward)

	// Today's reward
	todayReward := float64(0)
	config.DB.Model(&models.WalletLog{}).
		Where("user_id = ? AND type IN (?, ?) AND created_at > CURRENT_DATE", userID, "reward", "invite_reward").
		Select("COALESCE(SUM(amount), 0)").Scan(&todayReward)

	c.JSON(http.StatusOK, gin.H{
		"invite_count":      inviteCount,
		"total_invite_reward": totalInviteReward,
		"today_reward":      todayReward,
	})
}