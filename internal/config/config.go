package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-4o-mini"

	EnvAPIKey  = "CODEFORGE_API_KEY"
	EnvBaseURL = "CODEFORGE_BASE_URL"
	EnvModel   = "CODEFORGE_MODEL"

	userConfigRel    = ".codeforge/config.yaml"
	projectConfigRel = ".codeforge.yaml"
)

// Config is the merged CodeForge configuration.
// Precedence: defaults < ~/.codeforge/config.yaml < .codeforge.yaml < env.
type Config struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cwd: %w", err)
	}
	return LoadFrom(home, cwd)
}

func LoadFrom(home, startDir string) (*Config, error) {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
		Model:   DefaultModel,
	}

	userPath := filepath.Join(home, userConfigRel)
	if err := mergeFile(cfg, userPath); err != nil {
		return nil, err
	}

	if projectPath := findProjectConfig(startDir); projectPath != "" {
		if err := mergeFile(cfg, projectPath); err != nil {
			return nil, err
		}
	}

	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(EnvModel); v != "" {
		cfg.Model = v
	}
	return cfg, nil
}

func findProjectConfig(startDir string) string {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, projectConfigRel)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func mergeFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}
	var next Config
	if err := yaml.Unmarshal(data, &next); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if next.APIKey != "" {
		cfg.APIKey = next.APIKey
	}
	if next.BaseURL != "" {
		cfg.BaseURL = next.BaseURL
	}
	if next.Model != "" {
		cfg.Model = next.Model
	}
	return nil
}
