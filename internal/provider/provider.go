package provider

import "context"

// Message is a single chat message in an OpenAI-compatible request.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionResponse struct {
	Content string
	Model   string
	Usage   Usage
}

// Provider is the model backend used by test-gen / defect-blame / regress-suggest.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}
