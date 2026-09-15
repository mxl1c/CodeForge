package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/blame"
)

func newDefectBlameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defect-blame",
		Short: "Attribute a failure stack to file/function (no invented blame)",
		Long: `Read --stack, locate file/function, classify defect vs env/test, and print repro steps.

Golden: fixtures/failure-stack-zh.txt
Failure: fixtures/insufficient-stack.txt (or empty) exits non-zero without inventing a culprit.`,
		RunE: runDefectBlame,
	}
	cmd.Flags().String("stack", "", "path to failure stack (e.g. fixtures/failure-stack-zh.txt)")
	addOfflineFlag(cmd)
	return cmd
}

func runDefectBlame(cmd *cobra.Command, _ []string) error {
	if _, err := requireSeat("defect-blame"); err != nil {
		return err
	}
	stack, _ := cmd.Flags().GetString("stack")
	offline := isOffline(cmd)
	p, cfg, err := loadProvider(offline)
	if err != nil {
		return err
	}
	model := ""
	if cfg != nil {
		model = cfg.Model
	}
	res, err := blame.Run(cmd.Context(), blame.Input{
		StackPath: stack,
		Offline:   offline || p == nil,
		Provider:  p,
		Model:     model,
	})
	if res != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] defect-blame: %s\n", res.Classification)
		fmt.Fprintf(cmd.OutOrStdout(), "mode: %s\n", res.Mode)
		fmt.Fprintf(cmd.OutOrStdout(), "location: file=%s fn=%s line=%s\n", res.File, res.Function, res.Line)
		fmt.Fprintf(cmd.OutOrStdout(), "classification: %s\n", res.Classification)
		for _, e := range res.Evidence {
			fmt.Fprintf(cmd.OutOrStdout(), "evidence: %s\n", e)
		}
		for i, step := range res.Repro {
			fmt.Fprintf(cmd.OutOrStdout(), "repro %d: %s\n", i+1, step)
		}
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "refusing to invent blame\n")
		}
	}
	return err
}
