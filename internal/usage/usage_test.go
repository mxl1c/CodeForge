package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mxl1c/CodeForge/internal/provider"
)

func TestAppendExportCSVAndJSON(t *testing.T) {
	home := t.TempDir()
	st := Open(home)

	ts := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	if err := st.Append(Entry{
		Timestamp:        ts,
		Tenant:           "local",
		SeatID:           "seat-local",
		Tier:             "pro",
		Command:          "test-gen",
		Provider:         ProviderOffline,
		Model:            ModelOffline,
		PromptTokens:     0,
		CompletionTokens: 0,
		Status:           StatusOK,
	}); err != nil {
		t.Fatal(err)
	}

	csvPath := filepath.Join(t.TempDir(), "arpu.csv")
	if err := st.ExportToFile(csvPath, "csv"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, strings.Join(CSVHeader, ",")) {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "2026-09-15T12:00:00Z") || !strings.Contains(text, "test-gen") {
		t.Fatalf("csv row: %s", text)
	}
	if !strings.Contains(text, ",pro,") || !strings.Contains(text, ",offline,") {
		t.Fatalf("csv fields: %s", text)
	}

	jsonPath := filepath.Join(t.TempDir(), "arpu.json")
	if err := st.ExportToFile(jsonPath, ""); err != nil {
		t.Fatal(err)
	}
	jraw, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var entries []Entry
	if err := json.Unmarshal(jraw, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Command != "test-gen" || entries[0].PromptTokens != 0 {
		t.Fatalf("%+v", entries)
	}
}

func TestExportEmptyLedger(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteExport(&buf, nil, "csv"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(buf.String(), strings.Join(CSVHeader, ",")) {
		t.Fatalf("%s", buf.String())
	}
	buf.Reset()
	if err := WriteExport(&buf, nil, "json"); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Fatalf("empty json=%q", buf.String())
	}
}

func TestNormalizeFormat(t *testing.T) {
	got, err := NormalizeFormat("", "out.CSV")
	if err != nil || got != "csv" {
		t.Fatalf("csv ext: %s %v", got, err)
	}
	got, err = NormalizeFormat("", "out.json")
	if err != nil || got != "json" {
		t.Fatalf("json ext: %s %v", got, err)
	}
	if _, err := NormalizeFormat("xml", "out.xml"); err == nil {
		t.Fatal("xml should fail")
	}
}

type stubProvider struct {
	resp *provider.CompletionResponse
	err  error
}

func (s stubProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return s.resp, s.err
}

func TestCaptureWiresCompleteTokens(t *testing.T) {
	inner := stubProvider{resp: &provider.CompletionResponse{
		Content: "{}",
		Model:   "deepseek-chat",
		Usage:   provider.Usage{PromptTokens: 11, CompletionTokens: 7, TotalTokens: 18},
	}}
	cap := Wrap(inner)
	resp, err := cap.Complete(context.Background(), provider.CompletionRequest{Model: "ignored"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Usage.PromptTokens != 11 {
		t.Fatalf("resp tokens=%d", resp.Usage.PromptTokens)
	}
	if !cap.Called || cap.PromptTokens != 11 || cap.CompletionTokens != 7 || cap.Model != "deepseek-chat" {
		t.Fatalf("%+v", cap)
	}
}

func TestCaptureOfflineProviderNil(t *testing.T) {
	cap := Wrap(nil)
	if cap.Provider() != nil {
		t.Fatal("offline wrap should expose nil provider")
	}
	if cap.Called {
		t.Fatal("offline must not mark Complete called")
	}
}

func TestCaptureCompleteErrorStillCalled(t *testing.T) {
	inner := stubProvider{err: errors.New("boom")}
	cap := Wrap(inner)
	_, err := cap.Complete(context.Background(), provider.CompletionRequest{Model: "gpt-4o-mini"})
	if err == nil {
		t.Fatal("want error")
	}
	if !cap.Called {
		t.Fatal("failed Complete should still count as a request")
	}
	if cap.Model != "gpt-4o-mini" {
		t.Fatalf("model=%s", cap.Model)
	}
}

func TestDeterministicOfflineRow(t *testing.T) {
	st := Open(t.TempDir())
	if err := st.Append(Entry{
		Timestamp: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Tenant:    "local",
		SeatID:    "seat-local",
		Command:   "defect-blame",
		Status:    StatusOK,
	}); err != nil {
		t.Fatal(err)
	}
	got, err := st.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Provider != ProviderOffline || got[0].Model != ModelOffline {
		t.Fatalf("%+v", got[0])
	}
	if got[0].PromptTokens != 0 || got[0].CompletionTokens != 0 {
		t.Fatalf("offline tokens must be 0: %+v", got[0])
	}
}

func TestStatusOf(t *testing.T) {
	if StatusOf(nil) != StatusOK {
		t.Fatal("nil")
	}
	if StatusOf(errors.New("x")) != StatusError {
		t.Fatal("err")
	}
}
