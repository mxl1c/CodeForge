package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/testgen"
)

func newTestGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-gen",
		Short: "Generate QE tests from a Java/Go module (no fabricated cases)",
		Long: `Scan production source under --module and emit evidence-backed test cases.

Golden: samples/java-saas-admin or samples/go-saas-admin
Failure: empty module exits non-zero and does not invent tests.

With CODEFORGE_API_KEY (and not --offline), one provider Complete call may add
cases; symbols not present in the scan are dropped.`,
		RunE: runTestGen,
	}
	cmd.Flags().String("module", "", "path to Java or Go module (e.g. samples/go-saas-admin)")
	addOfflineFlag(cmd)
	return cmd
}

func runTestGen(cmd *cobra.Command, _ []string) error {
	if _, err := requireSeat("test-gen"); err != nil {
		return err
	}
	module, _ := cmd.Flags().GetString("module")
	offline := isOffline(cmd)
	p, cfg, err := loadProvider(offline)
	if err != nil {
		return err
	}
	model := ""
	if cfg != nil {
		model = cfg.Model
	}
	res, err := testgen.Run(cmd.Context(), testgen.Input{
		ModuleDir: module,
		Offline:   offline || p == nil,
		Provider:  p,
		Model:     model,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] test-gen: ok\n")
	fmt.Fprintf(cmd.OutOrStdout(), "module: %s\n", res.Module)
	fmt.Fprintf(cmd.OutOrStdout(), "mode: %s\n", res.Mode)
	fmt.Fprintf(cmd.OutOrStdout(), "cases: %d\n", len(res.Cases))
	for _, c := range res.Cases {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s %s file=%s fn=%s\n  %s\n  sketch: %s\n",
			c.Priority, c.Name, c.File, c.Function, c.Intent, c.Sketch)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "wedge: test/QE CLI (not a general IDE)\n")
	return nil
}
