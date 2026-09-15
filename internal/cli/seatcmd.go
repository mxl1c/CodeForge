package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mxl1c/CodeForge/internal/seat"
)

func newSeatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seat",
		Short: "Seat lifecycle + offline tiers (not payment)",
		Long: `Local seat ledger (not cloud billing / not online payment): ~/.codeforge/seat.json

Lifecycle (unchanged from M1):
  codeforge seat trial      # (none) → trial
  codeforge seat activate   # trial → active
  codeforge seat suspend    # active → suspended
  codeforge seat status     # print ledger (includes tier)

Offline commercial labels (no checkout):
  codeforge seat tier              # show current tier
  codeforge seat tier free         # Free
  codeforge seat tier pro          # ~¥140
  codeforge seat tier business     # ~¥700–1400

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
		newSeatTierCmd(),
	)
	return cmd
}

func newSeatStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show local seat ledger (state + tier)",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := seatStore()
			if err != nil {
				return err
			}
			rec, err := st.Load()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] seat status: %s tier=%s (%s)\n",
				rec.State, rec.EffectiveTier(), rec.EffectiveTier().Label())
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

func newSeatTierCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tier [free|pro|business]",
		Short: "Show or set offline seat tier (not payment)",
		Long: `Show or set the local commercial label. This does not charge or open a checkout.

  free      Free
  pro       ~¥140
  business  ~¥700–1400

Lifecycle stays trial → active → suspended. Setting a tier never invents payment.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := seatStore()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				rec, err := st.Load()
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] seat tier: %s (%s)\n", rec.EffectiveTier(), rec.EffectiveTier().Label())
				printSeat(cmd, rec)
				return nil
			}
			tier, err := seat.ParseTier(args[0])
			if err != nil {
				return err
			}
			rec, err := st.SetTier(tier, "seat tier")
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[codeforge] seat tier set: %s (%s) (offline label, not payment)\n",
				rec.EffectiveTier(), rec.EffectiveTier().Label())
			printSeat(cmd, rec)
			return nil
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
