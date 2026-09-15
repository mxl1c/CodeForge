package cli

import (
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Show credential setup (OpenAI-compatible gateways)",
		Long: `Print how to set provider credentials. Seat lifecycle is "codeforge seat".

CODEFORGE_API_KEY is required for live provider calls.
CODEFORGE_BASE_URL / CODEFORGE_MODEL select an OpenAI-compatible gateway
(not only official OpenAI). DeepSeek example:

  export CODEFORGE_BASE_URL=https://api.deepseek.com/v1
  export CODEFORGE_MODEL=deepseek-chat`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("[codeforge] login: credentials (QE wedge, not a general IDE)\n")
			cmd.Printf("export CODEFORGE_API_KEY=sk-...\n")
			cmd.Printf("export CODEFORGE_BASE_URL=https://api.deepseek.com/v1   # OpenAI-compatible default example (DeepSeek)\n")
			cmd.Printf("export CODEFORGE_MODEL=deepseek-chat\n")
			cmd.Printf("or: api_key / base_url / model in ~/.codeforge/config.yaml\n")
			cmd.Printf("seat lifecycle: codeforge seat trial|activate|suspend|status\n")
			return nil
		},
		Args: cobra.NoArgs,
	}
}
