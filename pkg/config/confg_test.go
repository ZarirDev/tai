package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/zarirdev/tai/pkg/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Create a temporary HOME so no user config exists
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	v, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if v.GetString("model") != "groq/compound-mini" {
		t.Errorf("default model = %q; want 'groq/compound-mini'", v.GetString("model"))
	}
	if v.GetInt("max_tokens") != 2048 {
		t.Errorf("default max_tokens = %d; want 2048", v.GetInt("max_tokens"))
	}
	if v.GetBool("debug") != false {
		t.Errorf("default debug = %t; want false", v.GetBool("debug"))
	}
}

func TestLoadUserConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Write a user config
	configDir := filepath.Join(tmpHome, ".config", "tai")
	os.MkdirAll(configDir, 0700)
	content := `
model: "custom-model"
max_tokens: 100
debug: true
include_domains:
  - example.com
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644)

	v, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if v.GetString("model") != "custom-model" {
		t.Errorf("model = %q; want 'custom-model'", v.GetString("model"))
	}
	if v.GetInt("max_tokens") != 100 {
		t.Errorf("max_tokens = %d; want 100", v.GetInt("max_tokens"))
	}
	if v.GetBool("debug") != true {
		t.Errorf("debug = %t; want true", v.GetBool("debug"))
	}
	inc := v.GetStringSlice("include_domains")
	if len(inc) != 1 || inc[0] != "example.com" {
		t.Errorf("include_domains = %v; want [example.com]", inc)
	}
}

func TestSaveUserConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	v := viper.New()
	v.Set("model", "test-model")
	v.Set("max_tokens", 300)
	v.Set("api_key_encrypted", "some-encrypted-key")

	err := config.SaveUserConfig(v)
	if err != nil {
		t.Fatalf("SaveUserConfig failed: %v", err)
	}

	// Check file exists and contains data
	configPath := filepath.Join(tmpHome, ".config", "tai", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Config file is empty")
	}
}
