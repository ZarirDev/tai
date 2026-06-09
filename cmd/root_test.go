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
	// Mock Groq API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := groq.ChatResponse{
			Choices: []groq.Choice{
				{Message: groq.Message{Role: "assistant", Content: "Paris"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the API URL used by the real client
	origBase := groq.BaseURL
	groq.BaseURL = server.URL
	defer func() { groq.BaseURL = origBase }()

	// Set up a temporary home directory with a dummy config
	home := t.TempDir()
	t.Setenv("HOME", home)

	configDir := home + "/.config/tai"
	os.MkdirAll(configDir, 0700)

	// Encrypt a test key
	encKey, err := keystore.Encrypt("test-key")
	if err != nil {
		t.Fatalf("failed to encrypt test key: %v", err)
	}

	configContent := `
api_key_encrypted: "` + encKey + `"
model: groq/compound-mini
max_tokens: 10
`
	err = os.WriteFile(configDir+"/config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Run the command
	rootCmd := cmd.GetRootCommand()
	rootCmd.SetArgs([]string{"What is the capital of France?"})

	output := bytes.NewBufferString("")
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	got := output.String()
	if got != "Paris\n" {
		t.Errorf("expected 'Paris\\n', got %q", got)
	}
}
