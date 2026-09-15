// Package testgen scans a Java/Go module and produces evidence-backed QE test cases.
// Empty modules fail; cases are never fabricated without production source.
package testgen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mxl1c/CodeForge/internal/provider"
)

var (
	ErrEmptyModule = errors.New("empty repo: no production source; refusing to fabricate tests")
	ErrNoModule    = errors.New("repo path is required (--repo)")
)

type Case struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Function string `json:"fn"`
	Priority string `json:"priority"`
	Intent   string `json:"intent"`
	Sketch   string `json:"sketch,omitempty"`
}

type FileScan struct {
	RelPath   string
	Lang      string
	Functions []string
	Text      string
}

type Result struct {
	Module string
	Mode   string
	Files  []FileScan
	Cases  []Case
}

type Input struct {
	ModuleDir string
	Offline   bool
	Provider  provider.Provider
	Model     string
}

var (
	goFuncRe     = regexp.MustCompile(`(?m)^func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)\s*\(`)
	javaMethodRe = regexp.MustCompile(`(?m)(?:public|protected)\s+(?:static\s+)?[\w.<>,\[\]]+\s+(\w+)\s*\(`)
)

func Run(ctx context.Context, in Input) (*Result, error) {
	if strings.TrimSpace(in.ModuleDir) == "" {
		return nil, ErrNoModule
	}
	abs, err := filepath.Abs(in.ModuleDir)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("module: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("module is not a directory: %s", abs)
	}

	files, err := scanModule(abs)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, ErrEmptyModule
	}

	out := &Result{
		Module: abs,
		Mode:   "offline",
		Files:  files,
		Cases:  casesFromScan(files),
	}
	if len(out.Cases) == 0 {
		return nil, ErrEmptyModule
	}

	if !in.Offline && in.Provider != nil {
		extra, err := completeCases(ctx, in, files)
		if err != nil {
			return nil, err
		}
		out.Mode = "provider"
		out.Cases = mergeCases(out.Cases, extra, files)
	}
	return out, nil
}

func scanModule(root string) ([]FileScan, error) {
	var files []FileScan
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" || name == "target" {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		lang := ""
		switch {
		case strings.HasSuffix(name, "_test.go"):
			return nil
		case strings.HasSuffix(name, ".go"):
			lang = "go"
		case strings.HasSuffix(name, "Test.java") || strings.HasSuffix(name, "Tests.java"):
			return nil
		case strings.HasSuffix(name, ".java"):
			lang = "java"
		default:
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		text := string(data)
		fns := functionsOf(lang, text)
		if len(fns) == 0 {
			return nil
		}
		files = append(files, FileScan{RelPath: filepath.ToSlash(rel), Lang: lang, Functions: fns, Text: text})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].RelPath < files[j].RelPath })
	return files, nil
}

func functionsOf(lang, text string) []string {
	var re *regexp.Regexp
	if lang == "go" {
		re = goFuncRe
	} else {
		re = javaMethodRe
	}
	seen := map[string]bool{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(text, -1) {
		fn := m[1]
		if fn == "" || fn == "main" || seen[fn] {
			continue
		}
		if lang == "java" && (fn == fnClassName(text) || fn == "toString" || fn == "hashCode" || fn == "equals") {
			continue
		}
		seen[fn] = true
		out = append(out, fn)
	}
	return out
}

func fnClassName(text string) string {
	re := regexp.MustCompile(`(?m)^\s*(?:public\s+)?(?:final\s+)?class\s+(\w+)`)
	m := re.FindStringSubmatch(text)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func casesFromScan(files []FileScan) []Case {
	var cases []Case
	for _, f := range files {
		for _, fn := range f.Functions {
			if c, ok := caseFor(f, fn); ok {
				cases = append(cases, c)
			}
		}
	}
	return cases
}

func caseFor(f FileScan, fn string) (Case, bool) {
	lower := strings.ToLower(fn)
	textLower := strings.ToLower(f.Text)
	switch {
	case strings.Contains(lower, "refund") || strings.Contains(lower, "debit"):
		c := Case{
			Name:     "RefundRejectsClosedOrder",
			File:     f.RelPath,
			Function: fn,
			Priority: "P0",
			Intent:   "closed/已关闭 orders must not refund or debit tenant balance (CF-W1-001); evidence from scanned " + fn,
			Sketch:   sketchRefund(f.Lang, fn),
		}
		return c, true
	case lower == "get" || strings.HasPrefix(lower, "get"):
		if strings.Contains(textLower, "tenant") {
			c := Case{
				Name:     "GetIsTenantScoped",
				File:     f.RelPath,
				Function: fn,
				Priority: "P1",
				Intent:   "order lookup must not return another tenant's order; evidence from scanned " + fn,
				Sketch:   sketchTenantGet(f.Lang, fn),
			}
			return c, true
		}
	}
	return Case{}, false
}

func sketchRefund(lang, fn string) string {
	if lang == "java" {
		return "seed closed order; " + fn + "(tenant, orderId) must throw; must not mark refunded"
	}
	return "Seed(closed); " + fn + " must return error; Status must stay closed"
}

func sketchTenantGet(lang, fn string) string {
	if lang == "java" {
		return "seed tenant A order; " + fn + "(tenantB, orderId) must be null"
	}
	return "Seed tenant A; " + fn + "(tenantB, id) must be nil"
}

func mergeCases(base, extra []Case, files []FileScan) []Case {
	index := map[string]bool{}
	for _, f := range files {
		for _, fn := range f.Functions {
			index[f.RelPath+"|"+fn] = true
			index[filepath.Base(f.RelPath)+"|"+fn] = true
		}
	}
	out := append([]Case{}, base...)
	seen := map[string]bool{}
	for _, c := range out {
		seen[c.Name+"|"+c.File+"|"+c.Function] = true
	}
	for _, c := range extra {
		if c.Name == "" || c.File == "" || c.Function == "" {
			continue
		}
		if !index[c.File+"|"+c.Function] && !index[filepath.Base(c.File)+"|"+c.Function] {
			continue // drop hallucinations
		}
		key := c.Name + "|" + c.File + "|" + c.Function
		if seen[key] {
			continue
		}
		seen[key] = true
		if c.Priority == "" {
			c.Priority = "P2"
		}
		out = append(out, c)
	}
	return out
}

type llmCases struct {
	Cases []Case `json:"cases"`
}

func completeCases(ctx context.Context, in Input, files []FileScan) ([]Case, error) {
	var b strings.Builder
	b.WriteString("Scanned production symbols (do not invent others):\n")
	for _, f := range files {
		fmt.Fprintf(&b, "- %s (%s) fns=%s\n", f.RelPath, f.Lang, strings.Join(f.Functions, ","))
	}
	b.WriteString("\nReturn JSON {\"cases\":[{\"name\":\"\",\"file\":\"\",\"fn\":\"\",\"priority\":\"P0|P1|P2\",\"intent\":\"\"}]}.\n")
	b.WriteString("Only QE cases for SaaS mid-office order/refund/tenant. Not a general IDE.\n")

	resp, err := in.Provider.Complete(ctx, provider.CompletionRequest{
		Model: in.Model,
		Messages: []provider.Message{
			{Role: "system", Content: "You are CodeForge, an enterprise test/QE CLI (not a general IDE or chat). Propose tests only for scanned symbols. JSON only."},
			{Role: "user", Content: b.String()},
		},
	})
	if err != nil {
		return nil, err
	}
	raw := extractJSON(resp.Content)
	var parsed llmCases
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, nil // ignore unparseable model output; local cases remain
	}
	return parsed.Cases, nil
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
