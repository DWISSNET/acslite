package config_test

import (
	"os"
	"testing"

	"github.com/DWISSNET/acsgo/pkg/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure no interference from environment
	os.Unsetenv("PORT")
	os.Unsetenv("CWMP_PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("JWT_SECRET")

	cfg := config.Load()

	if cfg.WebPort != 7548 {
		t.Errorf("default WebPort: want 7548, got %d", cfg.WebPort)
	}
	if cfg.CWMPPort != 7547 {
		t.Errorf("default CWMPPort: want 7547, got %d", cfg.CWMPPort)
	}
	if cfg.Env != "development" {
		t.Errorf("default Env: want development, got %q", cfg.Env)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default LogLevel: want info, got %q", cfg.LogLevel)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("CWMP_PORT", "7547")
	os.Setenv("ENV", "production")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CWMP_PORT")
		os.Unsetenv("ENV")
	}()

	cfg := config.Load()
	if cfg.WebPort != 8080 {
		t.Errorf("WebPort from env: want 8080, got %d", cfg.WebPort)
	}
	if cfg.Env != "production" {
		t.Errorf("Env from env: want production, got %q", cfg.Env)
	}
}
