package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Provider diagnostics (W1-05 smoke; not a product command)",
		Long: `Diagnostic commands for the OpenAI-compatible provider.

This is a diagnostic, not a QE vertical. Product/lifecycle commands:
login, init, test-gen, defect-blame, regress-suggest, seat.
W1-05 uses "provider ping" to make a single real Chat Completions Complete call.`,
		SilenceUsage: true,
	}
	cmd.AddCommand(newProviderPingCmd())
	return cmd
}

func newProviderPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "W1-05 smoke: one real Chat Completions Complete call",
		Long: `Calls the configured OpenAI-compatible Chat Completions endpoint once.

Requires CODEFORGE_API_KEY or api_key in ~/.codeforge/config.yaml.
Missing credentials: clear error, non-zero exit.
Success: prints model and token usage, exit 0.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         runProviderPing,
	}
}

func runProviderPing(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	p, err := provider.NewOpenAICompat(cfg.APIKey, cfg.BaseURL)
	if err != nil {
		return err
	}

	resp, err := p.Complete(cmd.Context(), provider.CompletionRequest{
		Model: cfg.Model,
		Messages: []provider.Message{
			{Role: "user", Content: "ping"},
		},
	})
	if err != nil {
		return err
	}

	model := resp.Model
	if model == "" {
		model = cfg.Model
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] provider ping: ok\n")
	fmt.Fprintf(cmd.OutOrStdout(), "model: %s\n", model)
	fmt.Fprintf(cmd.OutOrStdout(), "tokens: prompt=%d completion=%d total=%d\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	return nil
}
