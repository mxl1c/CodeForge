package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
	"github.com/mxl1c/CodeForge/internal/seat"
)

func addOfflineFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("offline", false, "deterministic local analysis (CI/fixture mode; do not call the provider)")
}

func isOffline(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("offline")
	if v {
		return true
	}
	env := strings.TrimSpace(os.Getenv("CODEFORGE_OFFLINE"))
	return env == "1" || strings.EqualFold(env, "true")
}

func loadProvider(offline bool) (provider.Provider, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	if offline || strings.TrimSpace(cfg.APIKey) == "" {
		return nil, cfg, nil
	}
	p, err := provider.NewOpenAICompat(cfg.APIKey, cfg.BaseURL)
	if err != nil {
		return nil, cfg, err
	}
	return p, cfg, nil
}

func seatStore() (*seat.Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return seat.Open(home), nil
}

func requireSeat(via string) (*seat.Record, error) {
	st, err := seatStore()
	if err != nil {
		return nil, err
	}
	return st.CheckUsable(via)
}

func printSeat(cmd *cobra.Command, rec *seat.Record) {
	fmt.Fprintf(cmd.OutOrStdout(), "seat: id=%s state=%s tenant=%s\n", rec.SeatID, rec.State, rec.Tenant)
	fmt.Fprintf(cmd.OutOrStdout(), "seat lifecycle: trial → active → suspended (local ledger ~/.codeforge/seat.json)\n")
}
