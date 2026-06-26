package handler

import (
	"net/http"

	"github.com/example/table-order/config"
	"github.com/example/table-order/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetShopProducts(c *gin.Context) {
	shopID := c.Param("id")

	var products []models.Product
	config.DB.Preload("Specs").Where("shop_id = ? AND status = 1", shopID).Order("category, id").Find(&products)

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
		config.DB.Preload("Specs").Where("shop_id IN ?", shopIDs).Order("shop_id, category, id").Find(&products)
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

// === Product specs (variants) ===

// specOwnedByMerchant loads a spec and verifies it belongs to the merchant via
// product -> shop. Returns false (and writes the error response) if not.
func specOwnedByMerchant(c *gin.Context, specID string, merchantID uint, spec *models.ProductSpec) bool {
	if err := config.DB.First(spec, specID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "spec not found"})
		return false
	}
	var product models.Product
	if err := config.DB.First(&product, spec.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return false
	}
	var shop models.Shop
	if err := config.DB.Where("id = ? AND merchant_id = ?", product.ShopID, merchantID).First(&shop).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return false
	}
	return true
}

type CreateProductSpecRequest struct {
	Name   string  `json:"name" binding:"required"`
	Price  float64 `json:"price" binding:"required,gt=0"`
	Status *int    `json:"status"` // pointer so explicit 0 (下架) is distinguishable from "not provided"
}

func CreateProductSpec(c *gin.Context) {
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

	var req CreateProductSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	status := 1 // default 上架
	if req.Status != nil {
		status = *req.Status
	}
	spec := models.ProductSpec{
		ProductID: product.ID,
		Name:      req.Name,
		Price:     req.Price,
		Status:    status,
	}
	// The model's default:1 tag overrides a zero Status on insert, so an
	// intentional 下架 (0) needs a follow-up write. Do both in one transaction
	// so a failure rolls back rather than leaving a spec live at the wrong status.
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&spec).Error; err != nil {
			return err
		}
		if status == 0 && spec.Status != 0 {
			if err := tx.Model(&spec).Update("status", 0).Error; err != nil {
				return err
			}
			spec.Status = 0
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create spec failed"})
		return
	}

	c.JSON(http.StatusOK, spec)
}

type UpdateProductSpecRequest struct {
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
	Status *int    `json:"status"`
}

func UpdateProductSpec(c *gin.Context) {
	merchantID := c.GetUint("user_id")
	specID := c.Param("id")

	var spec models.ProductSpec
	if !specOwnedByMerchant(c, specID, merchantID, &spec) {
		return
	}

	var req UpdateProductSpecRequest
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
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) > 0 {
		if err := config.DB.Model(&spec).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteProductSpec(c *gin.Context) {
	merchantID := c.GetUint("user_id")
	specID := c.Param("id")

	var spec models.ProductSpec
	if !specOwnedByMerchant(c, specID, merchantID, &spec) {
		return
	}

	config.DB.Delete(&spec)

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
