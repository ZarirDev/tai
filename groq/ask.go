package groq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BaseURL is the Groq API endpoint (overrideable for testing).
var BaseURL = "https://api.groq.com"

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the body sent to the Groq chat completions endpoint.
type ChatRequest struct {
	Messages       []Message       `json:"messages"`
	Model          string          `json:"model"`
	Stream         bool            `json:"stream"`
	MaxTokens      int             `json:"max_tokens"`
	CompoundCustom *CompoundCustom `json:"compound_custom,omitempty"`
}

// CompoundCustom holds optional search settings.
type CompoundCustom struct {
	SearchSettings SearchSettings `json:"search_settings,omitempty"`
}

// SearchSettings allows domain filtering.
type SearchSettings struct {
	IncludeDomains []string `json:"include_domains,omitempty"`
	ExcludeDomains []string `json:"exclude_domains,omitempty"`
}

// ChatResponse is the top‑level response from Groq.
type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice contains a single message choice.
type Choice struct {
	Message Message `json:"message"`
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

// Ask sends a query to Groq and returns the answer.
func Ask(query, apiKey, model string, maxTokens int, includeDomains, excludeDomains []string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("no API key provided")
	}

	reqBody := ChatRequest{
		Messages: []Message{
			{Role: "user", Content: query},
		},
		Model:     model,
		Stream:    false,
		MaxTokens: maxTokens,
	}

	if len(includeDomains) > 0 || len(excludeDomains) > 0 {
		reqBody.CompoundCustom = &CompoundCustom{
			SearchSettings: SearchSettings{
				IncludeDomains: includeDomains,
				ExcludeDomains: excludeDomains,
			},
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", BaseURL+"/openai/v1/chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("groq API error (%d): %s", resp.StatusCode, errBody.Error.Message)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no answer from Groq")
	}
	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
