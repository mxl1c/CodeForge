// Package provider defines the switchable model backend used by CodeForge.
// W1 ships an OpenAI-compatible adapter; other vendors implement this interface.
package provider

import (
	"context"
	"fmt"
)

// Message is a single chat turn.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is a vendor-neutral completion request.
type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
}

// ChatResponse is a vendor-neutral completion result.
type ChatResponse struct {
	Content string
	Model   string
}

// Provider is the model backend abstraction. Commands depend on this interface,
// not on a specific vendor SDK.
type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

// ErrNoAPIKey means a live call was requested but no credential is configured.
var ErrNoAPIKey = fmt.Errorf("no API key: set CODEFORGE_API_KEY or run `codeforge login`")

// ErrUnknownProvider is returned when config names a backend that is not registered.
type ErrUnknownProvider struct {
	Name string
}

func (e ErrUnknownProvider) Error() string {
	return fmt.Sprintf("unknown provider %q (supported: openai-compatible)", e.Name)
}
