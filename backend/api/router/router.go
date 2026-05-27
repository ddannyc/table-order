package router

import (
	"github.com/gin-gonic/gin"
	"github.com/example/table-order/api/handler"
	"github.com/example/table-order/middleware"
)

func Setup(r *gin.Engine) {
	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// Auth (public)
	auth := api.Group("/auth")
	{
		auth.POST("/login", handler.Login)
	}

	// User (authenticated)
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleware())
	{
		user.GET("/profile", handler.GetProfile)
		user.PUT("/profile", handler.UpdateProfile)
	}

	// Wallet (authenticated)
	wallet := api.Group("/wallet")
	wallet.Use(middleware.AuthMiddleware())
	{
		wallet.GET("/balance", handler.GetBalance)
		wallet.GET("/logs", handler.GetWalletLogs)
	}

	// Shop (public for scanning, but merchant management requires auth)
	shops := api.Group("/shops")
	{
		shops.GET("/:id", handler.GetShop)
		shops.GET("/:id/products", handler.GetShopProducts)
		shops.GET("/:id/qrcodes", handler.ListQRCodes)
		shops.POST("/:id/qrcodes", handler.GenerateQR) // TODO: should require merchant auth
	}

	// Orders (authenticated)
	orders := api.Group("/orders")
	orders.Use(middleware.AuthMiddleware())
	{
		orders.POST("", handler.CreateOrder)
		orders.GET("", handler.GetOrders)
		orders.GET("/:id", handler.GetOrder)
	}

	// Invite (authenticated)
	invite := api.Group("/invites")
	invite.Use(middleware.AuthMiddleware())
	{
		invite.POST("/bind", handler.BindInviteCode)
		invite.GET("/stats", handler.GetInviteStats)
		invite.GET("/qrcode", handler.GenerateInviteQR)
	}

	// Merchant - public endpoints
	merchantPublic := api.Group("/merchant")
	{
		merchantPublic.POST("/login", handler.MerchantLogin)
		merchantPublic.POST("/register", handler.RegisterMerchant)
	}

	// Merchant (merchant auth required)
	merchant := api.Group("/merchant")
	merchant.Use(middleware.AuthMiddleware(), middleware.MerchantAuth())
	{
		merchant.GET("/shops", handler.GetShops)
		merchant.POST("/shops", handler.CreateShop)
		merchant.PUT("/shops/:id", handler.UpdateShop)
		merchant.GET("/orders", handler.GetMerchantOrders)
		merchant.POST("/orders", handler.CreateMerchantOrder)
		merchant.GET("/stats", handler.GetMerchantStats)
		merchant.GET("/dashboard", handler.GetMerchantDashboard)
		merchant.GET("/products", handler.GetMerchantProducts)
		merchant.POST("/products", handler.CreateProduct)
		merchant.PUT("/products/:id", handler.UpdateProduct)
		merchant.DELETE("/products/:id", handler.DeleteProduct)
	}

	// Admin (admin auth required)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminAuth())
	{
		admin.GET("/users", handler.GetAllUsers)
		admin.PUT("/users/:id/ban", handler.BanUser)
		admin.GET("/merchants", handler.GetAllMerchants)
		admin.PUT("/config", handler.UpdateGlobalConfig)
	}
}