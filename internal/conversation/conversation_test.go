package conversation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"agent-desktop/internal/llm"
)

func TestNewConversation(t *testing.T) {
	conv := New()

	if conv.ID == "" {
		t.Error("Expected conversation to have an ID")
	}

	if conv.Title != "New Conversation" {
		t.Errorf("Expected default title 'New Conversation', got '%s'", conv.Title)
	}

	if conv.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if conv.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	if len(conv.Messages) != 0 {
		t.Errorf("Expected empty messages, got %d", len(conv.Messages))
	}
}

func TestConversationAddMessage(t *testing.T) {
	conv := New()
	originalUpdatedAt := conv.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	msg := llm.Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	conv.AddMessage(msg)

	if len(conv.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(conv.Messages))
	}

	if conv.Messages[0].Content != "Hello, world!" {
		t.Errorf("Expected message content 'Hello, world!', got '%s'", conv.Messages[0].Content)
	}

	if !conv.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after adding message")
	}
}

func TestConversationTurnCount(t *testing.T) {
	conv := New()

	// Add system message (shouldn't count as turn)
	conv.AddMessage(llm.Message{Role: "system", Content: "You are helpful"})

	// Add user message (turn 1)
	conv.AddMessage(llm.Message{Role: "user", Content: "Hi"})

	// Add assistant message
	conv.AddMessage(llm.Message{Role: "assistant", Content: "Hello!"})

	// Add user message (turn 2)
	conv.AddMessage(llm.Message{Role: "user", Content: "How are you?"})

	// Turn count = number of user messages
	if conv.TurnCount() != 2 {
		t.Errorf("Expected turn count 2, got %d", conv.TurnCount())
	}
}

func TestConversationToSummary(t *testing.T) {
	conv := New()
	conv.Title = "Test Conversation"
	conv.AddMessage(llm.Message{Role: "user", Content: "Hello"})
	conv.AddMessage(llm.Message{Role: "assistant", Content: "Hi there!"})

	summary := conv.ToSummary()

	if summary.ID != conv.ID {
		t.Errorf("Expected summary ID '%s', got '%s'", conv.ID, summary.ID)
	}

	if summary.Title != conv.Title {
		t.Errorf("Expected summary title '%s', got '%s'", conv.Title, summary.Title)
	}

	if summary.TurnCount != 1 {
		t.Errorf("Expected turn count 1, got %d", summary.TurnCount)
	}
}

// Store tests

func setupTestStore(t *testing.T) (*Store, func()) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "conversation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	store, err := NewStore(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return store, cleanup
}

func TestNewStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store == nil {
		t.Error("Expected store to be created")
	}

	// Check that index file was created
	indexPath := filepath.Join(store.basePath, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("Expected index.json to be created")
	}
}

func TestStoreSaveAndLoad(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create a conversation
	conv := New()
	conv.Title = "Test Save Load"
	conv.AddMessage(llm.Message{Role: "user", Content: "Hello"})
	conv.AddMessage(llm.Message{Role: "assistant", Content: "Hi!"})

	// Save it
	err := store.Save(conv)
	if err != nil {
		t.Fatalf("Failed to save conversation: %v", err)
	}

	// Load it back
	loaded, err := store.Load(conv.ID)
	if err != nil {
		t.Fatalf("Failed to load conversation: %v", err)
	}

	if loaded.ID != conv.ID {
		t.Errorf("Expected ID '%s', got '%s'", conv.ID, loaded.ID)
	}

	if loaded.Title != conv.Title {
		t.Errorf("Expected title '%s', got '%s'", conv.Title, loaded.Title)
	}

	if len(loaded.Messages) != len(conv.Messages) {
		t.Errorf("Expected %d messages, got %d", len(conv.Messages), len(loaded.Messages))
	}
}

func TestStoreList(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create and save multiple conversations
	conv1 := New()
	conv1.Title = "First Conversation"
	conv1.AddMessage(llm.Message{Role: "user", Content: "Hello"})
	store.Save(conv1)

	// Small delay to ensure different timestamps
	time.Sleep(20 * time.Millisecond)

	conv2 := New()
	conv2.Title = "Second Conversation"
	conv2.AddMessage(llm.Message{Role: "user", Content: "Hi"})
	conv2.AddMessage(llm.Message{Role: "user", Content: "How are you?"})
	store.Save(conv2)

	// List conversations
	summaries, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list conversations: %v", err)
	}

	if len(summaries) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(summaries))
	}

	// Should be sorted by UpdatedAt descending (most recent first)
	// conv2 was saved last, so it should be first
	if summaries[0].Title != "Second Conversation" {
		t.Errorf("Expected most recent conversation first, got '%s'", summaries[0].Title)
	}
}

func TestStoreDelete(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create and save a conversation
	conv := New()
	conv.Title = "To Be Deleted"
	store.Save(conv)

	// Verify it exists
	_, err := store.Load(conv.ID)
	if err != nil {
		t.Fatalf("Conversation should exist before delete: %v", err)
	}

	// Delete it
	err = store.Delete(conv.ID)
	if err != nil {
		t.Fatalf("Failed to delete conversation: %v", err)
	}

	// Verify it's gone
	_, err = store.Load(conv.ID)
	if err == nil {
		t.Error("Expected error loading deleted conversation")
	}

	// Verify it's not in the list
	summaries, _ := store.List()
	for _, s := range summaries {
		if s.ID == conv.ID {
			t.Error("Deleted conversation should not appear in list")
		}
	}
}

func TestStoreLoadNonExistent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	_, err := store.Load("nonexistent-id")
	if err == nil {
		t.Error("Expected error loading non-existent conversation")
	}
}

func TestStoreUpdateExisting(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Create and save
	conv := New()
	conv.Title = "Original Title"
	store.Save(conv)

	// Modify and save again
	conv.Title = "Updated Title"
	conv.AddMessage(llm.Message{Role: "user", Content: "New message"})
	store.Save(conv)

	// Load and verify
	loaded, err := store.Load(conv.ID)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if loaded.Title != "Updated Title" {
		t.Errorf("Expected updated title, got '%s'", loaded.Title)
	}

	if len(loaded.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(loaded.Messages))
	}

	// Verify list still has only 1 entry
	summaries, _ := store.List()
	if len(summaries) != 1 {
		t.Errorf("Expected 1 conversation in list, got %d", len(summaries))
	}
}
