package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/regress"
)

func newRegressSuggestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regress-suggest",
		Short: "Suggest P0 smoke / P1 core / P2 peripheral regression (no full suite)",
		Long: `Read --diff and suggest targeted regression. Full regression / 全量回归 is forbidden.

Golden: fixtures/fake-pr.diff
Failure: empty diff exits non-zero; docs-only diffs stay P2 and do not escalate.`,
		RunE: runRegressSuggest,
	}
	cmd.Flags().String("diff", "", "path to unified diff (e.g. fixtures/fake-pr.diff)")
	addOfflineFlag(cmd)
	return cmd
}

func runRegressSuggest(cmd *cobra.Command, _ []string) error {
	if _, err := requireSeat("regress-suggest"); err != nil {
		return err
	}
	diff, _ := cmd.Flags().GetString("diff")
	offline := isOffline(cmd)
	p, cfg, err := loadProvider(offline)
	if err != nil {
		return err
	}
	model := ""
	if cfg != nil {
		model = cfg.Model
	}
	res, err := regress.Run(cmd.Context(), regress.Input{
		DiffPath: diff,
		Offline:  offline || p == nil,
		Provider: p,
		Model:    model,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] regress-suggest: ok\n")
	fmt.Fprintf(cmd.OutOrStdout(), "mode: %s\n", res.Mode)
	fmt.Fprintf(cmd.OutOrStdout(), "policy: %s\n", res.Policy)
	if res.DocsOnly {
		fmt.Fprintf(cmd.OutOrStdout(), "docs-only: true (no P0/P1 escalation)\n")
	}
	for _, s := range res.Suggestions {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s %s file=%s\n  %s\n", s.Priority, s.Title, s.File, s.Reason)
	}
	return nil
}
