package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/api/router"
	"github.com/example/table-order/config"
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

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	router.Setup(r)

	log.Printf("Server started on :%s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
