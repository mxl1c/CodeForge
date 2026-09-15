package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/config"
	"github.com/mxl1c/CodeForge/internal/provider"
	"github.com/mxl1c/CodeForge/internal/seat"
	"github.com/mxl1c/CodeForge/internal/usage"
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

func usageStore() (*usage.Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return usage.Open(home), nil
}

func requireSeat(via string) (*seat.Record, error) {
	st, err := seatStore()
	if err != nil {
		return nil, err
	}
	return st.CheckUsable(via)
}

func printSeat(cmd *cobra.Command, rec *seat.Record) {
	tier := rec.EffectiveTier()
	fmt.Fprintf(cmd.OutOrStdout(), "seat: id=%s state=%s tenant=%s tier=%s (%s)\n",
		rec.SeatID, rec.State, rec.Tenant, tier, tier.Label())
	fmt.Fprintf(cmd.OutOrStdout(), "seat tiers (offline labels, not payment): free=Free  pro=~¥140  business=~¥700–1400\n")
	fmt.Fprintf(cmd.OutOrStdout(), "seat lifecycle: trial → active → suspended (local ledger ~/.codeforge/seat.json)\n")
}

type qeSession struct {
	rec *seat.Record
	cfg *config.Config
	cap *usage.Capture
}

func prepareQE(cmd *cobra.Command, command string) (*qeSession, provider.Provider, error) {
	sess := &qeSession{}
	rec, err := requireSeat(command)
	sess.rec = rec
	if err != nil {
		return sess, nil, err
	}
	offline := isOffline(cmd)
	p, cfg, err := loadProvider(offline)
	sess.cfg = cfg
	if err != nil {
		return sess, nil, err
	}
	sess.cap = usage.Wrap(p)
	return sess, sess.cap.Provider(), nil
}

func flushUsage(command string, sess *qeSession, runErr error) error {
	if sess == nil {
		return nil
	}
	st, err := usageStore()
	if err != nil {
		return err
	}
	e := usage.Entry{
		Command:  command,
		Provider: usage.ProviderOffline,
		Model:    usage.ModelOffline,
		Status:   usageStatus(runErr),
	}
	if sess.rec != nil {
		e.Tenant = sess.rec.Tenant
		e.SeatID = sess.rec.SeatID
		e.Tier = string(sess.rec.EffectiveTier())
	}
	if sess.cap != nil && sess.cap.Called {
		e.Provider = usage.ProviderCompat
		e.Model = sess.cap.Model
		if e.Model == "" && sess.cfg != nil {
			e.Model = sess.cfg.Model
		}
		e.PromptTokens = sess.cap.PromptTokens
		e.CompletionTokens = sess.cap.CompletionTokens
	}
	return st.Append(e)
}

func usageStatus(err error) string {
	if err == nil {
		return usage.StatusOK
	}
	if errors.Is(err, seat.ErrSuspended) {
		return usage.StatusBlocked
	}
	return usage.StatusError
}
