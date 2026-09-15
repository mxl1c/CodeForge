package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDefectBlameCmd() *cobra.Command {
	var stack string

	cmd := &cobra.Command{
		Use:   "defect-blame",
		Short: "根据失败栈做缺陷归因（W1 stub）",
		Long: `读取业务失败栈（默认中文中后台样例 fixtures/failure-stack-zh.txt），
定位可能的提交/模块。W1 仅脚手架。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "codeforge defect-blame: W1 stub — 缺陷归因尚未实现\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  stack=%s\n", stack)
			return nil
		},
	}
	cmd.Flags().StringVar(&stack, "stack", "fixtures/failure-stack-zh.txt", "失败栈文件路径")
	return cmd
}
