package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-4o-mini"
	UserDirName    = ".codeforge"
	UserFileName   = "config.yaml"
	ProjectFile    = ".codeforge.yaml"
)

// UserConfig is stored at ~/.codeforge/config.yaml (or $CODEFORGE_HOME/config.yaml).
type UserConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

// ProjectConfig is stored at the project root as .codeforge.yaml.
type ProjectConfig struct {
	Project struct {
		Name      string   `yaml:"name"`
		Languages []string `yaml:"languages"`
	} `yaml:"project"`
	Wedge string `yaml:"wedge"`
	Paths struct {
		Samples  map[string]string `yaml:"samples"`
		Fixtures map[string]string `yaml:"fixtures"`
	} `yaml:"paths"`
}

// Resolved is the merged runtime configuration.
type Resolved struct {
	APIKey  string
	BaseURL string
	Model   string
	User    UserConfig
	Project ProjectConfig
}

func UserDir() (string, error) {
	if home := os.Getenv("CODEFORGE_HOME"); home != "" {
		return home, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, UserDirName), nil
}

func UserConfigPath() (string, error) {
	dir, err := UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, UserFileName), nil
}

func ProjectConfigPath(dir string) string {
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, ProjectFile)
}

func LoadUser() (UserConfig, error) {
	path, err := UserConfigPath()
	if err != nil {
		return UserConfig{}, err
	}
	return loadUserFile(path)
}

func loadUserFile(path string) (UserConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UserConfig{}, nil
		}
		return UserConfig{}, fmt.Errorf("read user config %s: %w", path, err)
	}
	var cfg UserConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return UserConfig{}, fmt.Errorf("parse user config %s: %w", path, err)
	}
	return cfg, nil
}

func LoadProject(dir string) (ProjectConfig, error) {
	path := ProjectConfigPath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProjectConfig{}, nil
		}
		return ProjectConfig{}, fmt.Errorf("read project config %s: %w", path, err)
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ProjectConfig{}, fmt.Errorf("parse project config %s: %w", path, err)
	}
	return cfg, nil
}

func Resolve(projectDir string) (Resolved, error) {
	user, err := LoadUser()
	if err != nil {
		return Resolved{}, err
	}
	project, err := LoadProject(projectDir)
	if err != nil {
		return Resolved{}, err
	}

	resolved := Resolved{
		APIKey:  firstNonEmpty(os.Getenv("CODEFORGE_API_KEY"), user.APIKey),
		BaseURL: firstNonEmpty(os.Getenv("CODEFORGE_BASE_URL"), user.BaseURL, DefaultBaseURL),
		Model:   firstNonEmpty(user.Model, DefaultModel),
		User:    user,
		Project: project,
	}
	return resolved, nil
}

func SaveUser(cfg UserConfig) error {
	path, err := UserConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal user config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write user config: %w", err)
	}
	return nil
}

func WriteProjectExample(path string) error {
	const body = `# CodeForge 项目配置 — 测试/QE 垂直（非通用对话 Agent）
project:
  name: my-project
  languages:
    - java
    - go
wedge: test-qe
paths:
  samples:
    java: samples/java-saas-admin
    go: samples/go-saas-admin
  fixtures:
    failure_stack: fixtures/failure-stack-zh.txt
    fake_pr: fixtures/fake-pr.diff
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write project config: %w", err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
