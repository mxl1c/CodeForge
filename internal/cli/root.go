package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0-w1"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeforge",
		Short: "CodeForge — enterprise coding agent CLI (test/QA wedge)",
		Long: `CodeForge is an enterprise coding agent CLI focused on test and quality
engineering. W1 ships command skeletons, a switchable model provider, and
sample/fixture layouts. This is not a general chat app or IDE.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newLoginCmd(),
		newInitCmd(),
		newTestGenCmd(),
		newDefectBlameCmd(),
		newRegressSuggestCmd(),
	)
	return cmd
}

// Execute runs the CodeForge CLI.
func Execute() error {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return err
	}
	return nil
}
