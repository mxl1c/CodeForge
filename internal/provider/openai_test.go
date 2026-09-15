package provider_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func TestNewFromResolvedMissingKey(t *testing.T) {
	_, err := provider.NewFromResolved(config.Resolved{})
	if !errors.Is(err, provider.ErrMissingAPIKey) {
		t.Fatalf("err=%v, want ErrMissingAPIKey", err)
	}
}

func TestOpenAICompatPingSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("auth=%s", got)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"gpt-4o-mini","choices":[{"message":{"role":"assistant","content":"pong"}}]}`))
	}))
	defer srv.Close()

	p := provider.NewOpenAICompat("test-key", srv.URL, "gpt-4o-mini")
	if err := p.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestOpenAICompatPingHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	p := provider.NewOpenAICompat("bad-key", srv.URL, "gpt-4o-mini")
	err := p.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOpenAICompatMissingKeyOnComplete(t *testing.T) {
	p := provider.NewOpenAICompat("", "https://example.invalid/v1", "gpt-4o-mini")
	_, err := p.Complete(context.Background(), provider.CompletionRequest{
		Messages: []provider.Message{{Role: "user", Content: "hi"}},
	})
	if !errors.Is(err, provider.ErrMissingAPIKey) {
		t.Fatalf("err=%v", err)
	}
}
