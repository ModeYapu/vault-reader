package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all application configuration.
type Config struct {
	VaultDir    string
	DataDir     string
	Addr        string
	BaseURL     string
	Readonly    bool
	AuthEnabled bool
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Addr:     ":3000",
		Readonly: true,
	}
}

// ParseArgs parses command-line arguments and returns a validated Config.
func ParseArgs(args []string) (*Config, error) {
	fs := flag.NewFlagSet("vault-reader", flag.ContinueOnError)
	cfg := Default()

	fs.StringVar(&cfg.VaultDir, "vault", "", "Path to Obsidian Vault directory")
	fs.StringVar(&cfg.DataDir, "data", "", "Path to data directory for index database")
	fs.StringVar(&cfg.Addr, "addr", ":3000", "Listen address")
	fs.StringVar(&cfg.BaseURL, "base-url", "", "Optional base URL for reverse proxy")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Apply env vars as fallback for empty values
	cfg.ApplyEnv()

	if cfg.VaultDir == "" {
		return nil, fmt.Errorf("vault directory is required: use --vault flag or VAULT_DIR env")
	}

	// Default data dir to <vault>/.vault-reader-data if not set
	if cfg.DataDir == "" {
		cfg.DataDir = cfg.VaultDir + "/.vault-reader-data"
	}

	return cfg, nil
}

// ApplyEnv fills empty config fields from environment variables.
func (c *Config) ApplyEnv() {
	if v := os.Getenv("VAULT_DIR"); v != "" && c.VaultDir == "" {
		c.VaultDir = v
	}
	if v := os.Getenv("DATA_DIR"); v != "" && c.DataDir == "" {
		c.DataDir = v
	}
	if v := os.Getenv("ADDR"); v != "" && c.Addr == ":3000" {
		c.Addr = v
	}
	if v := os.Getenv("BASE_URL"); v != "" && c.BaseURL == "" {
		c.BaseURL = v
	}
	if v := os.Getenv("AUTH_ENABLED"); v == "true" {
		c.AuthEnabled = true
	}
}
