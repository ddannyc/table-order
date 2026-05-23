package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
)

func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	banStr := c.Query("ban")

	query := config.DB.Model(&models.User{})

	if banStr == "1" {
		query = query.Where("is_banned = ?", true)
	} else if banStr == "0" {
		query = query.Where("is_banned = ?", false)
	}

	var total int64
	query.Count(&total)

	var users []models.User
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at desc").Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func BanUser(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Ban bool `json:"ban"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	config.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_banned", req.Ban)

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func UpdateGlobalConfig(c *gin.Context) {
	// For MVP: platform-wide reward rate config
	var req struct {
		DefaultRewardRate float64 `json:"default_reward_rate"`
		DefaultInviteRate float64 `json:"default_invite_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// In production, store in config/Redis
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "config updated",
	})
}