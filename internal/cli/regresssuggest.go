package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRegressSuggestCmd() *cobra.Command {
	var diff string

	cmd := &cobra.Command{
		Use:   "regress-suggest",
		Short: "根据 PR diff 建议回归范围（W1 stub）",
		Long: `读取伪 PR diff（默认 fixtures/fake-pr.diff），建议应补的回归用例。
W1 仅脚手架。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "codeforge regress-suggest: W1 stub — 回归建议尚未实现\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  diff=%s\n", diff)
			return nil
		},
	}
	cmd.Flags().StringVar(&diff, "diff", "fixtures/fake-pr.diff", "PR diff 文件路径")
	return cmd
}
