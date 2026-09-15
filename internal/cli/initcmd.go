package cli

import (
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Show project config locations (QE wedge, not a general IDE)",
		Long: `Prints where CodeForge reads config. Does not write files.

Project: .codeforge.yaml (walks up from cwd)
User:    ~/.codeforge/config.yaml
Seat:    ~/.codeforge/seat.json via "codeforge seat trial|activate|suspend|status|tier"
Usage:   ~/.codeforge/usage.jsonl via "codeforge usage export --out <file>"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("[codeforge] init: QE project config (no files written)\n")
			cmd.Printf("project: .codeforge.yaml\n")
			cmd.Printf("user: ~/.codeforge/config.yaml\n")
			cmd.Printf("seat ledger: ~/.codeforge/seat.json  (codeforge seat trial|activate|suspend|status|tier)\n")
			cmd.Printf("usage ledger: ~/.codeforge/usage.jsonl  (codeforge usage export --out arpu.csv)\n")
			cmd.Printf("CodeForge is a test/quality wedge, not a general IDE or chat.\n")
			return nil
		},
	}
}
