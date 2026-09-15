// Package blame attributes a failing stack to file/function with a defect|env|test classification.
// Insufficient evidence never invents a blame location.
package blame

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mxl1c/CodeForge/internal/provider"
)

var (
	ErrNoStack              = errors.New("stack path is required")
	ErrInsufficientEvidence = errors.New("insufficient evidence: refusing to invent blame")
)

type Classification string

const (
	ClassDefect       Classification = "defect"
	ClassEnv          Classification = "env"
	ClassTest         Classification = "test"
	ClassInsufficient Classification = "insufficient"
)

type Frame struct {
	Raw      string
	Package  string
	Function string
	File     string
	Line     string
}

type Result struct {
	Mode           string
	Classification Classification
	File           string
	Function       string
	Line           string
	Evidence       []string
	Repro          []string
}

type Input struct {
	StackPath string
	Offline   bool
	Provider  provider.Provider
	Model     string
}

var (
	javaFrameRe = regexp.MustCompile(`^\s*at\s+((?:[\w$]+\.)+)([\w$]+)\(([\w$.]+):(\d+)\)`)
	goFrameRe   = regexp.MustCompile(`^\s*([^\s]+)\.(\w+)\(.*\)$`)
	goFileRe    = regexp.MustCompile(`^\s*(.+\.(?:go|java)):(\d+)`)
	ctxRe       = regexp.MustCompile(`tenant=(\S+).*orderId=(\S+).*status=(\S+)`)
)

func Run(ctx context.Context, in Input) (*Result, error) {
	if strings.TrimSpace(in.StackPath) == "" {
		return nil, ErrNoStack
	}
	data, err := os.ReadFile(in.StackPath)
	if err != nil {
		return nil, fmt.Errorf("stack: %w", err)
	}
	text := string(data)
	res := analyze(text)
	res.Mode = "offline"
	if res.Classification == ClassInsufficient || res.File == "" || res.Function == "" {
		return res, ErrInsufficientEvidence
	}

	if !in.Offline && in.Provider != nil {
		if err := enrich(ctx, in, text, res); err != nil {
			return nil, err
		}
		res.Mode = "provider"
	} else {
		res.Mode = "offline"
	}
	return res, nil
}

func analyze(text string) *Result {
	res := &Result{Classification: ClassInsufficient}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		res.Evidence = []string{"stack is empty"}
		return res
	}

	frames := parseFrames(text)
	app := firstAppFrame(frames)
	if app == nil {
		res.Classification = ClassInsufficient
		res.Evidence = []string{"no application file/function in stack; refusing to invent blame"}
		if envHint(text) {
			res.Evidence = append(res.Evidence, "environment/infrastructure exception present but not attributable")
		}
		return res
	}

	res.File = app.File
	res.Function = qualify(app)
	res.Line = app.Line
	res.Evidence = collectEvidence(text, app)

	switch {
	case testHint(text, app):
		res.Classification = ClassTest
	case envHint(text) && !businessHint(text):
		res.Classification = ClassEnv
	default:
		res.Classification = ClassDefect
	}
	res.Repro = reproSteps(text, app)
	return res
}

func qualify(f *Frame) string {
	pkg := strings.TrimSuffix(f.Package, ".")
	if pkg == "" {
		return f.Function
	}
	return pkg + "." + f.Function
}

func parseFrames(text string) []Frame {
	var frames []Frame
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if m := javaFrameRe.FindStringSubmatch(line); len(m) == 5 {
			frames = append(frames, Frame{Raw: strings.TrimSpace(line), Package: m[1], Function: m[2], File: m[3], Line: m[4]})
			continue
		}
		if m := goFrameRe.FindStringSubmatch(line); len(m) == 3 && i+1 < len(lines) {
			if fm := goFileRe.FindStringSubmatch(lines[i+1]); len(fm) == 3 {
				frames = append(frames, Frame{Raw: strings.TrimSpace(line), Package: m[1], Function: m[2], File: fm[1], Line: fm[2]})
			}
		}
	}
	return frames
}

