package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
)

func newLoginCmd() *cobra.Command {
	var (
		apiKey  string
		baseURL string
		model   string
		prov    string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store provider credentials in ~/.codeforge/config.yaml",
		Long: `login writes an OpenAI-compatible API key (and optional base URL / model)
to ~/.codeforge/config.yaml. The file is created with mode 0600.

You can also skip this command and export:

  export CODEFORGE_API_KEY=sk-...
  export CODEFORGE_BASE_URL=https://api.openai.com/v1   # optional
  export CODEFORGE_MODEL=gpt-4o-mini                   # optional

Environment variables always override the user config file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if apiKey == "" {
				fmt.Fprint(cmd.OutOrStdout(), "API key (CODEFORGE_API_KEY): ")
				line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if err != nil && len(strings.TrimSpace(line)) == 0 {
					return fmt.Errorf("read api key: %w", err)
				}
				apiKey = strings.TrimSpace(line)
			}
			if apiKey == "" {
				return fmt.Errorf("api key is required (flag --api-key, stdin, or skip and use CODEFORGE_API_KEY)")
			}

			existing, _ := config.Load()
			u := config.UserConfig{
				Provider: prov,
				APIKey:   apiKey,
				BaseURL:  baseURL,
				Model:    model,
			}
			if u.Provider == "" {
				u.Provider = existing.Provider
			}
			if u.BaseURL == "" {
				u.BaseURL = existing.BaseURL
			}
			if u.Model == "" {
				u.Model = existing.Model
			}

			if err := config.SaveUser(u); err != nil {
				return err
			}
			path, _ := config.UserConfigPath()
			fmt.Fprintf(cmd.OutOrStdout(), "saved credentials to %s\n", path)
			fmt.Fprintln(cmd.OutOrStdout(), "provider:", u.Provider)
			fmt.Fprintln(cmd.OutOrStdout(), "base_url:", u.BaseURL)
			fmt.Fprintln(cmd.OutOrStdout(), "model:", u.Model)
			return nil
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (otherwise prompted)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "OpenAI-compatible base URL")
	cmd.Flags().StringVar(&model, "model", "", "default model name")
	cmd.Flags().StringVar(&prov, "provider", "openai-compatible", "provider id (openai-compatible)")
	return cmd
}
