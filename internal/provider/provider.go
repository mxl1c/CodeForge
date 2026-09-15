package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/mxl1c/CodeForge/internal/config"
)

// ErrMissingAPIKey is returned when no API key is available from env or user config.
var ErrMissingAPIKey = errors.New("missing API key: set CODEFORGE_API_KEY, or run `codeforge login`, or add api_key to ~/.codeforge/config.yaml")

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model    string
	Messages []Message
}

type CompletionResponse struct {
	Content string
	Model   string
}

// Provider is the LLM backend used by CodeForge QE workflows.
type Provider interface {
	Name() string
	Ping(ctx context.Context) error
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}

func NewFromResolved(cfg config.Resolved) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	return NewOpenAICompat(cfg.APIKey, cfg.BaseURL, cfg.Model), nil
}

func MissingKeyError() error {
	return ErrMissingAPIKey
}

func WrapHTTPError(status int, body string) error {
	if len(body) > 240 {
		body = body[:240] + "…"
	}
	return fmt.Errorf("provider HTTP %d: %s", status, body)
}
