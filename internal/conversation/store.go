package conversation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Store handles persistence of conversations to disk.
type Store struct {
	basePath string
	mu       sync.RWMutex
}

// NewStore creates a new conversation store at the given path.
// It creates the directory and index file if they don't exist.
func NewStore(basePath string) (*Store, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	store := &Store{
		basePath: basePath,
	}

	// Initialize index file if it doesn't exist
	indexPath := filepath.Join(basePath, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		if err := store.writeIndex([]Summary{}); err != nil {
			return nil, fmt.Errorf("failed to create index file: %w", err)
		}
	}

	return store, nil
}

// Save persists a conversation to disk and updates the index.
func (s *Store) Save(conv *Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Write conversation file
	convPath := filepath.Join(s.basePath, fmt.Sprintf("conv_%s.json", conv.ID))
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	if err := os.WriteFile(convPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write conversation file: %w", err)
	}

	// Update index
	index, err := s.readIndex()
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	// Update or add summary in index
	summary := conv.ToSummary()
	found := false
	for i, existing := range index {
		if existing.ID == conv.ID {
			index[i] = summary
			found = true
			break
		}
	}
	if !found {
		index = append(index, summary)
	}

	// Sort by UpdatedAt descending (most recent first)
	sort.Slice(index, func(i, j int) bool {
		return index[i].UpdatedAt.After(index[j].UpdatedAt)
	})

	if err := s.writeIndex(index); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// Load retrieves a conversation by ID.
func (s *Store) Load(id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	convPath := filepath.Join(s.basePath, fmt.Sprintf("conv_%s.json", id))
	data, err := os.ReadFile(convPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read conversation file: %w", err)
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return &conv, nil
}

// List returns summaries of all conversations, sorted by most recent first.
func (s *Store) List() ([]Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.readIndex()
}

// Delete removes a conversation by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete conversation file
	convPath := filepath.Join(s.basePath, fmt.Sprintf("conv_%s.json", id))
	if err := os.Remove(convPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete conversation file: %w", err)
	}

	// Update index
	index, err := s.readIndex()
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	// Remove from index
	newIndex := make([]Summary, 0, len(index))
	for _, summary := range index {
		if summary.ID != id {
			newIndex = append(newIndex, summary)
		}
	}

	if err := s.writeIndex(newIndex); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// readIndex reads the index file (caller must hold lock).
func (s *Store) readIndex() ([]Summary, error) {
	indexPath := filepath.Join(s.basePath, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var index []Summary
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return index, nil
}

// writeIndex writes the index file (caller must hold lock).
func (s *Store) writeIndex(index []Summary) error {
	indexPath := filepath.Join(s.basePath, "index.json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}

// GetDefaultStorePath returns the default path for conversation storage.
func GetDefaultStorePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".agent-desktop", "conversations"), nil
}
