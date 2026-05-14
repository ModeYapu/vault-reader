package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Addr != ":3000" {
		t.Errorf("expected default addr :3000, got %s", cfg.Addr)
	}
	if cfg.Readonly != true {
		t.Error("expected readonly true by default")
	}
}

func TestConfigFromArgs(t *testing.T) {
	args := []string{
		"--vault", "/opt/vault",
		"--data", "/opt/data",
		"--addr", ":8080",
	}
	cfg, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cfg.VaultDir != "/opt/vault" {
		t.Errorf("expected vault /opt/vault, got %s", cfg.VaultDir)
	}
	if cfg.DataDir != "/opt/data" {
		t.Errorf("expected data /opt/data, got %s", cfg.DataDir)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("expected addr :8080, got %s", cfg.Addr)
	}
}

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("VAULT_DIR", "/env/vault")
	os.Setenv("DATA_DIR", "/env/data")
	os.Setenv("ADDR", ":9090")
	defer os.Unsetenv("VAULT_DIR")
	defer os.Unsetenv("DATA_DIR")
	defer os.Unsetenv("ADDR")

	cfg := Default()
	cfg.ApplyEnv()

	if cfg.VaultDir != "/env/vault" {
		t.Errorf("expected vault /env/vault, got %s", cfg.VaultDir)
	}
	if cfg.DataDir != "/env/data" {
		t.Errorf("expected data /env/data, got %s", cfg.DataDir)
	}
	if cfg.Addr != ":9090" {
		t.Errorf("expected addr :9090, got %s", cfg.Addr)
	}
}

func TestConfigMissingVault(t *testing.T) {
	args := []string{}
	_, err := ParseArgs(args)
	if err == nil {
		t.Error("expected error when vault dir is missing")
	}
}

func TestConfigEnvFallbackToArgs(t *testing.T) {
	// Args take precedence over env
	os.Setenv("VAULT_DIR", "/env/vault")
	defer os.Unsetenv("VAULT_DIR")

	args := []string{"--vault", "/args/vault"}
	cfg, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	// Args should win
	if cfg.VaultDir != "/args/vault" {
		t.Errorf("expected args vault /args/vault, got %s", cfg.VaultDir)
	}
}
