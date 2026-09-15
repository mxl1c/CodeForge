package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTestGenCmd() *cobra.Command {
	var path string
	var lang string

	cmd := &cobra.Command{
		Use:   "test-gen",
		Short: "为中后台代码生成测试（W1 stub）",
		Long: `扫描 Java/Go SaaS 中后台代码并生成单元/接口测试草稿。
W1 仅脚手架：校验参数后干净退出，不调用模型。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "codeforge test-gen: W1 stub — QE 测试生成尚未实现\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  path=%s lang=%s\n", path, lang)
			fmt.Fprintf(cmd.OutOrStdout(), "  样例路径: samples/java-saas-admin/  samples/go-saas-admin/\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "待生成测试的源码路径")
	cmd.Flags().StringVar(&lang, "lang", "", "java 或 go；空则按路径推断")
	return cmd
}
