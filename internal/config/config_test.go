package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mxl1c/CodeForge/internal/config"
)

func TestResolvePrefersEnvOverUserFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_HOME", home)
	t.Setenv("CODEFORGE_API_KEY", "env-key")
	t.Setenv("CODEFORGE_BASE_URL", "https://gateway.example/v1")

	if err := config.SaveUser(config.UserConfig{
		APIKey:  "file-key",
		BaseURL: "https://file.example/v1",
		Model:   "file-model",
	}); err != nil {
		t.Fatal(err)
	}

	got, err := config.Resolve(".")
	if err != nil {
		t.Fatal(err)
	}
	if got.APIKey != "env-key" {
		t.Fatalf("APIKey=%q, want env-key", got.APIKey)
	}
	if got.BaseURL != "https://gateway.example/v1" {
		t.Fatalf("BaseURL=%q", got.BaseURL)
	}
	if got.Model != "file-model" {
		t.Fatalf("Model=%q, want file-model", got.Model)
	}
}

func TestLoadUserMissingFileIsEmpty(t *testing.T) {
	t.Setenv("CODEFORGE_HOME", t.TempDir())
	cfg, err := config.LoadUser()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "" {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}

func TestWriteAndLoadProject(t *testing.T) {
	dir := t.TempDir()
	path := config.ProjectConfigPath(dir)
	if err := config.WriteProjectExample(path); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Wedge != "test-qe" {
		t.Fatalf("wedge=%q", cfg.Wedge)
	}
	if cfg.Paths.Fixtures["failure_stack"] != "fixtures/failure-stack-zh.txt" {
		t.Fatalf("fixtures: %+v", cfg.Paths.Fixtures)
	}
}

func TestSaveUserCreatesRestrictedFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEFORGE_HOME", home)
	if err := config.SaveUser(config.UserConfig{APIKey: "secret"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "config.yaml")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm=%o, want 0600", info.Mode().Perm())
	}
}
