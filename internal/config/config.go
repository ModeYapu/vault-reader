package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// ConfigFile represents the YAML configuration file structure.
type ConfigFile struct {
	Vault     string `yaml:"vault"`
	DataDir   string `yaml:"data_dir"`
	Addr      string `yaml:"addr"`
	BaseURL   string `yaml:"base_url"`
	Auth      *AuthConfig `yaml:"auth"`
	RateLimit *RateLimitConfig `yaml:"rate_limit"`
	CORS      *CORSConfig `yaml:"cors"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Requests int    `yaml:"requests"`
	Window   string `yaml:"window"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	Origins string `yaml:"origins"`
}

// Config holds all application configuration.
type Config struct {
	VaultDir     string
	DataDir      string
	Addr         string
	BaseURL      string
	AuthUsername string
	AuthPassword string
	RateLimit    int
	RateWindow   string
	CORSOrigins  string
	configPath   string // Path to config file for hot reload
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Addr: ":3000",
	}
}

// ParseArgs parses command-line arguments and returns a validated Config.
// It also attempts to load from config.yaml if present.
func ParseArgs(args []string) (*Config, error) {
	fs := flag.NewFlagSet("vault-reader", flag.ContinueOnError)
	cfg := Default()

	addrSet := false
	configFile := ""
	fs.StringVar(&configFile, "config", "", "Path to config YAML file")
	fs.StringVar(&cfg.VaultDir, "vault", "", "Path to Obsidian Vault directory")
	fs.StringVar(&cfg.DataDir, "data", "", "Path to data directory for index database")
	fs.Var(&stringFlag{target: &cfg.Addr, set: &addrSet}, "addr", "Listen address")
	fs.StringVar(&cfg.BaseURL, "base-url", "", "Optional base URL for reverse proxy")
	fs.StringVar(&cfg.AuthUsername, "auth-username", "", "Basic auth username")
	fs.StringVar(&cfg.AuthPassword, "auth-password", "", "Basic auth password")
	fs.IntVar(&cfg.RateLimit, "rate-limit", 0, "Rate limit requests per window (0 = disabled)")
	fs.StringVar(&cfg.RateWindow, "rate-window", "1m", "Rate limit time window (e.g., 1m, 1h)")
	fs.StringVar(&cfg.CORSOrigins, "cors-origins", "*", "CORS allowed origins (comma-separated, * for all)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Try to load config file
	if configFile == "" {
		// Check default locations
		for _, path := range []string{"config.yaml", "config.yml", ".vault-readerrc"} {
			if _, err := os.Stat(path); err == nil {
				configFile = path
				break
			}
		}
	}

	if configFile != "" {
		if err := cfg.LoadFromFile(configFile); err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
		cfg.configPath = configFile
	}

	// Apply env vars as fallback for empty values
	cfg.ApplyEnv(addrSet)

	// Command-line flags override config file
	if !addrSet && cfg.Addr == "" {
		cfg.Addr = ":3000"
	}

	if cfg.VaultDir == "" {
		return nil, fmt.Errorf("vault directory is required: use --vault flag, VAULT_DIR env, or config file")
	}

	// Default data dir to <vault>/.vault-reader-data if not set
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(cfg.VaultDir, ".vault-reader-data")
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a YAML file.
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cf ConfigFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return fmt.Errorf("parse YAML: %w", err)
	}

	// Only set fields that are empty (allow override by flags/env)
	if cf.Vault != "" && c.VaultDir == "" {
		c.VaultDir = cf.Vault
	}
	if cf.DataDir != "" && c.DataDir == "" {
		c.DataDir = cf.DataDir
	}
	if cf.Addr != "" && c.Addr == "" {
		c.Addr = cf.Addr
	}
	if cf.BaseURL != "" && c.BaseURL == "" {
		c.BaseURL = cf.BaseURL
	}
	if cf.Auth != nil {
		if cf.Auth.Username != "" && c.AuthUsername == "" {
			c.AuthUsername = cf.Auth.Username
		}
		if cf.Auth.Password != "" && c.AuthPassword == "" {
			c.AuthPassword = cf.Auth.Password
		}
	}
	if cf.RateLimit != nil {
		if cf.RateLimit.Requests > 0 && c.RateLimit == 0 {
			c.RateLimit = cf.RateLimit.Requests
		}
		if cf.RateLimit.Window != "" && c.RateWindow == "" {
			c.RateWindow = cf.RateLimit.Window
		}
	}
	if cf.CORS != nil && cf.CORS.Origins != "" && c.CORSOrigins == "" {
		c.CORSOrigins = cf.CORS.Origins
	}

	c.configPath = path
	return nil
}

// Reload reloads configuration from the original config file path.
// Returns true if configuration was actually reloaded.
func (c *Config) Reload() (bool, error) {
	if c.configPath == "" {
		return false, nil
	}

	info, err := os.Stat(c.configPath)
	if err != nil {
		return false, fmt.Errorf("stat config file: %w", err)
	}

	// Create a temporary config to load new values
	newCfg := &Config{}
	if err := newCfg.LoadFromFile(c.configPath); err != nil {
		return false, fmt.Errorf("reload config: %w", err)
	}

	// Update current config with new values
	c.VaultDir = newCfg.VaultDir
	c.DataDir = newCfg.DataDir
	c.Addr = newCfg.Addr
	c.BaseURL = newCfg.BaseURL
	c.AuthUsername = newCfg.AuthUsername
	c.AuthPassword = newCfg.AuthPassword
	c.RateLimit = newCfg.RateLimit
	c.RateWindow = newCfg.RateWindow
	c.CORSOrigins = newCfg.CORSOrigins

	_ = info // Track mod time for future optimization
	return true, nil
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
	if v := os.Getenv("AUTH_USERNAME"); v != "" && c.AuthUsername == "" {
		c.AuthUsername = v
	}
	if v := os.Getenv("AUTH_PASSWORD"); v != "" && c.AuthPassword == "" {
		c.AuthPassword = v
	}
	if v := os.Getenv("RATE_LIMIT"); v != "" && c.RateLimit == 0 {
		if val, err := strconv.Atoi(v); err == nil {
			c.RateLimit = val
		}
	}
	if v := os.Getenv("RATE_WINDOW"); v != "" && c.RateWindow == "" {
		c.RateWindow = v
	}
	if v := os.Getenv("CORS_ORIGINS"); v != "" && c.CORSOrigins == "" {
		c.CORSOrigins = v
	}
}

// GetConfigPath returns the path to the config file if one was loaded.
func (c *Config) GetConfigPath() string {
	return c.configPath
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
	if f.target == nil || f.set == nil {
		return fmt.Errorf("stringFlag not initialized")
	}
	*f.target = v
	*f.set = true
	return nil
}
