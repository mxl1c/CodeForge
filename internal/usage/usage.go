// Package usage is a self-built local usage ledger for pilot ARPU.
// It is not cloud billing, online payment, or a marketplace.
package usage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ProviderOffline = "offline"
	ProviderCompat  = "openai-compat"
	ModelOffline    = "offline"

	StatusOK      = "ok"
	StatusError   = "error"
	StatusBlocked = "blocked"
)

// Entry is one QE command invocation, enough to compute pilot ARPU.
type Entry struct {
	Timestamp        time.Time `json:"timestamp"`
	Tenant           string    `json:"tenant"`
	SeatID           string    `json:"seat_id"`
	Tier             string    `json:"tier"`
	Command          string    `json:"command"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	Status           string    `json:"status"`
}

var CSVHeader = []string{
	"timestamp",
	"tenant",
	"seat_id",
	"tier",
	"command",
	"provider",
	"model",
	"prompt_tokens",
	"completion_tokens",
	"status",
}

type Store struct {
	path string
}

func PathForHome(home string) string {
	return filepath.Join(home, ".codeforge", "usage.jsonl")
}

func Open(home string) *Store {
	return &Store{path: PathForHome(home)}
}

func (s *Store) Append(e Entry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	e.Timestamp = e.Timestamp.UTC()
	if e.Provider == "" {
		e.Provider = ProviderOffline
	}
	if e.Model == "" {
		e.Model = ModelOffline
	}
	if e.Status == "" {
		e.Status = StatusOK
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("usage dir: %w", err)
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open usage ledger: %w", err)
	}
	defer f.Close()
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write usage ledger: %w", err)
	}
	return nil
}

func (s *Store) LoadAll() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read usage ledger: %w", err)
	}
	var out []Entry
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("parse usage ledger line %d: %w", i+1, err)
		}
		out = append(out, e)
	}
	return out, nil
}

func NormalizeFormat(format, outPath string) (string, error) {
	f := strings.ToLower(strings.TrimSpace(format))
	if f == "" {
		switch strings.ToLower(filepath.Ext(outPath)) {
		case ".json":
			f = "json"
		default:
			f = "csv"
		}
	}
	switch f {
	case "csv", "json":
		return f, nil
	default:
		return "", fmt.Errorf("unsupported usage export format %q (csv or json)", format)
	}
}

func WriteExport(w io.Writer, entries []Entry, format string) error {
	switch format {
	case "csv":
		cw := csv.NewWriter(w)
		if err := cw.Write(CSVHeader); err != nil {
			return err
		}
		for _, e := range entries {
			if err := cw.Write(e.csvRow()); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case "json":
		if entries == nil {
			entries = []Entry{}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(entries); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported usage export format %q (csv or json)", format)
	}
}

func (s *Store) ExportToFile(path, format string) error {
	format, err := NormalizeFormat(format, path)
	if err != nil {
		return err
	}
	entries, err := s.LoadAll()
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("export dir: %w", err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("export file: %w", err)
	}
	defer f.Close()
	return WriteExport(f, entries, format)
}

func (e Entry) csvRow() []string {
	ts := e.Timestamp.UTC().Format(time.RFC3339)
	return []string{
		ts,
		e.Tenant,
		e.SeatID,
		e.Tier,
		e.Command,
		e.Provider,
		e.Model,
		strconv.Itoa(e.PromptTokens),
		strconv.Itoa(e.CompletionTokens),
		e.Status,
	}
}
