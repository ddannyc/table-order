package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/example/table-order/models"
)

var DB *gorm.DB

func buildDSN(cfg DatabaseConfig) string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
}

func InitDB(cfg DatabaseConfig) error {
	dsn := buildDSN(cfg)

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
		&models.RewardLog{},
	)
}
