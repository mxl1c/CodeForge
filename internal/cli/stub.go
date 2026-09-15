package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func printStub(cmd *cobra.Command, extra string) {
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] %s: stub (W1 skeleton)\n", cmd.Name())
	if extra != "" {
		fmt.Fprintln(cmd.OutOrStdout(), extra)
	}
}

func printProviderStatus(cmd *cobra.Command) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "config: %v\n", err)
		return
	}
	p, err := provider.NewOpenAICompat(cfg.APIKey, cfg.BaseURL)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "provider: %v\n", err)
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "provider: openai-compat ready (base_url=%s model=%s)\n", p.BaseURL(), cfg.Model)
}
