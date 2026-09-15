package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
)

func newRegressSuggestCmd() *cobra.Command {
	var diffPath string

	cmd := &cobra.Command{
		Use:   "regress-suggest",
		Short: "Suggest regression tests from a PR diff (W1 skeleton)",
		Long: `regress-suggest reads a unified diff (W1 ships fixtures/fake-pr.diff)
and prints a placeholder list of regression cases. Later weeks will map
changed symbols onto the Java/Go SaaS admin samples.

Example:
  codeforge regress-suggest --diff fixtures/fake-pr.diff`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if diffPath == "" {
				diffPath = "fixtures/fake-pr.diff"
			}
			raw, err := os.ReadFile(diffPath)
			if err != nil {
				return fmt.Errorf("read diff: %w", err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			p, err := provider.New(cfg.ProviderConfig())
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "[regress-suggest] W1 skeleton — reviewing %s\n", diffPath)
			fmt.Fprintf(cmd.OutOrStdout(), "[regress-suggest] provider=%s (no live call in this command yet)\n", p.Name())
			fmt.Fprintf(cmd.OutOrStdout(), "[regress-suggest] diff stats: %d files hinted, %d lines\n",
				countDiffFiles(string(raw)), strings.Count(string(raw), "\n")+1)
			fmt.Fprintln(cmd.OutOrStdout(), "[regress-suggest] placeholder cases:")
			fmt.Fprintln(cmd.OutOrStdout(), "  - 审批通过后库存不足应回滚订单状态")
			fmt.Fprintln(cmd.OutOrStdout(), "  - 跨租户管理员不能审批本租户以外的订单")
			fmt.Fprintln(cmd.OutOrStdout(), "  - 折扣字段为负时拒绝提交并记录审计")
			return nil
		},
	}

	cmd.Flags().StringVar(&diffPath, "diff", "", "path to unified diff (default fixtures/fake-pr.diff)")
	return cmd
}

func countDiffFiles(diff string) int {
	n := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			n++
		}
	}
	return n
}
