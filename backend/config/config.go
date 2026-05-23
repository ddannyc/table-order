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
	Port string
	Mode string
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
	AppID     string
	AppSecret string
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
	viper.SetDefault("jwt.expire_hours", 720)

	// config.yaml is optional — Railway uses env vars only
	viper.ReadInConfig()

	var cfg Config
	viper.Unmarshal(&cfg)

	// Env vars override config file (for Docker / Railway)
	cfg.Server.Port = getEnv("PORT", cfg.Server.Port)
	cfg.Database.Host = getEnv("PGHOST", getEnv("DB_HOST", cfg.Database.Host))
	cfg.Database.Port = getEnv("PGPORT", getEnv("DB_PORT", cfg.Database.Port))
	cfg.Database.User = getEnv("POSTGRES_USER", getEnv("DB_USER", cfg.Database.User))
	cfg.Database.Password = getEnv("POSTGRES_PASSWORD", getEnv("DB_PASSWORD", cfg.Database.Password))
	cfg.Database.DBName = getEnv("POSTGRES_DB", getEnv("DB_NAME", cfg.Database.DBName))

	return &cfg, nil
}
