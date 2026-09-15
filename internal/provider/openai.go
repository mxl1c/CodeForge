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

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "gpt-4o-mini"
)

// OpenAI is an OpenAI-compatible Chat Completions adapter.
// It talks to any host that implements POST {baseURL}/chat/completions
// (OpenAI, Azure OpenAI-compatible gateways, local vLLM/Ollama proxies, etc.).
type OpenAI struct {
	name    string
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// OpenAIOptions configures the adapter.
type OpenAIOptions struct {
	Name    string
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewOpenAI builds an OpenAI-compatible provider.
func NewOpenAI(opts OpenAIOptions) *OpenAI {
	name := opts.Name
	if name == "" {
		name = "openai-compatible"
	}
	base := strings.TrimRight(opts.BaseURL, "/")
	if base == "" {
		base = defaultOpenAIBaseURL
	}
	model := opts.Model
	if model == "" {
		model = defaultOpenAIModel
	}
	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &OpenAI{
		name:    name,
		baseURL: base,
		apiKey:  opts.APIKey,
		model:   model,
		client:  client,
	}
}

func (o *OpenAI) Name() string { return o.name }

type chatCompletionsRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type chatCompletionsResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Chat performs a real HTTP call to the OpenAI-compatible Chat Completions API.
func (o *OpenAI) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if strings.TrimSpace(o.apiKey) == "" {
		return ChatResponse{}, ErrNoAPIKey
	}

	model := req.Model
	if model == "" {
		model = o.model
	}

	payload, err := json.Marshal(chatCompletionsRequest{
		Model:       model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	})
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("build chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("chat completions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read chat response: %w", err)
	}

	var parsed chatCompletionsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ChatResponse{}, fmt.Errorf("decode chat response (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return ChatResponse{}, fmt.Errorf("provider error: %s", parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ChatResponse{}, fmt.Errorf("chat completions HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("chat completions: empty choices")
	}

	return ChatResponse{
		Content: parsed.Choices[0].Message.Content,
		Model:   parsed.Model,
	}, nil
}
