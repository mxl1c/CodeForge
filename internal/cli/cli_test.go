package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootListsAllFiveSubcommands(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, name := range []string{"login", "init", "test-gen", "defect-blame", "regress-suggest"} {
		if !strings.Contains(out, name) {
			t.Fatalf("help missing subcommand %s\n%s", name, out)
		}
	}
}

func TestTestGenSkipsLiveCallWithoutKey(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEFORGE_API_KEY", "")
	t.Setenv("CODEFORGE_BASE_URL", "")
	t.Setenv("CODEFORGE_MODEL", "")

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"test-gen", "--path", "."})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "skipping live model call") {
		t.Fatalf("expected skip message, got:\n%s", out)
	}
	if !strings.Contains(out, "provider=openai-compatible") {
		t.Fatalf("expected provider line, got:\n%s", out)
	}
}

func TestDefectBlameReadsFixture(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := repoRoot(t)
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"defect-blame", "--stack", filepath.Join(root, "fixtures", "failure-stack-zh.txt")})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "审批") && !strings.Contains(out, "placeholder suspects") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestRegressSuggestReadsFixture(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := repoRoot(t)
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"regress-suggest", "--diff", filepath.Join(root, "fixtures", "fake-pr.diff")})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "placeholder cases") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestInitWritesProjectYAML(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"init", "--project", "acme", "--language", "java", "--framework", "junit5"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".codeforge.yaml")); err != nil {
		t.Fatal(err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
