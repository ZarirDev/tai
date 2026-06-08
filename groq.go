package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	Stream    bool      `json:"stream"`
	MaxTokens int       `json:"max_tokens"`
	// Compound models accept optional search settings
	CompoundCustom *compoundCustom `json:"compound_custom,omitempty"`
}

type compoundCustom struct {
	SearchSettings searchSettings `json:"search_settings,omitempty"`
}

type searchSettings struct {
	IncludeDomains []string `json:"include_domains,omitempty"`
	ExcludeDomains []string `json:"exclude_domains,omitempty"`
}

type groqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

// Ask sends a user query to Groq's compound model which has built‑in web search.
func Ask(query string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	reqBody := groqRequest{
		Messages: []Message{
			{Role: "user", Content: query},
		},
		Model:     "groq/compound-mini", // built‑in web search enabled by default
		Stream:    false,
		MaxTokens: 2048,
		// Optionally restrict search domains (uncomment if needed):
		// CompoundCustom: &compoundCustom{
		// 	SearchSettings: searchSettings{
		// 		ExcludeDomains: []string{"pinterest.com", "facebook.com"},
		// 	},
		// },
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(data))
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

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no answer from Groq")
	}
	return strings.TrimSpace(groqResp.Choices[0].Message.Content), nil
}
