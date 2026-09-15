package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mxl1c/CodeForge/internal/provider"
	"github.com/mxl1c/CodeForge/internal/seat"
	"github.com/mxl1c/CodeForge/internal/testgen"
)

func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEFORGE_API_KEY", "")
	t.Setenv("CODEFORGE_BASE_URL", "")
	t.Setenv("CODEFORGE_MODEL", "")
	t.Setenv("CODEFORGE_OFFLINE", "")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func execRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestLoginAndInitNotStubs(t *testing.T) {
	isolateHome(t)
	out, err := execRoot(t, "login")
	if err != nil {
		t.Fatalf("login: %v\n%s", err, out)
	}
	if strings.Contains(out, "stub") {
		t.Fatalf("login still stub: %q", out)
	}
	if !strings.Contains(out, "trial") {
		t.Fatalf("want trial seat: %q", out)
	}

	out, err = execRoot(t, "init")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if strings.Contains(out, "stub") {
		t.Fatalf("init still stub: %q", out)
	}
}

func TestSeatLifecycleCommands(t *testing.T) {
	isolateHome(t)
	out, err := execRoot(t, "login", "trial")
	if err != nil {
		t.Fatalf("trial: %v\n%s", err, out)
	}
	if !strings.Contains(out, "trial") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "login", "activate")
	if err != nil {
		t.Fatalf("activate: %v\n%s", err, out)
	}
	if !strings.Contains(out, "active") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "login", "suspend")
	if err != nil {
		t.Fatalf("suspend: %v\n%s", err, out)
	}
	if !strings.Contains(out, "suspended") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "login", "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	if !strings.Contains(out, "suspended") {
		t.Fatalf("%s", out)
	}

	_, err = execRoot(t, "login", "activate")
	if !errors.Is(err, seat.ErrInvalidTransition) {
		t.Fatalf("suspended→active: %v", err)
	}
}

func TestTestGenGoldenAndEmpty(t *testing.T) {
	isolateHome(t)
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	out, err := execRoot(t, "test-gen", "--offline", "--module", mod)
	if err != nil {
		t.Fatalf("golden: %v\n%s", err, out)
	}
	if strings.Contains(out, "stub") {
		t.Fatalf("still stub: %q", out)
	}
	if !strings.Contains(out, "RefundRejectsClosedOrder") || !strings.Contains(out, "GetIsTenantScoped") {
		t.Fatalf("missing cases: %s", out)
	}

	empty := t.TempDir()
	out, err = execRoot(t, "test-gen", "--offline", "--module", empty)
	if !errors.Is(err, testgen.ErrEmptyModule) {
		t.Fatalf("empty: err=%v out=%s", err, out)
	}
	if strings.Contains(out, "RefundRejectsClosedOrder") {
		t.Fatalf("fabricated tests: %s", out)
	}
}

func TestTestGenJavaGolden(t *testing.T) {
	isolateHome(t)
	mod := filepath.Join(repoRoot(t), "samples", "java-saas-admin")
	out, err := execRoot(t, "test-gen", "--offline", "--module", mod)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "refund") {
		t.Fatalf("%s", out)
	}
}

func TestDefectBlameGoldenAndInsufficient(t *testing.T) {
	isolateHome(t)
	stack := filepath.Join(repoRoot(t), "fixtures", "failure-stack-zh.txt")
	out, err := execRoot(t, "defect-blame", "--offline", "--stack", stack)
	if err != nil {
		t.Fatalf("golden: %v\n%s", err, out)
	}
	if !strings.Contains(out, "classification: defect") {
		t.Fatalf("%s", out)
	}
	if !strings.Contains(out, "OrderService") || !strings.Contains(out, "refund") {
		t.Fatalf("location: %s", out)
	}
	if !strings.Contains(out, "repro") {
		t.Fatalf("missing repro: %s", out)
	}

	bad := filepath.Join(repoRoot(t), "fixtures", "insufficient-stack.txt")
	out, err = execRoot(t, "defect-blame", "--offline", "--stack", bad)
	if err == nil {
		t.Fatalf("insufficient should fail: %s", out)
	}
	if strings.Contains(out, "OrderService") {
		t.Fatalf("invented blame: %s", out)
	}
	if !strings.Contains(out, "refusing to invent blame") && !strings.Contains(err.Error(), "insufficient") {
		t.Fatalf("want insufficient, err=%v out=%s", err, out)
	}
}

