package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/example/table-order/models"
)

var DB *gorm.DB

func InitDB(cfg DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	log.Println("Database connected")
	return nil
}

func MigrateDB() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Merchant{},
		&models.Shop{},
		&models.TableQRCode{},
		&models.Order{},
		&models.OrderItem{},
		&models.Product{},
		&models.InviteRelation{},
		&models.WalletLog{},
	)
}