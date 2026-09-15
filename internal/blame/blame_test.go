package blame

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestGoldenFailureStack(t *testing.T) {
	path := filepath.Join(repoRoot(t), "fixtures", "failure-stack-zh.txt")
	res, err := Run(context.Background(), Input{StackPath: path, Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Classification != ClassDefect {
		t.Fatalf("class=%s", res.Classification)
	}
	if !strings.Contains(res.File, "OrderService.java") {
		t.Fatalf("file=%s", res.File)
	}
	if !strings.Contains(res.Function, "refund") {
		t.Fatalf("fn=%s", res.Function)
	}
	if len(res.Repro) == 0 {
		t.Fatal("missing repro")
	}
	joined := strings.Join(res.Repro, "\n")
	if !strings.Contains(joined, "ORD-20260915-0088") {
		t.Fatalf("repro should use stack order id: %s", joined)
	}
}

func TestInsufficientEmptyStack(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(p, []byte(" \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Input{StackPath: p, Offline: true})
	if !errors.Is(err, ErrInsufficientEvidence) {
		t.Fatalf("got res=%v err=%v", res, err)
	}
	if res == nil || res.Classification != ClassInsufficient {
		t.Fatalf("class=%v", res)
	}
	if res.File != "" || res.Function != "" {
		t.Fatalf("must not invent blame: file=%q fn=%q", res.File, res.Function)
	}
}

func TestInsufficientFixture(t *testing.T) {
	path := filepath.Join(repoRoot(t), "fixtures", "insufficient-stack.txt")
	res, err := Run(context.Background(), Input{StackPath: path, Offline: true})
	if !errors.Is(err, ErrInsufficientEvidence) {
		t.Fatalf("got %v", err)
	}
	if res.File != "" || strings.Contains(res.Function, "OrderService") {
		t.Fatalf("invented blame %+v", res)
	}
}

func TestInsufficientEnvOnlyNoAppFrame(t *testing.T) {
	p := filepath.Join(t.TempDir(), "timeout.txt")
	stack := "java.net.SocketTimeoutException: connect timed out\n" +
		"\tat java.net.Socket.connect(Socket.java:123)\n" +
		"\tat java.lang.Thread.run(Thread.java:840)\n"
	if err := os.WriteFile(p, []byte(stack), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Input{StackPath: p, Offline: true})
	if !errors.Is(err, ErrInsufficientEvidence) {
		t.Fatalf("got %v res=%+v", err, res)
	}
	if res.File != "" {
		t.Fatalf("invented file %q", res.File)
	}
	if strings.Contains(res.File+res.Function, "OrderService") {
		t.Fatal("must not invent OrderService")
	}
}

func TestNoStackPath(t *testing.T) {
	_, err := Run(context.Background(), Input{})
	if !errors.Is(err, ErrNoStack) {
		t.Fatalf("got %v", err)
	}
}
