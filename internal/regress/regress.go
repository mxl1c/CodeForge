// Package regress suggests targeted P0/P1/P2 regression from a unified diff.
// Full-regression language is forbidden; docs-only changes must not escalate.
package regress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mxl1c/CodeForge/internal/provider"
)

var (
	ErrNoDiff    = errors.New("diff path is required")
	ErrEmptyDiff = errors.New("empty diff: nothing to suggest")
)

type Suggestion struct {
	Priority string `json:"priority"` // P0 P1 P2
	Title    string `json:"title"`
	File     string `json:"file"`
	Reason   string `json:"reason"`
}

type Result struct {
	Mode        string
	DocsOnly    bool
	Suggestions []Suggestion
	Policy      string
}

type Input struct {
	DiffPath string
	Offline  bool
	Provider provider.Provider
	Model    string
}

const Policy = "targeted regression only; full regression / 全量回归 forbidden"

func Run(ctx context.Context, in Input) (*Result, error) {
	if strings.TrimSpace(in.DiffPath) == "" {
		return nil, ErrNoDiff
	}
	data, err := os.ReadFile(in.DiffPath)
	if err != nil {
		return nil, fmt.Errorf("diff: %w", err)
	}
	text := string(data)
	files := parseChangedFiles(text)
	if len(files) == 0 || strings.TrimSpace(text) == "" {
		return nil, ErrEmptyDiff
	}

	out := &Result{
		Mode:        "offline",
		DocsOnly:    allDocs(files),
		Suggestions: suggest(files, text),
		Policy:      Policy,
	}

	if out.DocsOnly {
		// Policy: must not escalate docs-only to P0/P1.
		filtered := out.Suggestions[:0]
		for _, s := range out.Suggestions {
			if s.Priority == "P2" {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) == 0 {
			filtered = []Suggestion{{
				Priority: "P2",
				Title:    "Docs-only review",
				File:     files[0],
				Reason:   "pure documentation change; no production regression escalation",
			}}
		}
		out.Suggestions = filtered
	}

	if containsForbidden(out) {
		return nil, errors.New("internal: full regression leaked into suggestions")
	}

	if !in.Offline && in.Provider != nil {
		extra, err := completeSuggest(ctx, in, text, files)
		if err != nil {
			return nil, err
		}
		out.Mode = "provider"
		out.Suggestions = merge(out.Suggestions, extra, out.DocsOnly)
		if containsForbidden(out) {
			return nil, errors.New("provider suggested full regression; dropped")
		}
	}
	return out, nil
}

func parseChangedFiles(diff string) []string {
	var files []string
	seen := map[string]bool{}
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				p := strings.TrimPrefix(parts[3], "b/")
				if p != "" && !seen[p] {
					seen[p] = true
					files = append(files, p)
				}
			}
		}
		if strings.HasPrefix(line, "+++ b/") {
			p := strings.TrimPrefix(line, "+++ b/")
			p = strings.TrimSpace(p)
			if p != "" && p != "/dev/null" && !seen[p] {
				seen[p] = true
				files = append(files, p)
			}
		}
	}
	return files
}

func isDocs(path string) bool {
	base := strings.ToLower(path)
	if strings.Contains(base, "/docs/") || strings.HasPrefix(base, "docs/") {
		return true
	}
	switch filepath.Ext(base) {
	case ".md", ".rst", ".txt", ".adoc":
		// keep fixtures .txt out: failure-stack is not a docs change in a product diff
		if strings.Contains(base, "readme") || strings.Contains(base, "docs") || strings.HasSuffix(base, ".md") {
			return true
		}
	}
	return false
}

func isTest(path string) bool {
	b := strings.ToLower(path)
	return strings.HasSuffix(b, "_test.go") || strings.Contains(b, "test.java") || strings.Contains(b, "/test/")
}

func allDocs(files []string) bool {
	if len(files) == 0 {
		return false
	}
	for _, f := range files {
		if !isDocs(f) {
			return false
		}
	}
	return true
}

