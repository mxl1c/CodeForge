package cli

import (
	"fmt"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var apiKey string
	var baseURL string
	var model string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "写入席位凭据到 ~/.codeforge/config.yaml",
		Long: `将 OpenAI 兼容 API Key 写入用户配置（~/.codeforge/config.yaml）。
也可直接导出 CODEFORGE_API_KEY；可选 CODEFORGE_BASE_URL 指向兼容网关。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			existing, err := config.LoadUser()
			if err != nil {
				return err
			}
			if apiKey == "" {
				apiKey = existing.APIKey
			}
			if apiKey == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "W1 stub: 未提供 API Key。可任选其一：")
				fmt.Fprintln(cmd.OutOrStdout(), "  1. codeforge login --api-key <KEY>")
				fmt.Fprintln(cmd.OutOrStdout(), "  2. export CODEFORGE_API_KEY=<KEY>")
				fmt.Fprintln(cmd.OutOrStdout(), "  3. 在 ~/.codeforge/config.yaml 设置 api_key")
				return nil
			}
			cfg := config.UserConfig{
				APIKey:  apiKey,
				BaseURL: firstFlag(baseURL, existing.BaseURL, config.DefaultBaseURL),
				Model:   firstFlag(model, existing.Model, config.DefaultModel),
			}
			if err := config.SaveUser(cfg); err != nil {
				return err
			}
			path, _ := config.UserConfigPath()
			fmt.Fprintf(cmd.OutOrStdout(), "已写入用户配置: %s（api_key 已保存，不会回显）\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "OpenAI 兼容 API Key")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "可选，兼容网关 Base URL（亦可用 CODEFORGE_BASE_URL）")
	cmd.Flags().StringVar(&model, "model", "", "默认模型名")
	return cmd
}

func firstFlag(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
