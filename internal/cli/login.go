package cli

import (
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			printStub(cmd, "W1 auth is a no-op. Export CODEFORGE_API_KEY (optional CODEFORGE_BASE_URL) or set api_key in ~/.codeforge/config.yaml.")
			return nil
		},
	}
}
