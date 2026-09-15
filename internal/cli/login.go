package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/seat"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Seat lifecycle (trial → active → suspended) and credential hints",
		Long: `Manage the local seat ledger (not cloud billing).

Walk the three states:
  codeforge login            # ensure trial (or print current state)
  codeforge login status     # print ledger; does not create a seat
  codeforge login trial      # (none) → trial
  codeforge login activate   # trial → active
  codeforge login suspend    # active → suspended

Credentials remain env/config: CODEFORGE_API_KEY and optional
CODEFORGE_BASE_URL (OpenAI-compatible gateways, e.g. DeepSeek).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoginEnsureTrial(cmd)
		},
	}
	cmd.AddCommand(
		newLoginStatusCmd(),
		newLoginTrialCmd(),
		newLoginActivateCmd(),
		newLoginSuspendCmd(),
	)
	return cmd
}

func newLoginStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show local seat ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := seatStore()
			if err != nil {
				return err
			}
			rec, err := st.Load()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] login status: %s\n", rec.State)
			printSeat(cmd, rec)
			return nil
		},
	}
}

func newLoginTrialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trial",
		Short: "Start trial seat (none → trial)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return transitionSeat(cmd, seat.StateTrial, "login trial")
		},
	}
}

func newLoginActivateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "activate",
		Short: "Activate seat (trial → active)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return transitionSeat(cmd, seat.StateActive, "login activate")
		},
	}
}

func newLoginSuspendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suspend",
		Short: "Suspend seat (active → suspended)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return transitionSeat(cmd, seat.StateSuspended, "login suspend")
		},
	}
}

func runLoginEnsureTrial(cmd *cobra.Command) error {
	st, err := seatStore()
	if err != nil {
		return err
	}
	rec, err := st.EnsureTrial("login")
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] login: seat %s\n", rec.State)
	printSeat(cmd, rec)
	fmt.Fprintf(cmd.OutOrStdout(), "credentials: export CODEFORGE_API_KEY (optional CODEFORGE_BASE_URL / CODEFORGE_MODEL)\n")
	fmt.Fprintf(cmd.OutOrStdout(), "OpenAI-compatible example (DeepSeek): CODEFORGE_BASE_URL=https://api.deepseek.com/v1 CODEFORGE_MODEL=deepseek-chat\n")
	return nil
}

func transitionSeat(cmd *cobra.Command, to seat.State, via string) error {
	st, err := seatStore()
	if err != nil {
		return err
	}
	rec, err := st.Transition(to, via)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] login: %s → %s\n", lastFrom(rec), rec.State)
	printSeat(cmd, rec)
	return nil
}

func lastFrom(rec *seat.Record) seat.State {
	if rec == nil || len(rec.History) == 0 {
		return "(none)"
	}
	from := rec.History[len(rec.History)-1].From
	if from == "" {
		return "(none)"
	}
	return from
}
