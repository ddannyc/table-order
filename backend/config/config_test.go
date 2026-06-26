package config

import (
	"os"
	"testing"
)

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
