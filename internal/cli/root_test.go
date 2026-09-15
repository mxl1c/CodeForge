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
	"github.com/mxl1c/CodeForge/internal/usage"
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
	if !strings.Contains(out, "https://api.deepseek.com/v1") {
		t.Fatalf("login should document DeepSeek-compatible base URL: %q", out)
	}
	if !strings.Contains(out, "codeforge seat") {
		t.Fatalf("login should point at seat command: %q", out)
	}

	out, err = execRoot(t, "init")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if strings.Contains(out, "stub") {
		t.Fatalf("init still stub: %q", out)
	}
	if !strings.Contains(out, "seat trial") {
		t.Fatalf("init should mention seat contract: %q", out)
	}
	if !strings.Contains(out, "usage export") {
		t.Fatalf("init should mention usage export: %q", out)
	}
}

func TestSeatLifecycleCommands(t *testing.T) {
	isolateHome(t)
	out, err := execRoot(t, "seat", "trial")
	if err != nil {
		t.Fatalf("trial: %v\n%s", err, out)
	}
	if !strings.Contains(out, "trial") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "seat", "activate")
	if err != nil {
		t.Fatalf("activate: %v\n%s", err, out)
	}
	if !strings.Contains(out, "active") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "seat", "suspend")
	if err != nil {
		t.Fatalf("suspend: %v\n%s", err, out)
	}
	if !strings.Contains(out, "suspended") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "seat", "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	if !strings.Contains(out, "suspended") {
		t.Fatalf("%s", out)
	}

	_, err = execRoot(t, "seat", "activate")
	if !errors.Is(err, seat.ErrInvalidTransition) {
		t.Fatalf("suspended→active: %v", err)
	}

	_, err = execRoot(t, "login", "trial")
	if err == nil {
		t.Fatal("login must not own seat subcommands")
	}
}

func TestTestGenGoldenAndEmpty(t *testing.T) {
	isolateHome(t)
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	out, err := execRoot(t, "test-gen", "--offline", "--repo", mod)
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
	out, err = execRoot(t, "test-gen", "--offline", "--repo", empty)
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
	out, err := execRoot(t, "test-gen", "--offline", "--repo", mod)
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
	out, err := execRoot(t, "defect-blame", "--offline", "--log", stack)
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
	out, err = execRoot(t, "defect-blame", "--offline", "--log", bad)
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
	if _, err := execRoot(t, "seat", "trial"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "seat", "activate"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "seat", "suspend"); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	_, err := execRoot(t, "test-gen", "--offline", "--repo", mod)
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
	if _, err := execRoot(t, "test-gen", "--offline", "--repo", mod); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "defect-blame", "--offline", "--log", stack); err != nil {
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
	out, err := execRoot(t, "test-gen", "--repo", mod)
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
		t.Fatal("test-gen without --repo must fail")
	}
}

