// Package seat is a minimal local seat ledger: trial → active → suspended,
// with offline commercial tier labels (not online payment).
package seat

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type State string

const (
	StateTrial     State = "trial"
	StateActive    State = "active"
	StateSuspended State = "suspended"
)

// Tier is an offline commercial label on the local ledger.
// There is no checkout, gateway, or payment flow.
type Tier string

const (
	TierFree     Tier = "free"
	TierPro      Tier = "pro"
	TierBusiness Tier = "business"
)

var (
	ErrInvalidTransition = errors.New("invalid seat transition")
	ErrSuspended         = errors.New("seat suspended: vertical QE commands are blocked until the seat is no longer suspended")
	ErrNoSeat            = errors.New("no seat ledger")
	ErrInvalidTier       = errors.New("invalid seat tier: use free, pro, or business (offline labels, not payment)")
)

// Record is the on-disk seat ledger (not cloud billing).
type Record struct {
	Tenant    string       `json:"tenant"`
	SeatID    string       `json:"seat_id"`
	State     State        `json:"state"`
	Tier      Tier         `json:"tier"`
	UpdatedAt time.Time    `json:"updated_at"`
	History   []Transition `json:"history"`
}

type Transition struct {
	From State     `json:"from"`
	To   State     `json:"to"`
	At   time.Time `json:"at"`
	Via  string    `json:"via"`
	Tier Tier      `json:"tier,omitempty"`
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

func ParseTier(s string) (Tier, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "free":
		return TierFree, nil
	case "pro":
		return TierPro, nil
	case "business":
		return TierBusiness, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidTier, s)
	}
}

// Label is the human-readable price band. These are labels only — not invoices.
func (t Tier) Label() string {
	switch t.Normalize() {
	case TierFree:
		return "Free"
	case TierPro:
		return "~¥140"
	case TierBusiness:
		return "~¥700–1400"
	default:
		return string(t)
	}
}

func (t Tier) Normalize() Tier {
	if t == "" {
		return TierFree
	}
	return t
}

func (r *Record) EffectiveTier() Tier {
	if r == nil {
		return TierFree
	}
	return r.Tier.Normalize()
}

func newRecord() *Record {
	return &Record{Tenant: "local", SeatID: "seat-local", Tier: TierFree}
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
	rec.Tier = rec.Tier.Normalize()
	if rec.Tenant == "" {
		rec.Tenant = "local"
	}
	if rec.SeatID == "" {
		rec.SeatID = "seat-local"
	}
	return &rec, nil
}

func (s *Store) save(rec *Record) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("seat dir: %w", err)
	}
	rec.Tier = rec.Tier.Normalize()
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
		rec = newRecord()
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
	rec.Tier = rec.Tier.Normalize()
	rec.UpdatedAt = now
	rec.History = append(rec.History, Transition{From: from, To: to, At: now, Via: via, Tier: rec.Tier})
	if err := s.save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// SetTier writes an offline commercial label. It does not charge, invoice, or change lifecycle.
func (s *Store) SetTier(tier Tier, via string) (*Record, error) {
	parsed, err := ParseTier(string(tier))
	if err != nil {
		return nil, err
	}
	rec, err := s.EnsureTrial(via)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rec.Tier = parsed
	rec.UpdatedAt = now
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
