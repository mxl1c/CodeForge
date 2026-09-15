// Package config loads user (~/.codeforge/config.yaml) and project
// (.codeforge.yaml) settings. Environment variables win over files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/mxl1c/CodeForge/internal/provider"
)

const (
	userDirName     = ".codeforge"
	userFileName    = "config.yaml"
	projectFileName = ".codeforge.yaml"

	envAPIKey  = "CODEFORGE_API_KEY"
	envBaseURL = "CODEFORGE_BASE_URL"
	envModel   = "CODEFORGE_MODEL"
)

// UserConfig is stored at ~/.codeforge/config.yaml (credentials + defaults).
type UserConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
	Model    string `yaml:"model"`
}

// ProjectConfig is stored at <repo>/.codeforge.yaml (per-project QA settings).
type ProjectConfig struct {
	Project  string `yaml:"project"`
	Language string `yaml:"language"`
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`
	Test     struct {
		Framework string `yaml:"framework"`
	} `yaml:"test"`
}

// Resolved is the merged runtime configuration.
type Resolved struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
	Project  ProjectConfig
	UserPath string
	ProjPath string
}

// UserConfigPath returns ~/.codeforge/config.yaml.
func UserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, userDirName, userFileName), nil
}

// UserDir returns ~/.codeforge.
func UserDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, userDirName), nil
}

// FindProjectConfig walks from start (or cwd) up to the filesystem root looking
// for .codeforge.yaml.
func FindProjectConfig(start string) (string, error) {
	dir := start
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}
	for {
		candidate := filepath.Join(dir, projectFileName)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// Load merges env, user config, and optional project config.
func Load() (Resolved, error) {
	res := Resolved{
		Provider: "openai-compatible",
		Model:    "gpt-4o-mini",
		BaseURL:  "https://api.openai.com/v1",
	}

	userPath, err := UserConfigPath()
	if err != nil {
		return res, err
	}
	res.UserPath = userPath

	if raw, err := os.ReadFile(userPath); err == nil {
		var u UserConfig
		if err := yaml.Unmarshal(raw, &u); err != nil {
			return res, fmt.Errorf("parse %s: %w", userPath, err)
		}
		applyUser(&res, u)
	} else if !errors.Is(err, os.ErrNotExist) {
		return res, fmt.Errorf("read %s: %w", userPath, err)
	}

	if projPath, err := FindProjectConfig(""); err == nil {
		res.ProjPath = projPath
		raw, err := os.ReadFile(projPath)
		if err != nil {
			return res, fmt.Errorf("read %s: %w", projPath, err)
		}
		var p ProjectConfig
		if err := yaml.Unmarshal(raw, &p); err != nil {
			return res, fmt.Errorf("parse %s: %w", projPath, err)
		}
		res.Project = p
		if p.Provider != "" {
			res.Provider = p.Provider
		}
		if p.Model != "" {
			res.Model = p.Model
		}
	}

	if v := os.Getenv(envAPIKey); v != "" {
		res.APIKey = v
	}
	if v := os.Getenv(envBaseURL); v != "" {
		res.BaseURL = v
	}
	if v := os.Getenv(envModel); v != "" {
		res.Model = v
	}

	return res, nil
}

func applyUser(res *Resolved, u UserConfig) {
	if u.Provider != "" {
		res.Provider = u.Provider
	}
	if u.APIKey != "" {
		res.APIKey = u.APIKey
	}
	if u.BaseURL != "" {
		res.BaseURL = u.BaseURL
	}
	if u.Model != "" {
		res.Model = u.Model
	}
}

// SaveUser writes ~/.codeforge/config.yaml, creating the directory if needed.
func SaveUser(u UserConfig) error {
	dir, err := UserDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	path := filepath.Join(dir, userFileName)
	raw, err := yaml.Marshal(u)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// WriteProject writes .codeforge.yaml into dir.
func WriteProject(dir string, p ProjectConfig) error {
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = wd
	}
	path := filepath.Join(dir, projectFileName)
	raw, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// ProviderConfig maps resolved settings onto the provider factory.
func (r Resolved) ProviderConfig() provider.Config {
	return provider.Config{
		Provider: r.Provider,
		APIKey:   r.APIKey,
		BaseURL:  r.BaseURL,
		Model:    r.Model,
	}
}

// HasAPIKey reports whether a live model call can be attempted.
func (r Resolved) HasAPIKey() bool {
	return r.APIKey != ""
}
