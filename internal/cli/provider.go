package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
	"github.com/spf13/cobra"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "LLM Provider（OpenAI 兼容适配器）",
	}
	cmd.AddCommand(newProviderPingCmd())
	return cmd
}

func newProviderPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "对当前 Provider 发起一次真实探测调用",
		Long: `读取 CODEFORGE_API_KEY（或 ~/.codeforge/config.yaml 中的 api_key）后，
向 OpenAI 兼容 /chat/completions 发起一次短请求。未配置 Key 时返回明确错误。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Resolve(".")
			if err != nil {
				return err
			}
			p, err := provider.NewFromResolved(cfg)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			if err := p.Ping(ctx); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "provider ping ok: %s (base_url=%s model=%s)\n", p.Name(), cfg.BaseURL, cfg.Model)
			return nil
		},
	}
}
