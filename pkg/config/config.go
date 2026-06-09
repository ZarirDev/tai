package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// LoadConfig reads configuration from standard locations:
//  1. ~/.config/tai/config.yaml (user config, highest priority)
//  2. /etc/tai/config.yaml (system config, fallback)
//  3. ./config.yaml (current directory, convenience)
func LoadConfig() (*viper.Viper, error) {
	v := viper.New()

	// Default values
	v.SetDefault("model", "groq/compound-mini")
	v.SetDefault("max_tokens", 2048)
	v.SetDefault("debug", false)
	v.SetDefault("include_domains", []string{})
	v.SetDefault("exclude_domains", []string{})
	v.SetDefault("api_key_encrypted", "")

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// *** ORDER MATTERS ***
	// Paths added first are searched first → highest priority.

	// 1. User config (top priority)
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "tai"))
	}

	// 2. System config (fallback)
	v.AddConfigPath("/etc/tai/")

	// 3. Current directory (convenience)
	v.AddConfigPath(".")

	// Read the first config file found
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	return v, nil
}

// SaveUserConfig writes the current Viper state to the user config file.
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
