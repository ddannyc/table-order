package main

import (
	"log"

	"github.com/example/table-order/api/router"
	"github.com/example/table-order/config"
	"github.com/example/table-order/services"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}
	config.AppConfig = cfg

	log.Printf("BaseURL=%s", cfg.Server.BaseURL)

	if err := config.InitDB(cfg.Database); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	if err := config.MigrateDB(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	if err := config.InitWechatPay(cfg.WeChat); err != nil {
		log.Fatalf("WeChat Pay init failed: %v", err)
	}

	config.InitR2(cfg.R2)

	services.InitShansongClient(cfg.Shansong.ClientID, cfg.Shansong.AppSecret, cfg.Shansong.ShopID, cfg.Shansong.BaseURL)

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// Pin which upstream may set X-Forwarded-For; otherwise gin.Default() trusts all
	// proxies and ClientIP() (the rate-limiter key) honors a client-spoofed header.
	// Empty list = trust none → ClientIP() is the real TCP peer.
	if err := router.ConfigureTrustedProxies(r, cfg.Server.TrustedProxies); err != nil {
		log.Fatalf("ConfigureTrustedProxies failed: %v", err)
	}
	if len(cfg.Server.TrustedProxies) == 0 {
		log.Println("WARNING: TRUSTED_PROXIES unset — trusting no proxy; behind a load balancer set it to the edge IP/CIDR so client IPs resolve correctly")
	}

	router.Setup(r)

	log.Printf("Server started on :%s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
