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
}

type groqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

var groqClient = &http.Client{Timeout: 30 * time.Second}

// callGroq sends a prompt and returns the reply.
func callGroq(prompt string, maxTokens int) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	reqBody := groqRequest{
		Messages:  []Message{{Role: "user", Content: prompt}},
		Model:     "llama-3.3-70b-versatile", // ✅ working free model
		Stream:    false,
		MaxTokens: maxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := groqClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// read body for more info
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
		return "", fmt.Errorf("decode: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no answer")
	}
	return strings.TrimSpace(groqResp.Choices[0].Message.Content), nil
}

// askGroq answers user query (optionally with context).
func askGroq(userQuery, context string) (string, error) {
	var prompt string
	if context == "" {
		prompt = fmt.Sprintf("Answer concisely. If unsure, say so.\nUser: %s", userQuery)
	} else {
		prompt = fmt.Sprintf("Context:\n%s\n\nAnswer based on context: %s", context, userQuery)
	}
	return callGroq(prompt, 1024)
}

// requiresWebSearch decides if we need live data.
func requiresWebSearch(query string) (bool, error) {
	prompt := fmt.Sprintf(`Reply ONLY "YES" or "NO". Need up‑to‑date info for: "%s"`, query)
	reply, err := callGroq(prompt, 5)
	if err != nil {
		return false, err
	}
	return strings.ToUpper(strings.TrimSpace(reply)) == "YES", nil
}
