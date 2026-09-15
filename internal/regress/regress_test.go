package regress

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestGoldenFakePR(t *testing.T) {
	path := filepath.Join(repoRoot(t), "fixtures", "fake-pr.diff")
	res, err := Run(context.Background(), Input{DiffPath: path, Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.DocsOnly {
		t.Fatal("fake PR is not docs-only")
	}
	var p0, p1 bool
	blob := res.Policy
	for _, s := range res.Suggestions {
		blob += s.Title + s.Reason
		if s.Priority == "P0" {
			p0 = true
		}
		if s.Priority == "P1" {
			p1 = true
		}
	}
	if !p0 || !p1 {
		t.Fatalf("want P0 and P1, got %+v", res.Suggestions)
	}
	low := strings.ToLower(blob)
	if strings.Contains(low, "full regression") && !strings.Contains(low, "forbidden") {
		t.Fatalf("must not recommend full regression: %s", blob)
	}
	for _, s := range res.Suggestions {
		low := strings.ToLower(s.Title + " " + s.Reason)
		if strings.Contains(low, "full regression") || strings.Contains(low, "全量回归") {
			t.Fatalf("suggestion recommends full regression: %+v", s)
		}
	}
}

func TestDocsOnlyDoesNotEscalate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "docs.diff")
	diff := "diff --git a/README.md b/README.md\n--- a/README.md\n+++ b/README.md\n@@ -1,2 +1,3 @@\n # CodeForge\n+\n+docs only\n"
	if err := os.WriteFile(p, []byte(diff), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Input{DiffPath: p, Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.DocsOnly {
		t.Fatal("expected docs-only")
	}
	for _, s := range res.Suggestions {
		if s.Priority == "P0" || s.Priority == "P1" {
			t.Fatalf("docs-only escalated: %+v", s)
		}
	}
}

func TestEmptyDiffFails(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.diff")
	if err := os.WriteFile(p, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Run(context.Background(), Input{DiffPath: p, Offline: true})
	if !errors.Is(err, ErrEmptyDiff) {
		t.Fatalf("got %v", err)
	}
}

func TestProviderCannotEscalateDocsOrFullRegression(t *testing.T) {
	p := filepath.Join(t.TempDir(), "docs.diff")
	diff := "diff --git a/docs/guide.md b/docs/guide.md\n+++ b/docs/guide.md\n+# hi\n"
	if err := os.WriteFile(p, []byte(diff), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := `{"suggestions":[{"priority":"P0","title":"full regression suite","file":"docs/guide.md","reason":"run 全量回归"}]}`
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": payload}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	pr, err := provider.NewOpenAICompat("k", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Input{DiffPath: p, Offline: false, Provider: pr, Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range res.Suggestions {
		if s.Priority == "P0" || s.Priority == "P1" {
			t.Fatalf("escalated: %+v", s)
		}
		tlow := strings.ToLower(s.Title + s.Reason)
		if strings.Contains(tlow, "full regression") || strings.Contains(tlow, "全量回归") {
			t.Fatalf("full regression leaked: %+v", s)
		}
	}
}

func TestNoDiffPath(t *testing.T) {
	_, err := Run(context.Background(), Input{})
	if !errors.Is(err, ErrNoDiff) {
		t.Fatalf("got %v", err)
	}
}
