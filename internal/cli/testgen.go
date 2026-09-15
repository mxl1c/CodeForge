package cli

import (
	"github.com/spf13/cobra"
)

func newTestGenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test-gen",
		Short: "Generate tests (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			printStub(cmd, "W2 will generate tests for Java/Go SaaS mid-office samples.")
			printProviderStatus(cmd)
			return nil
		},
	}
}
