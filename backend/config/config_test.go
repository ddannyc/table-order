package config

import (
	"os"
	"testing"
)

// JWT secret must be injectable via env: Railway has no config.yaml, so without
// an env binding the secret is empty — tokens get signed with an empty key and
// the quote-token secret fails closed, blocking delivery ("外卖配送暂未开通").
func TestLoadConfig_JWTSecretFromEnv(t *testing.T) {
	os.Setenv("JWT_SECRET", "super-secret-jwt-key")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := LoadConfig(".")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.JWT.Secret != "super-secret-jwt-key" {
		t.Errorf("expected JWT secret from env, got %q", cfg.JWT.Secret)
	}
}

// Shansong credentials must be injectable via environment variables (Railway has
// no config.yaml) and must never need to live in the repo.
func TestLoadConfig_ShansongFromEnv(t *testing.T) {
	os.Setenv("SHANSONG_CLIENT_ID", "cid-test-123")
	os.Setenv("SHANSONG_APP_SECRET", "secret-test-456")
	os.Setenv("SHANSONG_BASE_URL", "https://open.s.bingex.com")
	defer os.Unsetenv("SHANSONG_CLIENT_ID")
	defer os.Unsetenv("SHANSONG_APP_SECRET")
	defer os.Unsetenv("SHANSONG_BASE_URL")

	cfg, err := LoadConfig(".")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.Shansong.ClientID != "cid-test-123" {
		t.Errorf("expected ClientID from env, got %q", cfg.Shansong.ClientID)
	}
	if cfg.Shansong.AppSecret != "secret-test-456" {
		t.Errorf("expected AppSecret from env, got %q", cfg.Shansong.AppSecret)
	}
	if cfg.Shansong.BaseURL != "https://open.s.bingex.com" {
		t.Errorf("expected BaseURL from env, got %q", cfg.Shansong.BaseURL)
	}
}
