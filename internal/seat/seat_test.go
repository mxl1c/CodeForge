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
