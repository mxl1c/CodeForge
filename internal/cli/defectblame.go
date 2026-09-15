package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/blame"
)

const defaultLogPath = "fixtures/failure-stack-zh.txt"

func newDefectBlameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defect-blame",
		Short: "Attribute a failure log to file/function (no invented blame)",
		Long: `Read --log, locate file/function, classify defect vs env/test, and print repro steps.

Locked flag: --log <path>  (default: fixtures/failure-stack-zh.txt)

Golden: fixtures/failure-stack-zh.txt
Failure: fixtures/insufficient-stack.txt (or empty) exits non-zero without inventing a culprit.`,
		RunE: runDefectBlame,
	}
	cmd.Flags().String("log", defaultLogPath, "path to failure log/stack")
	addOfflineFlag(cmd)
	return cmd
}

func runDefectBlame(cmd *cobra.Command, _ []string) (err error) {
	sess, p, err := prepareQE(cmd, "defect-blame")
	defer func() {
		if uerr := flushUsage("defect-blame", sess, err); uerr != nil && err == nil {
			err = uerr
		}
	}()
	if err != nil {
		return err
	}
	logPath, _ := cmd.Flags().GetString("log")
	offline := isOffline(cmd)
	model := ""
	if sess.cfg != nil {
		model = sess.cfg.Model
	}
	res, err := blame.Run(cmd.Context(), blame.Input{
		StackPath: logPath,
		Offline:   offline || p == nil,
		Provider:  p,
		Model:     model,
	})
	if res != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] defect-blame: %s\n", res.Classification)
		fmt.Fprintf(cmd.OutOrStdout(), "mode: %s\n", res.Mode)
		fmt.Fprintf(cmd.OutOrStdout(), "log: %s\n", logPath)
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
