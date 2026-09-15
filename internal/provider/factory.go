package provider

import "strings"

// Config is the subset of runtime settings a provider needs.
type Config struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

// New constructs a Provider from config. The default (and currently only
// shipped) implementation is the OpenAI-compatible adapter.
func New(cfg Config) (Provider, error) {
	name := strings.TrimSpace(strings.ToLower(cfg.Provider))
	switch name {
	case "", "openai", "openai-compatible":
		return NewOpenAI(OpenAIOptions{
			Name:    "openai-compatible",
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
			Model:   cfg.Model,
		}), nil
	default:
		return nil, ErrUnknownProvider{Name: cfg.Provider}
	}
}
