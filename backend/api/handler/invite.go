package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
)

// GenerateInviteQR generates a WeChat mini-program QR code for invite sharing.
func GenerateInviteQR(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Ensure invite code exists
	if user.InviteCode == nil || *user.InviteCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "generate invite code first"})
		return
	}

	// scene: max 32 chars, encode invite_code
	scene := fmt.Sprintf("ic=%s", *user.InviteCode)
	if len(scene) > 32 {
		scene = scene[:32]
	}

	png, err := services.GetWXACodeUnlimited(scene, "pages/home/index")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("generate qr failed: %v", err)})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", png.Buffer)
}

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

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Idempotent: return existing code if already generated
	if user.InviteCode != nil && *user.InviteCode != "" {
		inviteURL := fmt.Sprintf("/pages/invite/index?invite_code=%s", *user.InviteCode)
		c.JSON(http.StatusOK, InviteResponse{
			InviteCode: *user.InviteCode,
			InviteURL:  inviteURL,
		})
		return
	}

	inviteCode := uuid.New().String()[:12]
	if err := config.DB.Model(&user).Update("invite_code", inviteCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate invite code failed"})
		return
	}

	inviteURL := fmt.Sprintf("/pages/invite/index?invite_code=%s", inviteCode)
	c.JSON(http.StatusOK, InviteResponse{
		InviteCode: inviteCode,
		InviteURL:  inviteURL,
	})
}

// BindInviteCode called when user opens invite share link
func BindInviteCode(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Code   string `json:"code"`
		ShopID uint   `json:"shop_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// Find inviter by invite code
	var inviter models.User
	if err := config.DB.Where("invite_code = ?", req.Code).First(&inviter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid invite code"})
		return
	}

	// Prevent self-referral
	if inviter.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot invite yourself"})
		return
	}

	// Fetch invitee
	var invitee models.User
	if err := config.DB.First(&invitee, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Already bound — first inviter wins
	if invitee.InviterID != nil {
		c.JSON(http.StatusOK, gin.H{"message": "already bound"})
		return
	}

	// Atomic: set InviterID + create InviteRelation
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
		return
	}

	if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("inviter_id", inviter.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bind failed"})
		return
	}

	relation := models.InviteRelation{
		InviterID: inviter.ID,
		InviteeID: userID,
		ShopID:    req.ShopID,
	}
	if err := tx.Create(&relation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bind failed"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "bound successfully"})
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