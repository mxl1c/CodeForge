package cli

import (
	"fmt"
	"os"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [dir]",
		Short: "在当前目录生成项目 .codeforge.yaml",
		Long:  "初始化当前仓库的 CodeForge 项目配置。面向测试/QE 垂直，不创建通用聊天应用。可选参数为目录，默认当前目录。",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			path := config.ProjectConfigPath(dir)
			if _, err := os.Stat(path); err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "已存在 %s，未覆盖\n", path)
				return nil
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			if err := config.WriteProjectExample(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "已创建 %s\n", path)
			return nil
		},
	}
}
