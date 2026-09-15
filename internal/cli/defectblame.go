package cli

import (
	"github.com/spf13/cobra"
)

func newDefectBlameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "defect-blame",
		Short: "Attribute failing tests/defects (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			printStub(cmd, "W2+ will attribute failures using fixtures/failure-stack-zh.txt.")
			printProviderStatus(cmd)
			return nil
		},
	}
}
