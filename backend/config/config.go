package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	WeChat   WeChatConfig
	R2       R2Config
	Shansong ShansongConfig
}

type ShansongConfig struct {
	ClientID  string `mapstructure:"clientid"`
	AppSecret string `mapstructure:"appsecret"`
	BaseURL   string `mapstructure:"base_url"`   // open.s.bingex.com（测试）/ open.ishansong.com（生产）
	NotifyURL string `mapstructure:"notify_url"` // 闪送状态回调地址
}

type R2Config struct {
	AccountID       string `mapstructure:"account_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Bucket          string `mapstructure:"bucket"`
	PublicBase      string `mapstructure:"public_base"` // e.g. https://bestluckbox.com
}

type ServerConfig struct {
	Port           string
	Mode           string
	BaseURL        string   // Public base URL for QR code generation (e.g. https://example.com)
	AllowedOrigins []string // CORS allowlist for browser clients (the merchant admin SPA)
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
	viper.SetDefault("shansong.base_url", "https://open.s.bingex.com")

	// config.yaml is optional — Railway uses env vars only
	viper.ReadInConfig()

	var cfg Config
	viper.Unmarshal(&cfg)

	// Env vars override config file (for Docker / Railway)
	cfg.Server.Port = getEnv("PORT", cfg.Server.Port)
	cfg.Server.Mode = getEnv("MODE", cfg.Server.Mode)
	cfg.Server.BaseURL = getEnv("BASE_URL", cfg.Server.BaseURL)
	// CORS allowlist: comma-separated origins. Defaults cover the deployed admin SPA and local dev.
	for _, o := range strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "https://table-order-admin.pages.dev,http://localhost:5173"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			cfg.Server.AllowedOrigins = append(cfg.Server.AllowedOrigins, o)
		}
	}
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

	// Shansong env vars (Railway) — credentials must never live in the repo
	cfg.Shansong.ClientID = getEnv("SHANSONG_CLIENT_ID", cfg.Shansong.ClientID)
	cfg.Shansong.AppSecret = getEnv("SHANSONG_APP_SECRET", cfg.Shansong.AppSecret)
	cfg.Shansong.BaseURL = getEnv("SHANSONG_BASE_URL", cfg.Shansong.BaseURL)
	cfg.Shansong.NotifyURL = getEnv("SHANSONG_NOTIFY_URL", cfg.Shansong.NotifyURL)

	// Cloudflare R2 env vars (Railway)
	cfg.R2.AccountID = getEnv("R2_ACCOUNT_ID", cfg.R2.AccountID)
	cfg.R2.AccessKeyID = getEnv("R2_ACCESS_KEY_ID", cfg.R2.AccessKeyID)
	cfg.R2.SecretAccessKey = getEnv("R2_SECRET_ACCESS_KEY", cfg.R2.SecretAccessKey)
	cfg.R2.Bucket = getEnv("R2_BUCKET", cfg.R2.Bucket)
	cfg.R2.PublicBase = getEnv("R2_PUBLIC_BASE", cfg.R2.PublicBase)

	return &cfg, nil
}
