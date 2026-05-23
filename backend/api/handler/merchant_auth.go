package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/example/table-order/config"
	"github.com/example/table-order/middleware"
	"github.com/example/table-order/models"
)

type MerchantLoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type MerchantLoginResponse struct {
	Token    string         `json:"token"`
	Merchant models.Merchant `json:"merchant"`
}

func MerchantLogin(c *gin.Context) {
	var req MerchantLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone and password required"})
		return
	}

	var merchant models.Merchant
	if err := config.DB.Where("phone = ?", req.Phone).First(&merchant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(merchant.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if merchant.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "account disabled"})
		return
	}

	// Generate JWT with role=1 (merchant)
	claims := &middleware.Claims{
		UserID: merchant.ID,
		Role:   1,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(config.AppConfig.JWT.Secret))

	c.JSON(http.StatusOK, MerchantLoginResponse{
		Token:    tokenString,
		Merchant: merchant,
	})
}

func RegisterMerchant(c *gin.Context) {
	var req MerchantLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone and password required"})
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	merchant := models.Merchant{
		Phone:    req.Phone,
		Password: string(hashed),
		Name:     "商户" + req.Phone[len(req.Phone)-4:],
		Status:   1,
	}

	if err := config.DB.Create(&merchant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	c.JSON(http.StatusOK, merchant)
}