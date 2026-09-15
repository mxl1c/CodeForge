package usage

import (
	"context"

	"github.com/mxl1c/CodeForge/internal/provider"
)

// Capture records token usage from a provider Complete call when one happens.
type Capture struct {
	inner            provider.Provider
	Called           bool
	PromptTokens     int
	CompletionTokens int
	Model            string
}

func Wrap(p provider.Provider) *Capture {
	return &Capture{inner: p}
}

// Provider returns the wrapped adapter, or nil when running offline/stub.
func (c *Capture) Provider() provider.Provider {
	if c == nil || c.inner == nil {
		return nil
	}
	return c
}

func (c *Capture) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	c.Called = true
	resp, err := c.inner.Complete(ctx, req)
	if resp != nil {
		c.PromptTokens = resp.Usage.PromptTokens
		c.CompletionTokens = resp.Usage.CompletionTokens
		if resp.Model != "" {
			c.Model = resp.Model
		} else {
			c.Model = req.Model
		}
	} else if req.Model != "" && c.Model == "" {
		c.Model = req.Model
	}
	return resp, err
}

func StatusOf(err error) string {
	if err == nil {
		return StatusOK
	}
	return StatusError
}
