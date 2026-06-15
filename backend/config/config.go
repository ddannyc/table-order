package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	WeChat   WeChatConfig
}

type ServerConfig struct {
	Port    string
	Mode    string
	BaseURL string // Public base URL for QR code generation (e.g. https://example.com)
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type WeChatConfig struct {
	AppID     string `mapstructure:"appid"`
	AppSecret string `mapstructure:"appsecret"`
	// WeChat Pay V3
	MchID                      string `mapstructure:"mch_id"`
	MchCertificateSerialNumber string `mapstructure:"mch_certificate_serial_number"`
	MchPrivateKeyPath          string `mapstructure:"mch_private_key_path"`
	MchAPIv3Key                string `mapstructure:"mch_api_v3_key"`
	PayNotifyURL               string `mapstructure:"pay_notify_url"`
	// Public key scheme (required since 2026)
	WechatPayPublicKeyID      string `mapstructure:"wechatpay_public_key_id"`
	WechatPayPublicKeyPath    string `mapstructure:"wechatpay_public_key_path"`
	WechatPayPublicKeyContent string `mapstructure:"wechatpay_public_key_content"`
	MchPrivateKeyContent      string `mapstructure:"mch_private_key_content"`
	// Mini-program env version for QR codes and URL Schemes: "develop", "trial", or "release"
	EnvVersion string `mapstructure:"env_version"`
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

var AppConfig *Config

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.baseurl", "")
	viper.SetDefault("jwt.expire_hours", 720)

	// config.yaml is optional — Railway uses env vars only
	viper.ReadInConfig()

	var cfg Config
	viper.Unmarshal(&cfg)

	// Env vars override config file (for Docker / Railway)
	cfg.Server.Port = getEnv("PORT", cfg.Server.Port)
	cfg.Server.Mode = getEnv("MODE", cfg.Server.Mode)
	cfg.Server.BaseURL = getEnv("BASE_URL", cfg.Server.BaseURL)
	cfg.Database.Host = getEnv("PGHOST", getEnv("DB_HOST", cfg.Database.Host))
	cfg.Database.Port = getEnv("PGPORT", getEnv("DB_PORT", cfg.Database.Port))
	cfg.Database.User = getEnv("POSTGRES_USER", getEnv("DB_USER", cfg.Database.User))
	cfg.Database.Password = getEnv("POSTGRES_PASSWORD", getEnv("DB_PASSWORD", cfg.Database.Password))
	cfg.Database.DBName = getEnv("POSTGRES_DB", getEnv("DB_NAME", cfg.Database.DBName))

	// WeChat env vars (Railway)
	cfg.WeChat.AppID = getEnv("WECHAT_APPID", cfg.WeChat.AppID)
	cfg.WeChat.AppSecret = getEnv("WECHAT_APPSECRET", cfg.WeChat.AppSecret)
	cfg.WeChat.MchID = getEnv("WECHAT_MCHID", cfg.WeChat.MchID)
	cfg.WeChat.MchCertificateSerialNumber = getEnv("WECHAT_MCH_CERT_SERIAL", cfg.WeChat.MchCertificateSerialNumber)
	cfg.WeChat.MchPrivateKeyPath = getEnv("WECHAT_MCH_PRIVATE_KEY_PATH", cfg.WeChat.MchPrivateKeyPath)
	cfg.WeChat.MchAPIv3Key = getEnv("WECHAT_MCH_API_V3_KEY", cfg.WeChat.MchAPIv3Key)
	cfg.WeChat.PayNotifyURL = getEnv("WECHAT_PAY_NOTIFY_URL", cfg.WeChat.PayNotifyURL)
	cfg.WeChat.WechatPayPublicKeyID = getEnv("WECHATPAY_PUBLIC_KEY_ID", cfg.WeChat.WechatPayPublicKeyID)
	cfg.WeChat.WechatPayPublicKeyPath = getEnv("WECHATPAY_PUBLIC_KEY_PATH", cfg.WeChat.WechatPayPublicKeyPath)
	cfg.WeChat.WechatPayPublicKeyContent = getEnv("WECHATPAY_PUBLIC_KEY_CONTENT", cfg.WeChat.WechatPayPublicKeyContent)
	cfg.WeChat.MchPrivateKeyContent = getEnv("WECHAT_MCH_PRIVATE_KEY_CONTENT", cfg.WeChat.MchPrivateKeyContent)
	cfg.WeChat.EnvVersion = getEnv("WECHAT_ENV_VERSION", cfg.WeChat.EnvVersion)

	return &cfg, nil
}
