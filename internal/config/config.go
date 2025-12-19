package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the runtime configuration
type Config struct {
	APIKey  string
	Model   string
	OS      string
	APIURL  string
}

// Load reads configuration from .env files and environment variables
// Priority: env vars > local .env > global .env
func Load() (*Config, error) {
	cfg := &Config{
		Model:  DefaultModel,
		OS:     DefaultOS,
		APIURL: DefaultAPIURL,
	}

	// Load global config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	globalEnvPath := filepath.Join(homeDir, GlobalConfigDir, GlobalEnvFile)
	if err := loadEnvFile(globalEnvPath, cfg); err != nil {
		// Global config is optional, continue
	}

	// Load local config (overrides global)
	if err := loadEnvFile(LocalEnvFile, cfg); err != nil {
		// Local config is optional, continue
	}

	// Environment variables override everything
	if v := os.Getenv("ASK_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("ASK_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("ASK_OS"); v != "" {
		cfg.OS = v
	}
	if v := os.Getenv("ASK_API_URL"); v != "" {
		cfg.APIURL = v
	}

	return cfg, nil
}

// loadEnvFile reads a .env file and applies values to the config
func loadEnvFile(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set (respect previous values)
		switch key {
		case "ASK_API_KEY":
			if cfg.APIKey == "" {
				cfg.APIKey = value
			}
		case "ASK_MODEL":
			if cfg.Model == DefaultModel {
				cfg.Model = value
			}
		case "ASK_OS":
			if cfg.OS == DefaultOS {
				cfg.OS = value
			}
		case "ASK_API_URL":
			if cfg.APIURL == DefaultAPIURL {
				cfg.APIURL = value
			}
		}
	}

	return scanner.Err()
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.APIKey == "" && c.APIURL == DefaultAPIURL {
		return fmt.Errorf("ASK_API_KEY is required for OpenAI API")
	}
	return nil
}
