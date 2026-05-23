package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/example/table-order/config"
	"github.com/example/table-order/middleware"
	"github.com/example/table-order/models"
	"github.com/example/table-order/services"
)

type LoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type LoginResponse struct {
	Token    string       `json:"token"`
	User     models.User  `json:"user"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	var openid string

	// Try real WeChat API
	if config.AppConfig.WeChat.AppID != "" && config.AppConfig.WeChat.AppSecret != "" {
		session, err := services.GetWeChatSession(req.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "微信登录失败: " + err.Error()})
			return
		}
		if session.OpenID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "微信登录失败: 未获取到 openid"})
			return
		}
		openid = session.OpenID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "微信未配置"})
		return
	}

	var user models.User
	if err := config.DB.Where("open_id = ?", openid).First(&user).Error; err != nil {
		// Create new user
		user = models.User{
			OpenID: openid,
			Nickname: "用户" + openid[len(openid)-6:],
			Role: 0,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
			return
		}
	}

	if user.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "user banned"})
		return
	}

	// Generate JWT
	claims := &middleware.Claims{
		UserID: user.ID,
		OpenID: user.OpenID,
		Role:   user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate token failed"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: tokenString,
		User:  user,
	})
}