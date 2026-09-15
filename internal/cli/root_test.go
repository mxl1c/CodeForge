package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestStubCommandsExitOK(t *testing.T) {
	t.Setenv("CODEFORGE_API_KEY", "")
	t.Setenv("CODEFORGE_BASE_URL", "")

	commands := []string{"login", "init", "test-gen", "defect-blame", "regress-suggest"}
	for _, name := range commands {
		t.Run(name, func(t *testing.T) {
			root := NewRoot()
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs([]string{name})
			if err := root.Execute(); err != nil {
				t.Fatalf("execute: %v\n%s", err, buf.String())
			}
			out := buf.String()
			if !strings.Contains(out, "stub (W1 skeleton)") {
				t.Fatalf("missing stub output: %q", out)
			}
		})
	}
}

func TestTestGenReportsMissingAPIKey(t *testing.T) {
	t.Setenv("CODEFORGE_API_KEY", "")
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"test-gen"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "missing API key") {
		t.Fatalf("output=%q", buf.String())
	}
}
