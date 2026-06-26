package handler

import (
	"net/http"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/gin-gonic/gin"
)

type CreateShopRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Hours       string `json:"hours"`
}

func CreateShop(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	var req CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}

	shop := models.Shop{
		MerchantID:  merchantID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		Hours:       req.Hours,
		Status:      1,
	}

	if err := config.DB.Create(&shop).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create shop failed"})
		return
	}

	c.JSON(http.StatusOK, shop)
}

func GetShops(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	var shops []models.Shop
	config.DB.Where("merchant_id = ?", merchantID).Find(&shops)

	c.JSON(http.StatusOK, shops)
}

// publicShopDTO is the shop view exposed to customers. It deliberately omits
// the merchant's commission rates (reward_rate_*) and excluded-category config;
// only reward_ceiling is needed client-side to show the discount cap.
type publicShopDTO struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	Phone         string  `json:"phone"`
	Hours         string  `json:"hours"`
	Logo          string  `json:"logo"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	RewardCeiling float64 `json:"reward_ceiling"`
	Status        int     `json:"status"`
}

func toPublicShopDTO(shop models.Shop) publicShopDTO {
	return publicShopDTO{
		ID:            shop.ID,
		Name:          shop.Name,
		Description:   shop.Description,
		Address:       shop.Address,
		City:          shop.City,
		Phone:         shop.Phone,
		Hours:         shop.Hours,
		Logo:          shop.Logo,
		Latitude:      shop.Latitude,
		Longitude:     shop.Longitude,
		RewardCeiling: shop.RewardCeiling,
		Status:        shop.Status,
	}
}

func GetShop(c *gin.Context) {
	shopID := c.Param("id")

	var shop models.Shop
	if err := config.DB.First(&shop, shopID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	c.JSON(http.StatusOK, toPublicShopDTO(shop))
}

type UpdateShopRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Phone       string `json:"phone"`
	Hours       string `json:"hours"`
	Logo        string `json:"logo"`
	Status      int    `json:"status"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	// Reward config — pointers so 0 is a valid, persisted value
	RewardRateSelf          *float64 `json:"reward_rate_self"`
	RewardRateLevel1        *float64 `json:"reward_rate_level1"`
	RewardRateLevel2        *float64 `json:"reward_rate_level2"`
	RewardCeiling           *float64 `json:"reward_ceiling"`
	RewardExcludeCategories *string  `json:"reward_exclude_categories"` // jsonb array string
}

func UpdateShop(c *gin.Context) {
	shopID := c.Param("id")

	// Verify ownership — merchant may only update their own shop
	merchantID := c.GetUint("user_id")
	if !merchantOwnsShop(shopID, merchantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Reward rates are fractions — reject out-of-range values (they feed the reward engine).
	for _, r := range []*float64{req.RewardRateSelf, req.RewardRateLevel1, req.RewardRateLevel2, req.RewardCeiling} {
		if r != nil && (*r < 0 || *r > 1) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "reward rates must be between 0 and 1"})
			return
		}
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Hours != "" {
		updates["hours"] = req.Hours
	}
	if req.Logo != "" {
		updates["logo"] = req.Logo
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}
	if req.RewardRateSelf != nil {
		updates["reward_rate_self"] = *req.RewardRateSelf
	}
	if req.RewardRateLevel1 != nil {
		updates["reward_rate_level1"] = *req.RewardRateLevel1
	}
	if req.RewardRateLevel2 != nil {
		updates["reward_rate_level2"] = *req.RewardRateLevel2
	}
	if req.RewardCeiling != nil {
		updates["reward_ceiling"] = *req.RewardCeiling
	}
	if req.RewardExcludeCategories != nil {
		updates["reward_exclude_categories"] = *req.RewardExcludeCategories
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}

	if err := config.DB.Model(&models.Shop{}).Where("id = ?", shopID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// ResolveDeliveryShop returns the shop a delivery order should be placed
// against. Single-shop reality: returns the active shop ordered by id.
// Forward stub for nearest-shop ranking once multiple shops + geocoding exist.
func ResolveDeliveryShop(c *gin.Context) {
	var shop models.Shop
	// Prefer an active shop that actually has active products (don't route
	// delivery to an empty menu); fall back to any active shop.
	withProducts := config.DB.Model(&models.Product{}).
		Select("DISTINCT shop_id").Where("status = 1")
	err := config.DB.Where("status = 1 AND id IN (?)", withProducts).
		Order("id").First(&shop).Error
	if err != nil {
		if err2 := config.DB.Where("status = 1").Order("id").First(&shop).Error; err2 != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no available shop"})
			return
		}
	}
	c.JSON(http.StatusOK, toPublicShopDTO(shop))
}

// merchantOwnsShop reports whether shopID belongs to merchantID.
func merchantOwnsShop(shopID interface{}, merchantID uint) bool {
	var shop models.Shop
	return config.DB.Where("id = ? AND merchant_id = ?", shopID, merchantID).First(&shop).Error == nil
}
