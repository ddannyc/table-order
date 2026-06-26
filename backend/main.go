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

	services.InitShansongClient(cfg.Shansong.ClientID, cfg.Shansong.AppSecret, cfg.Shansong.BaseURL)

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	router.Setup(r)

	log.Printf("Server started on :%s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
