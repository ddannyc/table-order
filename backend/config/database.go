package config

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"os"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/example/table-order/models"
)

var DB *gorm.DB
var WxPayClient *core.Client

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
		Logger: logger.Default.LogMode(logger.Warn),
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

func InitWechatPay(cfg WeChatConfig) error {
	if cfg.MchID == "" || cfg.MchAPIv3Key == "" {
		log.Println("WeChat Pay not configured — skipping")
		return nil
	}

	ctx := context.Background()

	// Load merchant private key: prefer content from env var, fall back to file path
	var mchPrivateKey *rsa.PrivateKey
	var err error
	if cfg.MchPrivateKeyContent != "" {
		mchPrivateKey, err = utils.LoadPrivateKey(cfg.MchPrivateKeyContent)
	} else {
		mchPrivateKey, err = utils.LoadPrivateKeyWithPath(cfg.MchPrivateKeyPath)
	}
	if err != nil {
		return fmt.Errorf("load merchant private key: %w", err)
	}

	// Determine if public key scheme is configured
	publicKeyConfigured := cfg.WechatPayPublicKeyID != "" &&
		(cfg.WechatPayPublicKeyPath != "" || cfg.WechatPayPublicKeyContent != "")

	if publicKeyConfigured {
		// Public key scheme (new, required since 2026)
		var wechatPayPublicKey *rsa.PublicKey
		if cfg.WechatPayPublicKeyContent != "" {
			wechatPayPublicKey, err = utils.LoadPublicKey(cfg.WechatPayPublicKeyContent)
		} else {
			wechatPayPublicKey, err = utils.LoadPublicKeyWithPath(cfg.WechatPayPublicKeyPath)
		}
		if err != nil {
			return fmt.Errorf("load wechat pay public key: %w", err)
		}

		client, err := core.NewClient(ctx,
			option.WithWechatPayPublicKeyAuthCipher(
				cfg.MchID,
				cfg.MchCertificateSerialNumber,
				mchPrivateKey,
				cfg.WechatPayPublicKeyID,
				wechatPayPublicKey,
			),
		)
		if err != nil {
			return fmt.Errorf("new wechat pay client (public key): %w", err)
		}

		WxPayClient = client
		log.Println("WeChat Pay client initialized (public key scheme)")
		return nil
	}

	// Fallback: certificate auto-download scheme (legacy)
	downloader.MgrInstance().RegisterDownloaderWithPrivateKey(
		ctx, mchPrivateKey, cfg.MchCertificateSerialNumber, cfg.MchID, cfg.MchAPIv3Key,
	)

	client, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(
			cfg.MchID,
			cfg.MchCertificateSerialNumber,
			mchPrivateKey,
			cfg.MchAPIv3Key,
		),
	)
	if err != nil {
		return fmt.Errorf("new wechat pay client: %w", err)
	}

	WxPayClient = client
	log.Println("WeChat Pay client initialized (certificate scheme)")
	return nil
}
