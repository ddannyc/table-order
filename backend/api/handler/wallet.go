package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/example/table-order/utils"
)

type WalletBalanceResponse struct {
	Balance       float64 `json:"balance"`
	RewardBalance float64 `json:"reward_balance"`
}

func GetBalance(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, WalletBalanceResponse{
		Balance:       user.Balance,
		RewardBalance: user.RewardBalance,
	})
}

func GetWalletLogs(c *gin.Context) {
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

	var logs []models.WalletLog
	var total int64

	config.DB.Model(&models.WalletLog{}).Where("user_id = ?", userID).Count(&total)
	config.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(int(offset)).Limit(int(pageSize)).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}