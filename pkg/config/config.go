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

	// System config (lowest priority)
	v.AddConfigPath("/etc/tai/")

	// User config (highest priority)
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "tai"))
	}

	// Also current directory (convenience)
	v.AddConfigPath(".")

	// Read the first config file found
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	return v, nil
}

// SaveUserConfig writes the current viper state to the user config file,
// creating ~/.config/tai/ if necessary.
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