func TestLockedCLIContracts(t *testing.T) {
	isolateHome(t)
	root := NewRoot()

	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"login", "init", "test-gen", "defect-blame", "regress-suggest", "seat", "usage"} {
		if !names[want] {
			t.Fatalf("missing command %s", want)
		}
	}

	tg, _, err := root.Find([]string{"test-gen"})
	if err != nil {
		t.Fatal(err)
	}
	if tg.Flags().Lookup("repo") == nil {
		t.Fatal("test-gen must expose --repo")
	}
	if tg.Flags().Lookup("module") != nil {
		t.Fatal("do not invent --module; locked flag is --repo")
	}

	db, _, err := root.Find([]string{"defect-blame"})
	if err != nil {
		t.Fatal(err)
	}
	logFlag := db.Flags().Lookup("log")
	if logFlag == nil {
		t.Fatal("defect-blame must expose --log")
	}
	if logFlag.DefValue != defaultLogPath {
		t.Fatalf("default --log=%q want %q", logFlag.DefValue, defaultLogPath)
	}
	if db.Flags().Lookup("stack") != nil {
		t.Fatal("do not invent --stack; locked flag is --log")
	}

	rs, _, err := root.Find([]string{"regress-suggest"})
	if err != nil {
		t.Fatal(err)
	}
	diffFlag := rs.Flags().Lookup("diff")
	if diffFlag == nil {
		t.Fatal("regress-suggest must expose --diff")
	}
	if diffFlag.DefValue != defaultDiffPath {
		t.Fatalf("default --diff=%q want %q", diffFlag.DefValue, defaultDiffPath)
	}

	seatCmd, _, err := root.Find([]string{"seat"})
	if err != nil {
		t.Fatal(err)
	}
	subs := map[string]bool{}
	for _, c := range seatCmd.Commands() {
		subs[c.Name()] = true
	}
	for _, want := range []string{"trial", "activate", "suspend", "status", "tier"} {
		if !subs[want] {
			t.Fatalf("seat missing %s", want)
		}
	}

	usageCmd, _, err := root.Find([]string{"usage"})
	if err != nil {
		t.Fatal(err)
	}
	usageSubs := map[string]bool{}
	for _, c := range usageCmd.Commands() {
		usageSubs[c.Name()] = true
	}
	if !usageSubs["export"] {
		t.Fatal("usage missing export")
	}
	exportCmd, _, err := root.Find([]string{"usage", "export"})
	if err != nil {
		t.Fatal(err)
	}
	if exportCmd.Flags().Lookup("out") == nil {
		t.Fatal("usage export must expose --out")
	}

	out, err := execRoot(t, "test-gen", "--offline", "--module", t.TempDir())
	if err == nil {
		t.Fatalf("unknown --module must fail, out=%s", out)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repoRoot(t)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	out, err = execRoot(t, "defect-blame", "--offline")
	if err != nil {
		t.Fatalf("default --log: %v\n%s", err, out)
	}
	if !strings.Contains(out, "classification: defect") {
		t.Fatalf("default log golden: %s", out)
	}

	out, err = execRoot(t, "regress-suggest", "--offline")
	if err != nil {
		t.Fatalf("default --diff: %v\n%s", err, out)
	}
	if !strings.Contains(out, "P0") {
		t.Fatalf("default diff golden: %s", out)
	}
}

func TestSeatTierShowAndSet(t *testing.T) {
	isolateHome(t)
	out, err := execRoot(t, "seat", "trial")
	if err != nil {
		t.Fatalf("trial: %v\n%s", err, out)
	}
	if !strings.Contains(out, "tier=free") || !strings.Contains(out, "Free") {
		t.Fatalf("default free: %s", out)
	}

	out, err = execRoot(t, "seat", "tier", "pro")
	if err != nil {
		t.Fatalf("tier pro: %v\n%s", err, out)
	}
	if !strings.Contains(out, "tier=pro") || !strings.Contains(out, "~¥140") {
		t.Fatalf("%s", out)
	}
	if !strings.Contains(out, "not payment") {
		t.Fatalf("must not invent payment: %s", out)
	}

	out, err = execRoot(t, "seat", "activate")
	if err != nil {
		t.Fatalf("activate: %v\n%s", err, out)
	}
	if !strings.Contains(out, "active") || !strings.Contains(out, "tier=pro") {
		t.Fatalf("lifecycle must keep pro: %s", out)
	}

	out, err = execRoot(t, "seat", "tier", "business")
	if err != nil {
		t.Fatalf("business: %v\n%s", err, out)
	}
	if !strings.Contains(out, "tier=business") || !strings.Contains(out, "~¥700–1400") {
		t.Fatalf("%s", out)
	}
	if !strings.Contains(out, "state=active") {
		t.Fatalf("tier must not change lifecycle: %s", out)
	}

	out, err = execRoot(t, "seat", "tier")
	if err != nil {
		t.Fatalf("show: %v\n%s", err, out)
	}
	if !strings.Contains(out, "business") {
		t.Fatalf("%s", out)
	}

	out, err = execRoot(t, "seat", "status")
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	if !strings.Contains(out, "active") || !strings.Contains(out, "business") {
		t.Fatalf("%s", out)
	}

	_, err = execRoot(t, "seat", "tier", "enterprise")
	if !errors.Is(err, seat.ErrInvalidTier) {
		t.Fatalf("invalid tier: %v", err)
	}
}

