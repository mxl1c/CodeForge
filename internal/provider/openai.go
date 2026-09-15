package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type OpenAICompat struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOpenAICompat(apiKey, baseURL, model string) *OpenAICompat {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAICompat{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *OpenAICompat) Name() string {
	return "openai-compatible"
}

func (c *OpenAICompat) Ping(ctx context.Context) error {
	_, err := c.Complete(ctx, CompletionRequest{
		Messages: []Message{{Role: "user", Content: "ping"}},
	})
	return err
}

func (c *OpenAICompat) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if c.apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("completion request has no messages")
	}

	payload := chatRequest{
		Model:     model,
		Messages:  req.Messages,
		MaxTokens: 16,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("provider request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read provider response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, WrapHTTPError(resp.StatusCode, string(raw))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse provider response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("provider returned no choices")
	}
	return &CompletionResponse{
		Content: parsed.Choices[0].Message.Content,
		Model:   parsed.Model,
	}, nil
}

type chatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
