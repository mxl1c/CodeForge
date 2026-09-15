package cli

import (
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeforge",
		Short: "CodeForge CLI — test/quality wedge (not a general IDE)",
		Long: `CodeForge is a CLI for test generation, defect attribution, and regression suggestions
on Java/Go SaaS mid-office services.

It is not a general-purpose IDE or chat interface.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(
		newLoginCmd(),
		newInitCmd(),
		newTestGenCmd(),
		newDefectBlameCmd(),
		newRegressSuggestCmd(),
		newSeatCmd(),
		newUsageCmd(),
		newProviderCmd(),
	)
	return cmd
}

func Execute() error {
	return NewRoot().Execute()
}
