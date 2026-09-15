package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// ErrNoAPIKey is returned when the OpenAI-compatible adapter has no credential.
var ErrNoAPIKey = errors.New("missing API key: set CODEFORGE_API_KEY or api_key in ~/.codeforge/config.yaml")

// OpenAICompat calls an OpenAI-compatible Chat Completions endpoint.
type OpenAICompat struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewOpenAICompat(apiKey, baseURL string) (*OpenAICompat, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrNoAPIKey
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAICompat{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (p *OpenAICompat) BaseURL() string {
	return p.baseURL
}

func (p *OpenAICompat) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return nil, ErrNoAPIKey
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("provider request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider HTTP %d: %s", resp.StatusCode, truncate(raw, 512))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, errors.New("provider returned no choices")
	}
	return &CompletionResponse{
		Content: parsed.Choices[0].Message.Content,
		Model:   parsed.Model,
		Usage:   parsed.Usage,
	}, nil
}

type chatResponse struct {
	Model   string `json:"model"`
	Usage   Usage  `json:"usage"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func truncate(b []byte, n int) string {
	s := strings.TrimSpace(string(b))
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
