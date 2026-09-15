package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/usage"
)

func newUsageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Self-built usage ledger for pilot ARPU (not online payment)",
		Long: `Local usage ledger (~/.codeforge/usage.jsonl) for pilot ARPU.

Fields: timestamp, tenant, seat_id, tier, command, provider, model,
prompt_tokens, completion_tokens, status.

Offline/stub QE commands still write deterministic rows (tokens=0, provider=offline).
When a provider Complete call happens, prompt/completion tokens are taken from the response.

This is not cloud billing, a Skill marketplace, or a general IDE.

  codeforge usage export --out arpu.csv
  codeforge usage export --out arpu.json --format json`,
		SilenceUsage: true,
	}
	cmd.AddCommand(newUsageExportCmd())
	return cmd
}

func newUsageExportCmd() *cobra.Command {
	var (
		out    string
		format string
	)
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export usage ledger to CSV or JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return fmt.Errorf("usage export requires --out <file>")
			}
			st, err := usageStore()
			if err != nil {
				return err
			}
			format, err = usage.NormalizeFormat(format, out)
			if err != nil {
				return err
			}
			if err := st.ExportToFile(out, format); err != nil {
				return err
			}
			entries, err := st.LoadAll()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] usage export: ok\n")
			fmt.Fprintf(cmd.OutOrStdout(), "file: %s\n", out)
			fmt.Fprintf(cmd.OutOrStdout(), "format: %s\n", format)
			fmt.Fprintf(cmd.OutOrStdout(), "rows: %d\n", len(entries))
			fmt.Fprintf(cmd.OutOrStdout(), "wedge: test/QE CLI (not a general IDE); ledger is self-built, not payment\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output file path (csv or json)")
	cmd.Flags().StringVar(&format, "format", "", "csv or json (default: from --out extension, else csv)")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}
