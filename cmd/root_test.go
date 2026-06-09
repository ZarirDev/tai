package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/zarirdev/tai/cmd"
	"github.com/zarirdev/tai/groq"
	"github.com/zarirdev/tai/pkg/keystore"
)

func TestRootCommand_WithMockAPI(t *testing.T) {
	// Prepare a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := groq.Response{
			Choices: []groq.Choice{
				{Message: groq.Message{Role: "assistant", Content: "Paris"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override base URL
	origBase := groq.BaseURL
	groq.BaseURL = server.URL
	defer func() { groq.BaseURL = origBase }()

	// Provide a dummy encrypted API key in config, or set env to bypass decryption?
	// The command expects api_key_encrypted in config, which requires decryption.
	// We'll create a temporary config with an encrypted key that decrypts to "test-key".
	// Use keystore.Encrypt to generate it.

	home := t.TempDir()
	t.Setenv("HOME", home)

	// Write user config with encrypted dummy key
	configDir := home + "/.config/tai"
	os.MkdirAll(configDir, 0700)
	encKey, _ := keystore.Encrypt("test-key")
	configContent := `
api_key_encrypted: "` + encKey + `"
model: groq/compound-mini
max_tokens: 10
`
	os.WriteFile(configDir+"/config.yaml", []byte(configContent), 0644)

	// Run the command
	rootCmd := cmd.GetRootCommand() // We need to export the root command
	rootCmd.SetArgs([]string{"What is the capital of France?"})

	// Capture stdout
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetErr(b)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := b.String()
	if output != "Paris\n" {
		t.Errorf("expected 'Paris\\n', got %q", output)
	}
}
