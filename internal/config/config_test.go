package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndFindProjectConfig(t *testing.T) {
	dir := t.TempDir()
	p := ProjectConfig{Project: "demo", Language: "java"}
	p.Test.Framework = "junit5"
	if err := WriteProject(dir, p); err != nil {
		t.Fatal(err)
	}
	found, err := FindProjectConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(found) != projectFileName {
		t.Fatalf("found = %s", found)
	}
}

func TestEnvOverridesAPIKey(t *testing.T) {
	t.Setenv("CODEFORGE_API_KEY", "from-env")
	t.Setenv("CODEFORGE_BASE_URL", "https://example.test/v1")
	t.Setenv("CODEFORGE_MODEL", "env-model")

	// Point HOME at empty dir so user config does not leak from the runner.
	home := t.TempDir()
	t.Setenv("HOME", home)

	res, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if res.APIKey != "from-env" {
		t.Fatalf("api key = %q", res.APIKey)
	}
	if res.BaseURL != "https://example.test/v1" {
		t.Fatalf("base url = %q", res.BaseURL)
	}
	if res.Model != "env-model" {
		t.Fatalf("model = %q", res.Model)
	}
}

func TestSaveUserCreatesRestrictedFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := SaveUser(UserConfig{Provider: "openai-compatible", APIKey: "sk-file"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, userDirName, userFileName)
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("perm = %o", st.Mode().Perm())
	}
}
