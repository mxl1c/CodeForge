package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func newTestGenCmd() *cobra.Command {
	var (
		target string
		live   bool
	)

	cmd := &cobra.Command{
		Use:   "test-gen",
		Short: "Generate tests for a source path (W1: skeleton + optional live model call)",
		Long: `test-gen is the W1 entry point for the test-generation wedge.

Without a configured API key the command prints a placeholder plan and
skips the live model call. With CODEFORGE_API_KEY (or codeforge login)
and --live, it performs one real OpenAI-compatible Chat Completions request.

Skip live call (default when no key):
  codeforge test-gen --path samples/java-saas-admin

Live call (requires key):
  export CODEFORGE_API_KEY=sk-...
  export CODEFORGE_BASE_URL=https://api.openai.com/v1   # optional
  codeforge test-gen --path samples/go-saas-admin --live`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if target == "" {
				target = "."
			}
			abs, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "[test-gen] W1 skeleton — test generation for %s\n", abs)
			fmt.Fprintln(cmd.OutOrStdout(), "[test-gen] planned: map sources → unit tests, keep fixtures under testdata/")

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			p, err := provider.New(cfg.ProviderConfig())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[test-gen] provider=%s model=%s base_url=%s\n", p.Name(), cfg.Model, cfg.BaseURL)

			if !live || !cfg.HasAPIKey() {
				if !cfg.HasAPIKey() {
					fmt.Fprintln(cmd.OutOrStdout(), "[test-gen] skipping live model call: no API key")
					fmt.Fprintln(cmd.OutOrStdout(), "[test-gen] set CODEFORGE_API_KEY or run `codeforge login`, then pass --live")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "[test-gen] skipping live model call (pass --live to send one Chat Completions request)")
				}
				fmt.Fprintln(cmd.OutOrStdout(), "[test-gen] placeholder: would emit tests next to sources (not written in W1)")
				return nil
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 45*time.Second)
			defer cancel()

			snippet := readPreview(abs, 4000)
			resp, err := p.Chat(ctx, provider.ChatRequest{
				Model: cfg.Model,
				Messages: []provider.Message{
					{Role: "system", Content: "You are CodeForge, an enterprise test/QA coding agent. Propose a short unit-test plan. Do not write a general chat reply."},
					{Role: "user", Content: "Generate a concise test plan for this path:\n" + snippet},
				},
			})
			if err != nil {
				if errors.Is(err, provider.ErrNoAPIKey) {
					fmt.Fprintln(cmd.OutOrStdout(), "[test-gen] skipping live model call:", err)
					return nil
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[test-gen] live response from %s:\n%s\n", resp.Model, strings.TrimSpace(resp.Content))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "path", "", "source path (file or directory)")
	cmd.Flags().BoolVar(&live, "live", false, "perform a real model API call when an API key is configured")
	return cmd
}

func readPreview(path string, limit int) string {
	st, err := os.Stat(path)
	if err != nil {
		return path
	}
	if st.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return path
		}
		var b strings.Builder
		fmt.Fprintf(&b, "directory %s\n", path)
		for i, e := range entries {
			if i >= 20 {
				break
			}
			fmt.Fprintf(&b, "- %s\n", e.Name())
		}
		return b.String()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return path
	}
	if len(raw) > limit {
		raw = raw[:limit]
	}
	return string(raw)
}