func TestRegressSuggestGoldenAndDocsOnly(t *testing.T) {
	isolateHome(t)
	diff := filepath.Join(repoRoot(t), "fixtures", "fake-pr.diff")
	out, err := execRoot(t, "regress-suggest", "--offline", "--diff", diff)
	if err != nil {
		t.Fatalf("golden: %v\n%s", err, out)
	}
	if !strings.Contains(out, "P0") || !strings.Contains(out, "P1") {
		t.Fatalf("want P0 and P1: %s", out)
	}
	if strings.Contains(strings.ToLower(out), "full regression") && !strings.Contains(out, "forbidden") {
		t.Fatalf("full regression: %s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "- ") {
			low := strings.ToLower(line)
			if strings.Contains(low, "full regression") || strings.Contains(low, "全量回归") {
				t.Fatalf("suggestion line: %s", line)
			}
		}
	}

	docs := filepath.Join(repoRoot(t), "fixtures", "docs-only.diff")
	out, err = execRoot(t, "regress-suggest", "--offline", "--diff", docs)
	if err != nil {
		t.Fatalf("docs: %v\n%s", err, out)
	}
	if !strings.Contains(out, "docs-only") {
		t.Fatalf("%s", out)
	}
	if strings.Contains(out, "- P0") || strings.Contains(out, "- P1") {
		t.Fatalf("docs escalated: %s", out)
	}

	empty := filepath.Join(t.TempDir(), "empty.diff")
	if err := os.WriteFile(empty, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = execRoot(t, "regress-suggest", "--offline", "--diff", empty)
	if err == nil {
		t.Fatal("empty diff should fail")
	}
}

func TestSuspendedBlocksVerticalCommands(t *testing.T) {
	isolateHome(t)
	if _, err := execRoot(t, "login", "trial"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "login", "activate"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "login", "suspend"); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	_, err := execRoot(t, "test-gen", "--offline", "--module", mod)
	if !errors.Is(err, seat.ErrSuspended) {
		t.Fatalf("want ErrSuspended, got %v", err)
	}
}

func TestProviderPing_MissingAPIKey(t *testing.T) {
	isolateHome(t)
	_, err := execRoot(t, "provider", "ping")
	if !errors.Is(err, provider.ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey, got %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "CODEFORGE_API_KEY") {
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

	out, err := execRoot(t, "provider", "ping")
	if err != nil {
		t.Fatalf("execute: %v\n%s", err, out)
	}
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

func TestOfflineVerticalDoesNotCallComplete(t *testing.T) {
	isolateHome(t)

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	t.Setenv("CODEFORGE_API_KEY", "test-key")
	t.Setenv("CODEFORGE_BASE_URL", srv.URL)

	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	stack := filepath.Join(repoRoot(t), "fixtures", "failure-stack-zh.txt")
	diff := filepath.Join(repoRoot(t), "fixtures", "fake-pr.diff")

	if _, err := execRoot(t, "login"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "test-gen", "--offline", "--module", mod); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "defect-blame", "--offline", "--stack", stack); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "regress-suggest", "--offline", "--diff", diff); err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("--offline must not call Complete, got %d", got)
	}
}

func TestTestGenCallsProviderWhenKeyPresent(t *testing.T) {
	isolateHome(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"cases":[]}`}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("CODEFORGE_API_KEY", "test-key")
	t.Setenv("CODEFORGE_BASE_URL", srv.URL)

	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	out, err := execRoot(t, "test-gen", "--module", mod)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d", calls.Load())
	}
	if !strings.Contains(out, "mode: provider") {
		t.Fatalf("%s", out)
	}
}

func TestMissingFlagsFail(t *testing.T) {
	isolateHome(t)
	if _, err := execRoot(t, "test-gen", "--offline"); err == nil {
		t.Fatal("test-gen without module")
	}
	if _, err := execRoot(t, "defect-blame", "--offline"); err == nil {
		t.Fatal("defect-blame without stack")
	}
	if _, err := execRoot(t, "regress-suggest", "--offline"); err == nil {
		t.Fatal("regress-suggest without diff")
	}
}
