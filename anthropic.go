package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var _ MessageClient = (*Anthropic)(nil)

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeMessagesRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []ClaudeMessage `json:"messages"`
}

type ClaudeMessagesResponseContent struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type ClaudeMessagesResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ClaudeMessagesResponse struct {
	ID      string                          `json:"id"`
	Role    string                          `json:"role"`
	Model   string                          `json:"model"`
	Content []ClaudeMessagesResponseContent `json:"content"`
	Usage   ClaudeMessagesResponseUsage     `json:"usage"`
}

type Anthropic struct {
	apiKey string
	model  string
}

func NewAnthropic(apiKey, model string) *Anthropic {
	return &Anthropic{
		apiKey,
		model,
	}
}

func (a *Anthropic) CreateMessage(ctx context.Context, system string, prompt string) (string, error) {
	body, err := json.Marshal(map[string]interface{}{
		"model":      a.model,
		"max_tokens": 4096,
		"system":     system,
		"messages": []ClaudeMessage{
			{
				Role:    MessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("error marshaling JSON payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data ClaudeMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return data.Content[0].Text, nil
}
