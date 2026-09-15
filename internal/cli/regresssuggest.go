package cli

import (
	"github.com/spf13/cobra"
)

func newRegressSuggestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "regress-suggest",
		Short: "Suggest regression tests (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			printStub(cmd, "W2+ will suggest regression coverage from fixtures/fake-pr.diff.")
			printProviderStatus(cmd)
			return nil
		},
	}
}