func TestUsageExportOfflineAndProvider(t *testing.T) {
	isolateHome(t)
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	out, err := execRoot(t, "test-gen", "--offline", "--repo", mod)
	if err != nil {
		t.Fatalf("test-gen: %v\n%s", err, out)
	}

	csvPath := filepath.Join(t.TempDir(), "arpu.csv")
	out, err = execRoot(t, "usage", "export", "--out", csvPath)
	if err != nil {
		t.Fatalf("export csv: %v\n%s", err, out)
	}
	if !strings.Contains(out, "usage export: ok") || !strings.Contains(out, "not a general IDE") {
		t.Fatalf("wedge/export: %s", out)
	}
	raw, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "timestamp,tenant,seat_id,tier,command,provider,model,prompt_tokens,completion_tokens,status") {
		t.Fatalf("csv header: %s", text)
	}
	if !strings.Contains(text, "test-gen") || !strings.Contains(text, "offline") {
		t.Fatalf("csv: %s", text)
	}
	st := usage.Open(os.Getenv("HOME"))
	rows, err := st.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].Command != "test-gen" || rows[0].Provider != usage.ProviderOffline {
		t.Fatalf("%+v", rows[0])
	}
	if rows[0].PromptTokens != 0 || rows[0].CompletionTokens != 0 || rows[0].Status != usage.StatusOK {
		t.Fatalf("deterministic offline row: %+v", rows[0])
	}
	if rows[0].Tenant == "" || rows[0].SeatID == "" || rows[0].Tier != "free" {
		t.Fatalf("seat fields: %+v", rows[0])
	}

	empty := t.TempDir()
	_, err = execRoot(t, "test-gen", "--offline", "--repo", empty)
	if !errors.Is(err, testgen.ErrEmptyModule) {
		t.Fatalf("empty: %v", err)
	}
	rows, err = st.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1].Status != usage.StatusError || rows[1].Provider != usage.ProviderOffline {
		t.Fatalf("error row: %+v", rows)
	}

	jsonPath := filepath.Join(t.TempDir(), "arpu.json")
	out, err = execRoot(t, "usage", "export", "--out", jsonPath)
	if err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	jraw, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var exported []usage.Entry
	if err := json.Unmarshal(jraw, &exported); err != nil {
		t.Fatal(err)
	}
	if len(exported) != 2 {
		t.Fatalf("json rows=%d %s", len(exported), jraw)
	}
}

func TestUsageExportWiresCompleteTokens(t *testing.T) {
	isolateHome(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "deepseek-chat",
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"cases":[]}`}},
			},
			"usage": map[string]int{"prompt_tokens": 21, "completion_tokens": 5, "total_tokens": 26},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("CODEFORGE_API_KEY", "test-key")
	t.Setenv("CODEFORGE_BASE_URL", srv.URL)
	t.Setenv("CODEFORGE_MODEL", "deepseek-chat")

	if _, err := execRoot(t, "seat", "tier", "pro"); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	out, err := execRoot(t, "test-gen", "--repo", mod)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d", calls.Load())
	}

	st := usage.Open(os.Getenv("HOME"))
	rows, err := st.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%d", len(rows))
	}
	got := rows[0]
	if got.Provider != usage.ProviderCompat || got.Model != "deepseek-chat" {
		t.Fatalf("%+v", got)
	}
	if got.PromptTokens != 21 || got.CompletionTokens != 5 {
		t.Fatalf("tokens %+v", got)
	}
	if got.Tier != "pro" || got.Status != usage.StatusOK {
		t.Fatalf("%+v", got)
	}

	csvPath := filepath.Join(t.TempDir(), "tokens.csv")
	if _, err := execRoot(t, "usage", "export", "--out", csvPath, "--format", "csv"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "21") || !strings.Contains(string(raw), "5") || !strings.Contains(string(raw), "openai-compat") {
		t.Fatalf("export missing complete usage: %s", raw)
	}
}

func TestSuspendedUsageIsBlocked(t *testing.T) {
	isolateHome(t)
	if _, err := execRoot(t, "seat", "trial"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "seat", "activate"); err != nil {
		t.Fatal(err)
	}
	if _, err := execRoot(t, "seat", "suspend"); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(repoRoot(t), "samples", "go-saas-admin")
	_, err := execRoot(t, "test-gen", "--offline", "--repo", mod)
	if !errors.Is(err, seat.ErrSuspended) {
		t.Fatalf("want ErrSuspended, got %v", err)
	}
	rows, err := usage.Open(os.Getenv("HOME")).LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Status != usage.StatusBlocked {
		t.Fatalf("%+v", rows)
	}
}

func TestWedgeRemainsQEOnly(t *testing.T) {
	isolateHome(t)
	root := NewRoot()
	if strings.Contains(strings.ToLower(root.Short), "ide plugin") {
		t.Fatal(root.Short)
	}
	for _, name := range []string{"chat", "plugin", "marketplace", "pay", "checkout"} {
		if _, _, err := root.Find([]string{name}); err == nil {
			t.Fatalf("wedge leak: %s", name)
		}
	}
}
