// Package audit provides an append-only record of every az command the
// agent executes or is denied, so a consultant can show a client (or their
// own compliance team) exactly what the agent did and who approved it.
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Event is one row in the audit trail.
type Event struct {
	Timestamp      string `json:"timestamp"` // RFC3339
	SessionID      string `json:"session_id"`
	SubscriptionID string `json:"subscription_id,omitempty"`
	Tier           string `json:"tier"`
	Command        string `json:"command"`
	ProposedBy     string `json:"proposed_by"` // "agent"
	Decision       string `json:"decision"`    // auto | approved:<user> | denied:<user>
	ExitCode       int    `json:"exit_code"`
	DurationMs     int64  `json:"duration_ms"`
	OutputHash     string `json:"output_hash"` // sha256 of (possibly truncated) output, hex
}

// Sink is where audit events go. JSONLSink is the only implementation today;
// the interface exists so a future remote observability stream can be added
// without touching any call site that writes an Event.
type Sink interface {
	Write(Event) error
}

// JSONLSink appends one JSON object per line to a file.
type JSONLSink struct {
	path string
	mu   sync.Mutex
}

// NewJSONLSink creates a sink writing to the given path, creating parent
// directories as needed.
func NewJSONLSink(path string) (*JSONLSink, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return &JSONLSink{path: path}, nil
}

// Write appends one event as a JSON line.
func (s *JSONLSink) Write(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// ReadAll reads every event currently in the JSONL file, oldest first.
// Missing file is not an error — it returns an empty slice.
func ReadAll(path string) ([]Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, err
	}

	var events []Event
	dec := json.NewDecoder(strings.NewReader(string(data)))
	for {
		var e Event
		if err := dec.Decode(&e); err != nil {
			break
		}
		events = append(events, e)
	}
	return events, nil
}

var (
	defaultSink   Sink
	defaultSinkMu sync.Mutex
)

// SetDefaultSink overrides the process-wide default sink. Useful for tests.
func SetDefaultSink(s Sink) {
	defaultSinkMu.Lock()
	defer defaultSinkMu.Unlock()
	defaultSink = s
}

// DefaultPath returns ~/.agent_desktop/audit.jsonl.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agent_desktop", "audit.jsonl"), nil
}

// Record writes an event to the default sink, initializing it from
// DefaultPath on first use.
func Record(e Event) error {
	defaultSinkMu.Lock()
	sink := defaultSink
	defaultSinkMu.Unlock()

	if sink == nil {
		path, err := DefaultPath()
		if err != nil {
			return err
		}
		s, err := NewJSONLSink(path)
		if err != nil {
			return err
		}
		SetDefaultSink(s)
		sink = s
	}

	return sink.Write(e)
}

// HashOutput returns a truncated sha256 hex digest of output, capping the
// input considered to the first 8000 bytes so very large tool outputs don't
// slow down audit writes.
func HashOutput(output string) string {
	const maxBytes = 8000
	if len(output) > maxBytes {
		output = output[:maxBytes]
	}
	sum := sha256.Sum256([]byte(output))
	return hex.EncodeToString(sum[:])[:16]
}
