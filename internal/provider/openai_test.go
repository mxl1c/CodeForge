package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewDefaultIsOpenAICompatible(t *testing.T) {
	p, err := New(Config{APIKey: "sk-test", BaseURL: "http://example.invalid/v1"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openai-compatible" {
		t.Fatalf("name = %q", p.Name())
	}
}

func TestNewUnknownProvider(t *testing.T) {
	_, err := New(Config{Provider: "not-a-real-vendor"})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(ErrUnknownProvider); !ok {
		t.Fatalf("got %T %v", err, err)
	}
}

func TestOpenAIChatLivePath(t *testing.T) {
	var seenAuth, seenPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		seenPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		var req chatCompletionsRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("request json: %v", err)
		}
		if req.Model != "gpt-test" {
			t.Errorf("model = %q", req.Model)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(chatCompletionsResponse{
			Model: "gpt-test",
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{Message: Message{Role: "assistant", Content: "ok-from-adapter"}},
			},
		})
	}))
	defer srv.Close()

	p := NewOpenAI(OpenAIOptions{
		BaseURL: srv.URL + "/v1",
		APIKey:  "sk-live",
		Model:   "gpt-test",
		Client:  srv.Client(),
	})
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "ok-from-adapter" {
		t.Fatalf("content = %q", resp.Content)
	}
	if seenAuth != "Bearer sk-live" {
		t.Fatalf("auth = %q", seenAuth)
	}
	if seenPath != "/v1/chat/completions" {
		t.Fatalf("path = %q", seenPath)
	}
}

func TestOpenAIChatSkipsWithoutKey(t *testing.T) {
	p := NewOpenAI(OpenAIOptions{BaseURL: "http://127.0.0.1:1", APIKey: ""})
	_, err := p.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "x"}},
	})
	if err != ErrNoAPIKey {
		t.Fatalf("err = %v", err)
	}
}

func TestOpenAIChatHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"bad key","type":"invalid_request_error"}}`)
	}))
	defer srv.Close()

	p := NewOpenAI(OpenAIOptions{BaseURL: srv.URL, APIKey: "sk-bad", Client: srv.Client()})
	_, err := p.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "x"}},
	})
	if err == nil || !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("err = %v", err)
	}
}
