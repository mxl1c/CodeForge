package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeforge",
		Short: "CodeForge — 企业测试/QE Agent CLI（席位制，非通用对话、非 IDE）",
		Long: `CodeForge 面向中国企业中后台的测试与质量工程（QE）垂直场景。
提供测试生成、缺陷归因与回归建议，按席位交付。本周仅 CLI，不含 IDE 插件。`,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	cmd.AddCommand(
		newLoginCmd(),
		newInitCmd(),
		newTestGenCmd(),
		newDefectBlameCmd(),
		newRegressSuggestCmd(),
		newProviderCmd(),
	)
	return cmd
}
