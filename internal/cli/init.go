package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
)

func newInitCmd() *cobra.Command {
	var (
		project   string
		language  string
		framework string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a project-level .codeforge.yaml",
		Long: `init writes .codeforge.yaml in the current directory so CodeForge can
target this repo's language and test framework. User credentials stay in
~/.codeforge/config.yaml; this file is project-scoped QA metadata only.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			if project == "" {
				project = filepath.Base(wd)
			}
			p := config.ProjectConfig{
				Project:  project,
				Language: language,
			}
			p.Test.Framework = framework
			if err := config.WriteProject(wd, p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", filepath.Join(wd, ".codeforge.yaml"))
			fmt.Fprintf(cmd.OutOrStdout(), "project=%s language=%s test.framework=%s\n", project, language, framework)
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "project name (default: directory name)")
	cmd.Flags().StringVar(&language, "language", "go", "primary language (go|java)")
	cmd.Flags().StringVar(&framework, "framework", "testing", "test framework (testing|junit5|testng)")
	return cmd
}
