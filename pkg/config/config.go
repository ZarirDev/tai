package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Model           string   `mapstructure:"model"`
	MaxTokens       int      `mapstructure:"max_tokens"`
	Debug           bool     `mapstructure:"debug"`
	IncludeDomains  []string `mapstructure:"include_domains"`
	ExcludeDomains  []string `mapstructure:"exclude_domains"`
	APIKeyEncrypted string   `mapstructure:"api_key_encrypted"` // encrypted base64
}

// LoadConfig reads config from standard locations:
// 1. ~/.config/tai/config.yaml (user config, highest priority)
// 2. /etc/tai/config.yaml (system config, fallback)
// Flags can override these values later.
func LoadConfig() (*viper.Viper, error) {
	v := viper.New()

	// Default values
	v.SetDefault("model", "groq/compound-mini")
	v.SetDefault("max_tokens", 2048)
	v.SetDefault("debug", false)
	v.SetDefault("include_domains", []string{})
	v.SetDefault("exclude_domains", []string{})
	v.SetDefault("api_key_encrypted", "")

	// System config (lowest priority)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/tai/")

	// User config (highest priority)
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "tai"))
	}

	// Also look in current directory as a convenience
	v.AddConfigPath(".")

	// Read config file(s)
	if err := v.ReadInConfig(); err != nil {
		// It's okay if no config file exists; we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	return v, nil
}

// SaveUserConfig writes the current viper state to the user config file
// (creating ~/.config/tai/ if needed).
func SaveUserConfig(v *viper.Viper) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}
	configDir := filepath.Join(home, ".config", "tai")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	return v.WriteConfigAs(configPath)
}
