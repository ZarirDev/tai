package groq_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zarirdev/tai/groq"
)

func TestAsk_MockSuccess(t *testing.T) {
	// Fake Groq API that always returns "mock answer"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic request validation
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/openai/v1/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		// Decode request to verify fields
		var reqBody groq.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if reqBody.Model != "test-model" {
			t.Errorf("model = %q; want 'test-model'", reqBody.Model)
		}
		if reqBody.MaxTokens != 5 {
			t.Errorf("max_tokens = %d; want 5", reqBody.MaxTokens)
		}

		// Return a valid response
		resp := groq.ChatResponse{
			Choices: []groq.Choice{
				{Message: groq.Message{Role: "assistant", Content: "mock answer"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the base URL
	origBase := groq.BaseURL
	groq.BaseURL = server.URL
	defer func() { groq.BaseURL = origBase }()

	answer, err := groq.Ask("hello?", "test-key", "test-model", 5, nil, nil)
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if answer != "mock answer" {
		t.Errorf("answer = %q; want 'mock answer'", answer)
	}
}

func TestAsk_NoAPIKey(t *testing.T) {
	_, err := groq.Ask("test", "", "model", 10, nil, nil)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestAsk_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "invalid key"},
		})
	}))
	defer server.Close()

	origBase := groq.BaseURL
	groq.BaseURL = server.URL
	defer func() { groq.BaseURL = origBase }()

	_, err := groq.Ask("test", "bad-key", "model", 10, nil, nil)
	if err == nil {
		t.Fatal("expected error from API")
	}
}
