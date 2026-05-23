package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/example/table-order/models"
)

var DB *gorm.DB

func buildDSN(cfg DatabaseConfig) string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		// Parse DATABASE_URL for individual fields if needed
		if strings.HasPrefix(dsn, "postgresql://") || strings.HasPrefix(dsn, "postgres://") {
			u, err := url.Parse(dsn)
			if err == nil {
				host := u.Hostname()
				port := u.Port()
				if port == "" {
					port = "5432"
				}
				password, _ := u.User.Password()
				user := u.User.Username()
				dbname := strings.TrimPrefix(u.Path, "/")
				sslmode := u.Query().Get("sslmode")
				if sslmode == "" {
					sslmode = "disable"
				}
				return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
					host, port, user, password, dbname, sslmode)
			}
		}
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
	)
}