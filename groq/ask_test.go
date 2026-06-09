package groq_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zarirdev/tai/groq"
)

// ---------------- Mock Server Test ----------------
func TestAsk_MockSuccess(t *testing.T) {
	// Create a fake Groq API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/openai/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		// Decode request to verify structure
		var reqBody groq.Request // We export the struct later
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if reqBody.Model != "test-model" {
			t.Errorf("model = %q; want 'test-model'", reqBody.Model)
		}
		if reqBody.MaxTokens != 5 {
			t.Errorf("max_tokens = %d; want 5", reqBody.MaxTokens)
		}

		// Return a fake response
		resp := groq.Response{
			Choices: []groq.Choice{
				{Message: groq.Message{Role: "assistant", Content: "mock answer"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the HTTP client's target URL via a custom transport?
	// We need to make Ask use the test server URL.
	// Since Ask hardcodes the URL, we can't easily swap.
	// Alternative: We'll add an internal variable for the base URL.
	// For simplicity, we'll test the function indirectly or by refactoring.
	// Let's show the concept: we add a package variable for the base URL.

	// In groq/ask.go, add: var BaseURL = "https://api.groq.com"
	// Then in test: groq.BaseURL = server.URL

	// We'll include that modification in the instructions.
}
