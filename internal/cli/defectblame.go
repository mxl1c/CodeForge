package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func newDefectBlameCmd() *cobra.Command {
	var stackPath string

	cmd := &cobra.Command{
		Use:   "defect-blame",
		Short: "Map a failure stack to likely owning code (W1 skeleton)",
		Long: `defect-blame reads a failure stack (Chinese business-context fixture
in W1) and prints a placeholder blame report. Later weeks will rank likely
files/methods using repo samples.

Example:
  codeforge defect-blame --stack fixtures/failure-stack-zh.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if stackPath == "" {
				stackPath = "fixtures/failure-stack-zh.txt"
			}
			raw, err := os.ReadFile(stackPath)
			if err != nil {
				return fmt.Errorf("read stack: %w", err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			p, err := provider.New(cfg.ProviderConfig())
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "[defect-blame] W1 skeleton — analyzing %s\n", stackPath)
			fmt.Fprintf(cmd.OutOrStdout(), "[defect-blame] provider=%s (no live call in this command yet)\n", p.Name())
			fmt.Fprintln(cmd.OutOrStdout(), "[defect-blame] excerpt:")
			fmt.Fprintln(cmd.OutOrStdout(), indent(firstLines(string(raw), 8)))
			fmt.Fprintln(cmd.OutOrStdout(), "[defect-blame] placeholder suspects:")
			fmt.Fprintln(cmd.OutOrStdout(), "  1. OrderApprovalService.approve  (permission / 审批状态)")
			fmt.Fprintln(cmd.OutOrStdout(), "  2. TenantGuard.assertSameTenant  (跨租户访问)")
			fmt.Fprintln(cmd.OutOrStdout(), "  3. AuditLogInterceptor.after     (审计落库失败掩盖原异常)")
			return nil
		},
	}

	cmd.Flags().StringVar(&stackPath, "stack", "", "path to failure stack (default fixtures/failure-stack-zh.txt)")
	return cmd
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = "  " + l
		}
	}
	return strings.Join(lines, "\n")
}
