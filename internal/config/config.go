package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all application configuration.
type Config struct {
	VaultDir string
	DataDir  string
	Addr     string
	BaseURL  string
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Addr: ":3000",
	}
}

// ParseArgs parses command-line arguments and returns a validated Config.
func ParseArgs(args []string) (*Config, error) {
	fs := flag.NewFlagSet("vault-reader", flag.ContinueOnError)
	cfg := Default()

	addrSet := false
	fs.StringVar(&cfg.VaultDir, "vault", "", "Path to Obsidian Vault directory")
	fs.StringVar(&cfg.DataDir, "data", "", "Path to data directory for index database")
	fs.Var(&stringFlag{target: &cfg.Addr, set: &addrSet}, "addr", "Listen address")
	fs.StringVar(&cfg.BaseURL, "base-url", "", "Optional base URL for reverse proxy")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Apply env vars as fallback for empty values
	cfg.ApplyEnv(addrSet)

	if cfg.VaultDir == "" {
		return nil, fmt.Errorf("vault directory is required: use --vault flag or VAULT_DIR env")
	}

	// Default data dir to <vault>/.vault-reader-data if not set
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(cfg.VaultDir, ".vault-reader-data")
	}

	return cfg, nil
}

// ApplyEnv fills empty config fields from environment variables.
func (c *Config) ApplyEnv(addrSet bool) {
	if v := os.Getenv("VAULT_DIR"); v != "" && c.VaultDir == "" {
		c.VaultDir = v
	}
	if v := os.Getenv("DATA_DIR"); v != "" && c.DataDir == "" {
		c.DataDir = v
	}
	if v := os.Getenv("ADDR"); v != "" && !addrSet {
		c.Addr = v
	}
	if v := os.Getenv("BASE_URL"); v != "" && c.BaseURL == "" {
		c.BaseURL = v
	}
}

// stringFlag tracks whether the flag was explicitly set.
type stringFlag struct {
	target *string
	set    *bool
}

func (f *stringFlag) String() string {
	if f.target == nil {
		return ""
	}
	return *f.target
}

func (f *stringFlag) Set(v string) error {
	*f.target = v
	*f.set = true
	return nil
}
