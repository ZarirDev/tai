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

// ---------- Groq API types ----------
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

// ---------- common client ----------
var groqHTTPClient = &http.Client{Timeout: 30 * time.Second}

// callGroq sends a prompt to Groq and returns the assistant's reply.
func callGroq(prompt string, maxTokens int) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable not set")
	}

	reqBody := groqRequest{
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Model:     "groq/compound-mini",
		Stream:    false,
		MaxTokens: maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := groqHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API error: %s", resp.Status)
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

// ---------- public functions ----------

// askGroq answers the user query using optional context.
func askGroq(userQuery, context string) (string, error) {
	var prompt string
	if context == "" {
		// No context – answer from internal knowledge
		prompt = fmt.Sprintf(`You are a helpful assistant. Answer the user's question concisely and accurately.
If you don't know the answer, say so.

User question: %s`, userQuery)
	} else {
		prompt = fmt.Sprintf(`You are a helpful assistant. Use the following context (scraped from the web or user-provided URLs) to answer the user's question.

Context:
%s

User question: %s

Answer concisely and accurately. If the context doesn't contain the answer, say so.`, context, userQuery)
	}
	return callGroq(prompt, 1024)
}

// requiresWebSearch asks Groq whether the user's query needs fresh web search.
func requiresWebSearch(query string) (bool, error) {
	prompt := fmt.Sprintf(`You are a decision engine. Answer ONLY with "YES" or "NO" (no extra words, no punctuation).
Does the user need up‑to‑date information from the internet for this question?
Examples:
- "current weather in Paris" → YES
- "what is a dog" → NO
- "latest news about AI" → YES
- "who wrote Hamlet" → NO
- "stock price of AAPL today" → YES
- "hello" → NO

Question: "%s"
Answer:`, query)

	reply, err := callGroq(prompt, 5)
	if err != nil {
		return false, err
	}
	reply = strings.ToUpper(strings.TrimSpace(reply))
	return reply == "YES", nil
}
