// Package seat is a minimal local seat ledger: trial → active → suspended.
package seat

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State string

const (
	StateTrial     State = "trial"
	StateActive    State = "active"
	StateSuspended State = "suspended"
)

var (
	ErrInvalidTransition = errors.New("invalid seat transition")
	ErrSuspended         = errors.New("seat suspended: vertical QE commands are blocked until the seat is no longer suspended")
	ErrNoSeat            = errors.New("no seat ledger")
)

// Record is the on-disk seat ledger (not cloud billing).
type Record struct {
	Tenant    string       `json:"tenant"`
	SeatID    string       `json:"seat_id"`
	State     State        `json:"state"`
	UpdatedAt time.Time    `json:"updated_at"`
	History   []Transition `json:"history"`
}

type Transition struct {
	From State     `json:"from"`
	To   State     `json:"to"`
	At   time.Time `json:"at"`
	Via  string    `json:"via"`
}

type Store struct {
	path string
}

func PathForHome(home string) string {
	return filepath.Join(home, ".codeforge", "seat.json")
}

func Open(home string) *Store {
	return &Store{path: PathForHome(home)}
}

func (s *Store) Load() (*Record, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoSeat
		}
		return nil, fmt.Errorf("read seat ledger: %w", err)
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parse seat ledger: %w", err)
	}
	return &rec, nil
}

func (s *Store) save(rec *Record) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("seat dir: %w", err)
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write seat ledger: %w", err)
	}
	return nil
}

func allowed(from State, to State) bool {
	switch {
	case from == "" && to == StateTrial:
		return true
	case from == StateTrial && to == StateActive:
		return true
	case from == StateActive && to == StateSuspended:
		return true
	default:
		return false
	}
}

func (s *Store) Transition(to State, via string) (*Record, error) {
	rec, err := s.Load()
	var from State
	if errors.Is(err, ErrNoSeat) {
		rec = &Record{Tenant: "local", SeatID: "seat-local"}
		from = ""
	} else if err != nil {
		return nil, err
	} else {
		from = rec.State
	}
	if !allowed(from, to) {
		cur := from
		if cur == "" {
			cur = "(none)"
		}
		return nil, fmt.Errorf("%w: %s → %s (allowed: trial → active → suspended)", ErrInvalidTransition, cur, to)
	}
	now := time.Now().UTC()
	rec.State = to
	rec.UpdatedAt = now
	rec.History = append(rec.History, Transition{From: from, To: to, At: now, Via: via})
	if err := s.save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// EnsureTrial creates a trial seat if none exists. Existing seats are left unchanged.
func (s *Store) EnsureTrial(via string) (*Record, error) {
	rec, err := s.Load()
	if errors.Is(err, ErrNoSeat) {
		return s.Transition(StateTrial, via)
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// CheckUsable loads the seat (creating trial if missing) and rejects suspended seats.
func (s *Store) CheckUsable(via string) (*Record, error) {
	rec, err := s.EnsureTrial(via)
	if err != nil {
		return nil, err
	}
	if rec.State == StateSuspended {
		return rec, ErrSuspended
	}
	return rec, nil
}
