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

// TrustedProxies must be injectable via env so the deployment can pin which
// upstream (Railway's edge) is allowed to set X-Forwarded-For. Without it the
// rate limiter keys on a client-spoofable IP.
func TestLoadConfig_TrustedProxiesFromEnv(t *testing.T) {
	os.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 172.16.0.1")
	defer os.Unsetenv("TRUSTED_PROXIES")

	cfg, err := LoadConfig(".")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(cfg.Server.TrustedProxies) != 2 ||
		cfg.Server.TrustedProxies[0] != "10.0.0.0/8" ||
		cfg.Server.TrustedProxies[1] != "172.16.0.1" {
		t.Errorf("expected trusted proxies parsed+trimmed from env, got %#v", cfg.Server.TrustedProxies)
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
