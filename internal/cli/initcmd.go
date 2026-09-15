package cli

import (
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize project config (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			printStub(cmd, "W1 does not write files. Project config: .codeforge.yaml ; user config: ~/.codeforge/config.yaml")
			return nil
		},
	}
}
