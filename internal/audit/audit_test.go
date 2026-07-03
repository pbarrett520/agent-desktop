package audit

import (
	"path/filepath"
	"testing"
)

func TestJSONLSink_WriteAndReadAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	sink, err := NewJSONLSink(path)
	if err != nil {
		t.Fatalf("NewJSONLSink: %v", err)
	}

	events := []Event{
		{Timestamp: "2026-07-03T10:00:00Z", SessionID: "s1", Tier: "READ", Command: "az vm list", ProposedBy: "agent", Decision: "auto", ExitCode: 0, DurationMs: 120, OutputHash: "abc123"},
		{Timestamp: "2026-07-03T10:01:00Z", SessionID: "s1", Tier: "DESTRUCTIVE", Command: "az group delete --name rg-old", ProposedBy: "agent", Decision: "denied:patrick", ExitCode: 0, DurationMs: 0, OutputHash: ""},
	}
	for _, e := range events {
		if err := sink.Write(e); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	got, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != len(events) {
		t.Fatalf("ReadAll returned %d events, want %d", len(got), len(events))
	}
	if got[0].Command != events[0].Command || got[1].Decision != events[1].Decision {
		t.Errorf("ReadAll returned mismatched events: %+v", got)
	}
}

func TestReadAll_MissingFileReturnsEmpty(t *testing.T) {
	got, err := ReadAll(filepath.Join(t.TempDir(), "does-not-exist.jsonl"))
	if err != nil {
		t.Fatalf("ReadAll on missing file should not error, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d events", len(got))
	}
}

func TestHashOutput_Deterministic(t *testing.T) {
	h1 := HashOutput("some tool output")
	h2 := HashOutput("some tool output")
	if h1 != h2 {
		t.Errorf("HashOutput should be deterministic: %q != %q", h1, h2)
	}
	if HashOutput("different output") == h1 {
		t.Errorf("HashOutput should differ for different input")
	}
	if len(h1) != 16 {
		t.Errorf("expected 16-char truncated hash, got %d chars", len(h1))
	}
}

func TestRecord_UsesDefaultSink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	sink, err := NewJSONLSink(path)
	if err != nil {
		t.Fatalf("NewJSONLSink: %v", err)
	}
	SetDefaultSink(sink)
	defer SetDefaultSink(nil)

	if err := Record(Event{Command: "az vm list", Tier: "READ", Decision: "auto"}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	got, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 1 || got[0].Command != "az vm list" {
		t.Errorf("Record did not write through default sink, got: %+v", got)
	}
}