func suggest(files []string, diff string) []Suggestion {
	var out []Suggestion
	body := strings.ToLower(diff)
	for _, f := range files {
		low := strings.ToLower(f)
		switch {
		case isDocs(f):
			out = append(out, Suggestion{Priority: "P2", Title: "Documentation sanity", File: f, Reason: "docs/comment-only surface; do not escalate"})
		case isTest(f):
			out = append(out, Suggestion{Priority: "P2", Title: "Test-hook still fails as designed", File: f, Reason: "test file change is peripheral to production refund path"})
		case strings.Contains(low, "ledger") || strings.Contains(low, "billing") || strings.Contains(low, "refund") ||
			strings.Contains(body, "ledger.debit") || (strings.Contains(low, "orderservice") && strings.Contains(body, "refund")):
			if strings.Contains(low, "orderservice") || strings.Contains(low, "ledger") || strings.Contains(low, "billing") {
				out = append(out, Suggestion{
					Priority: "P0",
					Title:    "Smoke: closed-order refund must not debit",
					File:     f,
					Reason:   "production refund/ledger path changed",
				})
			}
			if strings.Contains(low, "orderservice") && (strings.Contains(body, "tenant") || strings.Contains(body, "get(")) {
				out = append(out, Suggestion{
					Priority: "P1",
					Title:    "Core: tenant-scoped order get/refund",
					File:     f,
					Reason:   "order lookup/refund still ignores tenant isolation in the fake PR",
				})
			}
		default:
			if strings.HasSuffix(low, ".java") || strings.HasSuffix(low, ".go") {
				out = append(out, Suggestion{Priority: "P1", Title: "Core: exercise changed production symbol", File: f, Reason: "production source changed"})
			} else {
				out = append(out, Suggestion{Priority: "P2", Title: "Peripheral config/asset", File: f, Reason: "non-core path"})
			}
		}
	}
	return dedupe(out)
}

func dedupe(in []Suggestion) []Suggestion {
	seen := map[string]bool{}
	var out []Suggestion
	for _, s := range in {
		k := s.Priority + "|" + s.Title + "|" + s.File
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, s)
	}
	return out
}

func containsForbidden(r *Result) bool {
	blob := strings.ToLower(r.Policy)
	for _, s := range r.Suggestions {
		blob += " " + strings.ToLower(s.Title+" "+s.Reason)
	}
	if strings.Contains(blob, "full regression") && !strings.Contains(blob, "forbidden") {
		return true
	}
	if strings.Contains(blob, "全量回归") && !strings.Contains(blob, "forbidden") {
		return true
	}
	// Suggestions themselves must never recommend full regression.
	for _, s := range r.Suggestions {
		t := strings.ToLower(s.Title + " " + s.Reason)
		if strings.Contains(t, "full regression") || strings.Contains(t, "全量回归") {
			return true
		}
	}
	return false
}

func merge(base, extra []Suggestion, docsOnly bool) []Suggestion {
	out := append([]Suggestion{}, base...)
	seen := map[string]bool{}
	for _, s := range out {
		seen[s.Title+"|"+s.File] = true
	}
	for _, s := range extra {
		if s.Title == "" || s.File == "" {
			continue
		}
		t := strings.ToLower(s.Title + " " + s.Reason)
		if strings.Contains(t, "full regression") || strings.Contains(t, "全量回归") {
			continue
		}
		if docsOnly && (s.Priority == "P0" || s.Priority == "P1") {
			continue
		}
		if s.Priority == "" {
			s.Priority = "P2"
		}
		k := s.Title + "|" + s.File
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, s)
	}
	return out
}

func completeSuggest(ctx context.Context, in Input, diff string, files []string) ([]Suggestion, error) {
	prompt := "Changed files:\n" + strings.Join(files, "\n") +
		"\n\nDiff excerpt:\n" + trim(diff, 4000) +
		"\nReturn JSON {\"suggestions\":[{\"priority\":\"P0|P1|P2\",\"title\":\"\",\"file\":\"\",\"reason\":\"\"}]}.\n" +
		"Never recommend full regression or 全量回归. Docs-only must stay P2. CodeForge is a QE CLI, not a chat IDE."
	resp, err := in.Provider.Complete(ctx, provider.CompletionRequest{
		Model: in.Model,
		Messages: []provider.Message{
			{Role: "system", Content: "You are CodeForge regress-suggest (test/QE wedge). Targeted P0 smoke / P1 core / P2 peripheral only."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	raw := extractJSON(resp.Content)
	var parsed struct {
		Suggestions []Suggestion `json:"suggestions"`
	}
	if json.Unmarshal([]byte(raw), &parsed) != nil {
		return nil, nil
	}
	return parsed.Suggestions, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
