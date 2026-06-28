package router

import (
	"net/http"
	"slices"

	"github.com/example/table-order/api/handler"
	"github.com/example/table-order/config"
	"github.com/example/table-order/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	// CORS — restrict to the configured admin origins. The WeChat mini-program is
	// not a browser and ignores CORS, so tightening this only affects the SPA.
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && slices.Contains(config.AppConfig.Server.AllowedOrigins, origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
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

	// QR code scan redirect (public — accessed from WeChat browser when scanning table QR)
	r.GET("/scan", handler.ScanRedirect)

	// Static file serving — serve GP1H2hwSZU.txt at root path
	r.GET("/GP1H2hwSZU.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("bf96f49038d75fcf638c1bb9ab593d9a\n"))
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
		// QR code generation/listing is merchant-authenticated — see /merchant/shops/:id/qrcodes
	}

	// Delivery (外卖) — resolve the shop for a delivery order (single-shop stub)
	api.GET("/delivery/shop", handler.ResolveDeliveryShop)
	// Realtime delivery quote (authenticated — returns a signed fee token)
	api.POST("/delivery/quote", middleware.AuthMiddleware(), handler.DeliveryQuote)

	// Orders (authenticated)
	orders := api.Group("/orders")
	orders.Use(middleware.AuthMiddleware())
	{
		orders.POST("", handler.CreateOrder)
		orders.GET("", handler.GetOrders)
		orders.GET("/:id", handler.GetOrder)
	}

	// WeChat Pay notification callback (public — WeChat server calls this)
	api.POST("/orders/notify", handler.WechatPayNotify)

	// Shansong delivery status callback (public — verified by Shansong signature)
	api.POST("/shansong/callback", handler.ShansongCallback)

	// Invite (authenticated)
	invite := api.Group("/invites")
	invite.Use(middleware.AuthMiddleware())
	{
		invite.POST("/bind", handler.BindInviteCode)
		invite.GET("/stats", handler.GetInviteStats)
		invite.GET("/qrcode", handler.GenerateInviteQR)
	}

	// Reward (authenticated)
	reward := api.Group("/reward")
	reward.Use(middleware.AuthMiddleware())
	{
		reward.GET("/balance", handler.GetRewardBalance)
		reward.GET("/logs", handler.GetRewardLogs)
		reward.GET("/expiry-info", handler.GetRewardExpiryInfo)
	}

	// Auth - phone verification
	auth.POST("/verify-phone", middleware.AuthMiddleware(), handler.VerifyPhone)

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
		merchant.POST("/orders/:id/prepare", handler.PrepareOrder)
		merchant.GET("/stats", handler.GetMerchantStats)
		merchant.GET("/dashboard", handler.GetMerchantDashboard)
		merchant.GET("/products", handler.GetMerchantProducts)
		merchant.POST("/products", handler.CreateProduct)
		merchant.PUT("/products/:id", handler.UpdateProduct)
		merchant.DELETE("/products/:id", handler.DeleteProduct)
		merchant.POST("/products/:id/specs", handler.CreateProductSpec)
		merchant.PUT("/specs/:id", handler.UpdateProductSpec)
		merchant.DELETE("/specs/:id", handler.DeleteProductSpec)
		merchant.GET("/shops/:id/qrcodes", handler.ListMerchantQRCodes)
		merchant.POST("/shops/:id/qrcodes", handler.GenerateMerchantQR)
		merchant.POST("/upload", handler.UploadImage)
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