func isAppFrame(f Frame) bool {
	p := strings.ToLower(f.Package + f.File + f.Raw)
	skip := []string{
		"org.springframework", "org.apache.catalina", "org.apache.coyote",
		"jdk.internal", "java.lang.thread", "java.lang.reflect",
		"org.junit", "org.apache.tomcat",
	}
	for _, s := range skip {
		if strings.Contains(p, s) {
			return false
		}
	}
	if strings.Contains(p, "com.example") || strings.Contains(p, "saasadmin") {
		return true
	}
	if strings.HasSuffix(strings.ToLower(f.File), ".java") && !strings.Contains(p, "java.") {
		if strings.Contains(p, "order") || strings.Contains(p, "billing") || strings.Contains(p, "tenant") {
			return true
		}
	}
	if strings.HasSuffix(strings.ToLower(f.File), ".go") && strings.Contains(p, "order") {
		return true
	}
	return false
}

func firstAppFrame(frames []Frame) *Frame {
	var caused *Frame
	for i := range frames {
		f := frames[i]
		if !isAppFrame(f) {
			continue
		}
		// Prefer the business method (refund) over filter/ledger helpers when both exist.
		fn := strings.ToLower(f.Function)
		if fn == "refund" || fn == "get" {
			return &f
		}
		if caused == nil {
			copy := f
			caused = &copy
		}
	}
	return caused
}

func envHint(text string) bool {
	l := strings.ToLower(text)
	for _, k := range []string{"sockettimeoutexception", "connectexception", "unknownhostexception", "connection refused", "timed out", "oom", "outofmemory"} {
		if strings.Contains(l, k) {
			return true
		}
	}
	return false
}

func testHint(text string, app *Frame) bool {
	l := strings.ToLower(text + app.File + app.Function)
	return strings.Contains(l, "assertionerror") || strings.Contains(l, "junit") ||
		strings.HasSuffix(strings.ToLower(app.File), "test.java") || strings.Contains(strings.ToLower(app.File), "_test.go")
}

func businessHint(text string) bool {
	l := strings.ToLower(text)
	for _, k := range []string{"退款", "refund", "tenant", "租户", "illegalstate", "ledger", "closed"} {
		if strings.Contains(l, k) {
			return true
		}
	}
	return false
}

func collectEvidence(text string, app *Frame) []string {
	var ev []string
	ev = append(ev, "frame: "+app.Raw)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "业务上下文") || strings.Contains(strings.ToLower(line), "illegalstate") ||
			strings.Contains(line, "LedgerConstraint") || strings.Contains(line, "Caused by") {
			if line != "" {
				ev = append(ev, line)
			}
		}
	}
	return ev
}

func reproSteps(text string, app *Frame) []string {
	tenant, orderID, status := "unknown-tenant", "unknown-order", "unknown-status"
	if m := ctxRe.FindStringSubmatch(text); len(m) == 4 {
		tenant, orderID, status = m[1], m[2], m[3]
	}
	fn := app.Function
	return []string{
		fmt.Sprintf("使用租户 %s，订单 %s（状态=%s）", tenant, orderID, status),
		fmt.Sprintf("调用 %s（%s:%s）", qualify(app), app.File, app.Line),
		"观察是否扣减租户可用余额或将关闭订单标为已退款",
		"期望：关闭订单不得退款；证据不足时不得编造责任人",
		"函数 " + fn + " 为堆栈中的应用帧，不是猜测",
	}
}

func enrich(ctx context.Context, in Input, text string, res *Result) error {
	prompt := "Stack (do not invent file/function not present):\n" + trim(text, 4000) +
		"\n\nLocal attribution JSON:\n" + mustJSON(map[string]string{
		"file": res.File, "function": res.Function, "classification": string(res.Classification),
	}) +
		"\nReturn JSON {\"repro\":[\"...\"],\"evidence\":[\"...\"]}. Do not change file/function. Not a general IDE."
	resp, err := in.Provider.Complete(ctx, provider.CompletionRequest{
		Model: in.Model,
		Messages: []provider.Message{
			{Role: "system", Content: "You are CodeForge QE defect-blame (not a chat IDE). Never invent blame without stack evidence."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return err
	}
	raw := extractJSON(resp.Content)
	var extra struct {
		Repro    []string `json:"repro"`
		Evidence []string `json:"evidence"`
	}
	if json.Unmarshal([]byte(raw), &extra) != nil {
		return nil
	}
	if len(extra.Repro) > 0 {
		res.Repro = append(res.Repro, extra.Repro...)
	}
	if len(extra.Evidence) > 0 {
		res.Evidence = append(res.Evidence, extra.Evidence...)
	}
	return nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
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
