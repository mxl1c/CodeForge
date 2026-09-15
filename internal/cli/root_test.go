package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mxl1c/CodeForge/internal/provider"
)

func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEFORGE_API_KEY", "")
	t.Setenv("CODEFORGE_BASE_URL", "")
}

func TestStubCommandsExitOK(t *testing.T) {
	isolateHome(t)

	commands := []string{"login", "init", "test-gen", "defect-blame", "regress-suggest"}
	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			root := NewRoot()
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{name})
			if err := root.Execute(); err != nil {
				t.Fatalf("execute: %v\n%s", err, buf.String())
			}
			out := buf.String()
			if !strings.Contains(out, "stub (W1 skeleton)") {
				t.Fatalf("missing stub output: %q", out)
			}
		})
	}
}

func TestTestGenReportsMissingAPIKey(t *testing.T) {
	isolateHome(t)
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"test-gen"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "missing API key") {
		t.Fatalf("output=%q", buf.String())
	}
}

func TestProviderPing_MissingAPIKey(t *testing.T) {
	isolateHome(t)
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"provider", "ping"})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error, got output %q", buf.String())
	}
	if !errors.Is(err, provider.ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey, got %v", err)
	}
	if !strings.Contains(err.Error(), "CODEFORGE_API_KEY") {
		t.Fatalf("error should mention CODEFORGE_API_KEY: %v", err)
	}
}

func TestProviderPing_CompleteOnce(t *testing.T) {
	isolateHome(t)

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodPost {
			t.Errorf("method=%s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization=%s", got)
		}
		var req provider.CompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
		}
		if req.Model == "" {
			t.Error("missing model")
		}
		if len(req.Messages) == 0 || req.Messages[0].Content != "ping" {
			t.Errorf("messages=%v", req.Messages)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-4o-mini",
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "pong"}},
			},
			"usage": map[string]int{"prompt_tokens": 3, "completion_tokens": 1, "total_tokens": 4},
		})
	}))
	t.Cleanup(srv.Close)

	t.Setenv("CODEFORGE_API_KEY", "test-key")
	t.Setenv("CODEFORGE_BASE_URL", srv.URL)

	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"provider", "ping"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "provider ping: ok") {
		t.Fatalf("missing success: %q", out)
	}
	if !strings.Contains(out, "model: gpt-4o-mini") {
		t.Fatalf("missing model: %q", out)
	}
	if !strings.Contains(out, "tokens:") || !strings.Contains(out, "total=4") {
		t.Fatalf("missing token usage: %q", out)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("Complete calls=%d, want 1", got)
	}
}

func TestStubsDoNotCallComplete(t *testing.T) {
	isolateHome(t)

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	t.Setenv("CODEFORGE_API_KEY", "test-key")
	t.Setenv("CODEFORGE_BASE_URL", srv.URL)

	for _, name := range []string{"login", "init", "test-gen", "defect-blame", "regress-suggest"} {
		root := NewRoot()
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetErr(buf)
		root.SetArgs([]string{name})
		if err := root.Execute(); err != nil {
			t.Fatalf("%s: %v\n%s", name, err, buf.String())
		}
		if !strings.Contains(buf.String(), "stub (W1 skeleton)") {
			t.Fatalf("%s missing stub output: %q", name, buf.String())
		}
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("stubs must not call Complete, got %d HTTP requests", got)
	}
}
