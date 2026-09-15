package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/seat"
)

func newSeatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seat",
		Short: "Seat lifecycle: trial → active → suspended",
		Long: `Local seat ledger (not cloud billing): ~/.codeforge/seat.json

Locked M1 contract:
  codeforge seat trial      # (none) → trial
  codeforge seat activate   # trial → active
  codeforge seat suspend    # active → suspended
  codeforge seat status     # print ledger

Illegal jumps (e.g. trial→suspend, suspended→active) fail and do not mutate state.
Suspended seats block test-gen / defect-blame / regress-suggest.

CodeForge is a test/QE wedge, not a general IDE.`,
		SilenceUsage: true,
	}
	cmd.AddCommand(
		newSeatTrialCmd(),
		newSeatActivateCmd(),
		newSeatSuspendCmd(),
		newSeatStatusCmd(),
	)
	return cmd
}

func newSeatStatusCmd() *cobra.Command {
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
			fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] seat status: %s\n", rec.State)
			printSeat(cmd, rec)
			return nil
		},
	}
}

func newSeatTrialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trial",
		Short: "Start trial seat (none → trial)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return transitionSeat(cmd, seat.StateTrial, "seat trial")
		},
	}
}

func newSeatActivateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "activate",
		Short: "Activate seat (trial → active)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return transitionSeat(cmd, seat.StateActive, "seat activate")
		},
	}
}

func newSeatSuspendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suspend",
		Short: "Suspend seat (active → suspended)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return transitionSeat(cmd, seat.StateSuspended, "seat suspend")
		},
	}
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
	fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] seat: %s → %s\n", lastFrom(rec), rec.State)
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
