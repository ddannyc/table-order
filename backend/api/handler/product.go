package handler

import (
	"net/http"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/gin-gonic/gin"
)

func GetShopProducts(c *gin.Context) {
	shopID := c.Param("id")

	var products []models.Product
	config.DB.Where("shop_id = ? AND status = 1", shopID).Order("category, id").Find(&products)

	c.JSON(http.StatusOK, products)
}

type CreateProductRequest struct {
	ShopID      uint    `json:"shop_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
	Status      int     `json:"status"`
}

func CreateProduct(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Verify shop belongs to merchant
	var shop models.Shop
	if err := config.DB.Where("id = ? AND merchant_id = ?", req.ShopID, merchantID).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	status := req.Status
	if status == 0 {
		status = 1
	}

	product := models.Product{
		ShopID:      req.ShopID,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Image:       req.Image,
		Category:    req.Category,
		Status:      status,
	}

	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create product failed"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func GetMerchantProducts(c *gin.Context) {
	merchantID := c.GetUint("user_id")

	var shops []models.Shop
	config.DB.Where("merchant_id = ?", merchantID).Find(&shops)

	shopIDs := make([]uint, len(shops))
	for i, s := range shops {
		shopIDs[i] = s.ID
	}

	var products []models.Product
	if len(shopIDs) > 0 {
		config.DB.Where("shop_id IN ?", shopIDs).Order("shop_id, category, id").Find(&products)
	}

	c.JSON(http.StatusOK, products)
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
	Status      *int    `json:"status"` // pointer so 0 (下架) is distinguishable from "not provided"
}

func UpdateProduct(c *gin.Context) {
	merchantID := c.GetUint("user_id")
	productID := c.Param("id")

	// Get product and verify ownership
	var product models.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	var shop models.Shop
	if err := config.DB.Where("id = ? AND merchant_id = ?", product.ShopID, merchantID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Image != "" {
		updates["image"] = req.Image
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&product).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteProduct(c *gin.Context) {
	merchantID := c.GetUint("user_id")
	productID := c.Param("id")

	var product models.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	var shop models.Shop
	if err := config.DB.Where("id = ? AND merchant_id = ?", product.ShopID, merchantID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	config.DB.Delete(&product)

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
