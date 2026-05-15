package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Addr != ":3000" {
		t.Errorf("expected default addr :3000, got %s", cfg.Addr)
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
	t.Setenv("VAULT_DIR", "/env/vault")
	t.Setenv("DATA_DIR", "/env/data")
	t.Setenv("ADDR", ":9090")

	cfg := Default()
	cfg.ApplyEnv(false)

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
	t.Setenv("VAULT_DIR", "/env/vault")

	args := []string{"--vault", "/args/vault"}
	cfg, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	if cfg.VaultDir != "/args/vault" {
		t.Errorf("expected args vault /args/vault, got %s", cfg.VaultDir)
	}
}

func TestConfigDefaultDataDir(t *testing.T) {
	args := []string{"--vault", "/my/vault"}
	cfg, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs failed: %v", err)
	}
	expected := filepath.Join("/my/vault", ".vault-reader-data")
	if cfg.DataDir != expected {
		t.Errorf("expected default data dir %s, got %s", expected, cfg.DataDir)
	}
}

func TestConfigInvalidFlag(t *testing.T) {
	args := []string{"--unknown-flag"}
	_, err := ParseArgs(args)
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestStringFlagNilGuard(t *testing.T) {
	f := stringFlag{} // nil target and set
	err := f.Set("value")
	if err == nil {
		t.Error("expected error when Set called on uninitialized stringFlag")
	}
}
