package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom_DefaultsAndEnv(t *testing.T) {
	t.Setenv(EnvAPIKey, "env-key")
	t.Setenv(EnvBaseURL, "https://gateway.example/v1")
	t.Setenv(EnvModel, "deepseek-chat")

	home := t.TempDir()
	start := t.TempDir()

	cfg, err := LoadFrom(home, start)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "env-key" {
		t.Fatalf("APIKey=%q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://gateway.example/v1" {
		t.Fatalf("BaseURL=%q", cfg.BaseURL)
	}
	if cfg.Model != "deepseek-chat" {
		t.Fatalf("Model=%q", cfg.Model)
	}
}

func TestLoadFrom_UserThenProject(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvModel, "")

	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".codeforge"), 0o755); err != nil {
		t.Fatal(err)
	}
	userYAML := []byte("api_key: user-key\nbase_url: https://user.example/v1\nmodel: user-model\n")
	if err := os.WriteFile(filepath.Join(home, ".codeforge", "config.yaml"), userYAML, 0o600); err != nil {
		t.Fatal(err)
	}

	start := t.TempDir()
	projYAML := []byte("base_url: https://project.example/v1\nmodel: project-model\n")
	if err := os.WriteFile(filepath.Join(start, ".codeforge.yaml"), projYAML, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(home, start)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "user-key" {
		t.Fatalf("APIKey=%q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://project.example/v1" {
		t.Fatalf("BaseURL=%q", cfg.BaseURL)
	}
	if cfg.Model != "project-model" {
		t.Fatalf("Model=%q", cfg.Model)
	}
}
