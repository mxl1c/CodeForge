package seat

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLifecycleTrialActiveSuspended(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	rec, err := st.Transition(StateTrial, "login")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateTrial {
		t.Fatalf("state=%s", rec.State)
	}

	rec, err = st.Transition(StateActive, "seat activate")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateActive {
		t.Fatalf("state=%s", rec.State)
	}

	rec, err = st.Transition(StateSuspended, "seat suspend")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateSuspended {
		t.Fatalf("state=%s", rec.State)
	}
	if len(rec.History) != 3 {
		t.Fatalf("history=%d", len(rec.History))
	}

	if _, err := os.Stat(filepath.Join(home, ".codeforge", "seat.json")); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidTransitions(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	if _, err := st.Transition(StateActive, "x"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("none→active: %v", err)
	}
	if _, err := st.Transition(StateSuspended, "x"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("none→suspended: %v", err)
	}

	if _, err := st.Transition(StateTrial, "login"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Transition(StateSuspended, "x"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("trial→suspended: %v", err)
	}
	if _, err := st.Transition(StateTrial, "x"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("trial→trial: %v", err)
	}

	if _, err := st.Transition(StateActive, "seat activate"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Transition(StateTrial, "x"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("active→trial: %v", err)
	}

	if _, err := st.Transition(StateSuspended, "seat suspend"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Transition(StateActive, "x"); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("suspended→active: %v", err)
	}
}

func TestCheckUsable(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	rec, err := st.CheckUsable("test-gen")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateTrial {
		t.Fatalf("auto trial: %s", rec.State)
	}

	if _, err := st.Transition(StateActive, "seat activate"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CheckUsable("test-gen"); err != nil {
		t.Fatal(err)
	}

	if _, err := st.Transition(StateSuspended, "seat suspend"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CheckUsable("test-gen"); !errors.Is(err, ErrSuspended) {
		t.Fatalf("want ErrSuspended, got %v", err)
	}
}

func TestDefaultTierFreeAndLabels(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	rec, err := st.Transition(StateTrial, "login")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Tier != TierFree {
		t.Fatalf("default tier=%s", rec.Tier)
	}
	if rec.EffectiveTier().Label() != "Free" {
		t.Fatalf("label=%s", rec.EffectiveTier().Label())
	}
	if rec.State != StateTrial {
		t.Fatalf("lifecycle must stay trial: %s", rec.State)
	}
}

func TestSetTierIndependentOfLifecycle(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	if _, err := st.Transition(StateTrial, "seat trial"); err != nil {
		t.Fatal(err)
	}
	rec, err := st.SetTier(TierPro, "seat tier")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateTrial {
		t.Fatalf("setting tier must not change lifecycle: %s", rec.State)
	}
	if rec.Tier != TierPro || rec.EffectiveTier().Label() != "~¥140" {
		t.Fatalf("pro: tier=%s label=%s", rec.Tier, rec.EffectiveTier().Label())
	}

	rec, err = st.Transition(StateActive, "seat activate")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Tier != TierPro {
		t.Fatalf("activate must keep pro: %s", rec.Tier)
	}

	rec, err = st.SetTier(TierBusiness, "seat tier")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateActive {
		t.Fatalf("business set changed lifecycle: %s", rec.State)
	}
	if rec.EffectiveTier().Label() != "~¥700–1400" {
		t.Fatalf("business label=%s", rec.EffectiveTier().Label())
	}

	rec, err = st.Transition(StateSuspended, "seat suspend")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Tier != TierBusiness {
		t.Fatalf("suspend dropped tier: %s", rec.Tier)
	}
}

func TestSetTierCreatesTrialAndRejectsUnknown(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	rec, err := st.SetTier(TierPro, "seat tier")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateTrial {
		t.Fatalf("missing seat should open trial: %s", rec.State)
	}
	if rec.Tier != TierPro {
		t.Fatalf("tier=%s", rec.Tier)
	}

	if _, err := st.SetTier(Tier("enterprise"), "x"); !errors.Is(err, ErrInvalidTier) {
		t.Fatalf("want ErrInvalidTier, got %v", err)
	}
	rec, err = st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Tier != TierPro {
		t.Fatalf("invalid tier mutated ledger: %s", rec.Tier)
	}
}

func TestParseTierAndLegacyEmptyDefaultsFree(t *testing.T) {
	if _, err := ParseTier("FREE"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTier("paid"); !errors.Is(err, ErrInvalidTier) {
		t.Fatalf("paid: %v", err)
	}

	home := t.TempDir()
	st := Open(home)
	legacy := []byte(`{"tenant":"local","seat_id":"seat-local","state":"trial"}` + "\n")
	if err := os.MkdirAll(filepath.Join(home, ".codeforge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".codeforge", "seat.json"), legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	rec, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if rec.EffectiveTier() != TierFree {
		t.Fatalf("legacy empty tier=%q", rec.Tier)
	}
}
