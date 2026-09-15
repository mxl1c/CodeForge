package cli_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mxl1c/CodeForge/internal/cli"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func run(args ...string) (string, error) {
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestRootHelpListsFiveCommands(t *testing.T) {
	out, err := run("--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"login", "init", "test-gen", "defect-blame", "regress-suggest"} {
		if !strings.Contains(out, name) {
			t.Fatalf("help missing %q:\n%s", name, out)
		}
	}
}

func TestSubcommandHelp(t *testing.T) {
	for _, name := range []string{"login", "init", "test-gen", "defect-blame", "regress-suggest", "provider"} {
		out, err := run(name, "--help")
		if err != nil {
			t.Fatalf("%s --help: %v", name, err)
		}
		if !strings.Contains(out, "Usage:") {
			t.Fatalf("%s help has no Usage:\n%s", name, out)
		}
	}
}

func TestStubsExitCleanly(t *testing.T) {
	cases := [][]string{
		{"login"},
		{"test-gen"},
		{"defect-blame"},
		{"regress-suggest"},
	}
	for _, args := range cases {
		if _, err := run(args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

func TestInitWritesProjectConfig(t *testing.T) {
	dir := t.TempDir()
	out, err := run("init", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, ".codeforge.yaml") {
		t.Fatalf("out=%s", out)
	}
	out2, err := run("init", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out2, "未覆盖") {
		t.Fatalf("second init: %s", out2)
	}
}

func TestLoginWritesUserConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_HOME", home)
	out, err := run("login", "--api-key", "sk-test")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "已写入用户配置") {
		t.Fatalf("out=%s", out)
	}
}

func TestProviderPingMissingKey(t *testing.T) {
	t.Setenv("CODEFORGE_HOME", t.TempDir())
	t.Setenv("CODEFORGE_API_KEY", "")
	_, err := run("provider", "ping")
	if err == nil {
		t.Fatal("expected missing key error")
	}
	if !errors.Is(err, provider.ErrMissingAPIKey) {
		t.Fatalf("err=%v", err)
	}
}

func TestProviderPingLiveCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"gpt-4o-mini","choices":[{"message":{"role":"assistant","content":"pong"}}]}`))
	}))
	defer srv.Close()

	t.Setenv("CODEFORGE_HOME", t.TempDir())
	t.Setenv("CODEFORGE_API_KEY", "sk-live")
	t.Setenv("CODEFORGE_BASE_URL", srv.URL)

	out, err := run("provider", "ping")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "provider ping ok") {
		t.Fatalf("out=%s", out)
	}
}
