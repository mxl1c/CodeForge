package testgen

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/mxl1c/CodeForge/internal/provider"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestGoldenGoSample(t *testing.T) {
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	res, err := Run(context.Background(), Input{ModuleDir: mod, Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "offline" {
		t.Fatalf("mode=%s", res.Mode)
	}
	if !hasCase(res, "Refund", "RefundRejectsClosedOrder") {
		t.Fatalf("missing refund case: %+v", res.Cases)
	}
	if !hasCase(res, "Get", "GetIsTenantScoped") {
		t.Fatalf("missing tenant get case: %+v", res.Cases)
	}
}

func TestGoldenJavaSample(t *testing.T) {
	mod := filepath.Join(repoRoot(t), "samples", "java-saas-admin")
	res, err := Run(context.Background(), Input{ModuleDir: mod, Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasCase(res, "refund", "RefundRejectsClosedOrder") {
		t.Fatalf("missing refund case: %+v", res.Cases)
	}
	if !hasCase(res, "get", "GetIsTenantScoped") {
		t.Fatalf("missing tenant get case: %+v", res.Cases)
	}
}

func TestEmptyModuleFailsNoFabrication(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# empty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Input{ModuleDir: dir, Offline: true})
	if !errors.Is(err, ErrEmptyModule) {
		t.Fatalf("want ErrEmptyModule, got res=%v err=%v", res, err)
	}
}

func TestEmptyDirFails(t *testing.T) {
	_, err := Run(context.Background(), Input{ModuleDir: t.TempDir(), Offline: true})
	if !errors.Is(err, ErrEmptyModule) {
		t.Fatalf("got %v", err)
	}
}

func TestProviderEnrichesButDropsHallucinations(t *testing.T) {
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		payload := map[string]any{
			"cases": []map[string]string{
				{"name": "RefundIdempotent", "file": "internal/order/service.go", "fn": "Refund", "priority": "P1", "intent": "repeat refund is rejected"},
				{"name": "Invented", "file": "internal/secret/backdoor.go", "fn": "Pwn", "priority": "P0", "intent": "hallucination"},
			},
		}
		body, _ := json.Marshal(payload)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "fixture",
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": string(body)}},
			},
			"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		})
	}))
	t.Cleanup(srv.Close)

	p, err := provider.NewOpenAICompat("k", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Input{ModuleDir: mod, Offline: false, Provider: p, Model: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "provider" {
		t.Fatalf("mode=%s", res.Mode)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d", calls.Load())
	}
	if !hasNamed(res, "RefundIdempotent") {
		t.Fatalf("missing provider case: %+v", res.Cases)
	}
	if hasNamed(res, "Invented") {
		t.Fatalf("hallucinated case must be dropped: %+v", res.Cases)
	}
}

func TestEmptyModuleDoesNotCallProvider(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"cases\":[{\"name\":\"X\",\"file\":\"a.go\",\"fn\":\"Y\"}]}"}}]}`))
	}))
	t.Cleanup(srv.Close)
	p, err := provider.NewOpenAICompat("k", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Run(context.Background(), Input{ModuleDir: t.TempDir(), Offline: false, Provider: p, Model: "m"})
	if !errors.Is(err, ErrEmptyModule) {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("must not call provider on empty module, calls=%d", calls.Load())
	}
}

func hasCase(res *Result, fn, name string) bool {
	for _, c := range res.Cases {
		if c.Function == fn && c.Name == name {
			return true
		}
	}
	return false
}

func hasNamed(res *Result, name string) bool {
	for _, c := range res.Cases {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestNoModule(t *testing.T) {
	_, err := Run(context.Background(), Input{})
	if !errors.Is(err, ErrNoModule) {
		t.Fatalf("got %v", err)
	}
}

func TestTestsOnlyModuleFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x_test.go"), []byte("package x\nfunc TestA(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Run(context.Background(), Input{ModuleDir: dir, Offline: true})
	if !errors.Is(err, ErrEmptyModule) {
		t.Fatalf("got %v", err)
	}
}
