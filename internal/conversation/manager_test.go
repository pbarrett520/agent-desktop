package conversation

import (
	"context"
	"os"
	"testing"

	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"
)

// MockClient implements the Client interface for testing
type MockClient struct {
	ChatCompletionFunc func(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error)
}

func (m *MockClient) ChatCompletion(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error) {
	if m.ChatCompletionFunc != nil {
		return m.ChatCompletionFunc(ctx, messages, toolDefs)
	}
	return &llm.Response{Content: "Test Title"}, nil
}

func setupTestManager(t *testing.T) (*Manager, func()) {
	tempDir, err := os.MkdirTemp("", "manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	store, err := NewStore(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	mockClient := &MockClient{}
	manager := NewManager(store, mockClient, "You are a helpful assistant.")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return manager, cleanup
}

func TestNewManager(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	if manager == nil {
		t.Error("Expected manager to be created")
	}

	if manager.GetActive() != nil {
		t.Error("Expected no active conversation initially")
	}
}

func TestManagerNewConversation(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	conv := manager.New()

	if conv == nil {
		t.Fatal("Expected conversation to be created")
	}

	if conv.ID == "" {
		t.Error("Expected conversation to have ID")
	}

	// Should have system prompt as first message
	if len(conv.Messages) != 1 {
		t.Errorf("Expected 1 message (system prompt), got %d", len(conv.Messages))
	}

	if conv.Messages[0].Role != "system" {
		t.Errorf("Expected first message to be system, got '%s'", conv.Messages[0].Role)
	}

	// Should be the active conversation
	if manager.GetActive() != conv {
		t.Error("Expected new conversation to be active")
	}
}

func TestManagerAddUserMessage(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	manager.New()
	err := manager.AddUserMessage("Hello, world!")

	if err != nil {
		t.Fatalf("Failed to add user message: %v", err)
	}

	active := manager.GetActive()
	if len(active.Messages) != 2 { // system + user
		t.Errorf("Expected 2 messages, got %d", len(active.Messages))
	}

	if active.Messages[1].Role != "user" {
		t.Errorf("Expected user role, got '%s'", active.Messages[1].Role)
	}

	if active.Messages[1].Content != "Hello, world!" {
		t.Errorf("Expected 'Hello, world!', got '%s'", active.Messages[1].Content)
	}
}

func TestManagerAddUserMessageWithoutActiveConversation(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	err := manager.AddUserMessage("Hello")
	if err == nil {
		t.Error("Expected error when adding message without active conversation")
	}
}

func TestManagerAddAssistantMessage(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	manager.New()
	manager.AddUserMessage("Hello")

	msg := llm.Message{
		Role:    "assistant",
		Content: "Hi there! How can I help you?",
	}
	err := manager.AddAssistantMessage(msg)

	if err != nil {
		t.Fatalf("Failed to add assistant message: %v", err)
	}

	active := manager.GetActive()
	if len(active.Messages) != 3 { // system + user + assistant
		t.Errorf("Expected 3 messages, got %d", len(active.Messages))
	}

	if active.Messages[2].Content != "Hi there! How can I help you?" {
		t.Errorf("Unexpected assistant message content")
	}
}

func TestManagerGetMessages(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	manager.New()
	manager.AddUserMessage("Hello")
	manager.AddAssistantMessage(llm.Message{Role: "assistant", Content: "Hi!"})

	messages := manager.GetMessages()

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}

	// Should be a copy, not the original
	messages[0].Content = "Modified"
	if manager.GetActive().Messages[0].Content == "Modified" {
		t.Error("GetMessages should return a copy")
	}
}

func TestManagerAutoSave(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	conv := manager.New()
	manager.AddUserMessage("Hello")

	// Load from store to verify it was saved
	loaded, err := manager.store.Load(conv.ID)
	if err != nil {
		t.Fatalf("Failed to load saved conversation: %v", err)
	}

	if len(loaded.Messages) != 2 { // system + user
		t.Errorf("Expected saved conversation to have 2 messages, got %d", len(loaded.Messages))
	}
}

func TestManagerLoadConversation(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	// Create and save a conversation
	conv := manager.New()
	convID := conv.ID
	manager.AddUserMessage("Hello")
	manager.AddAssistantMessage(llm.Message{Role: "assistant", Content: "Hi!"})

	// Create a new conversation (switches active)
	manager.New()

	// Load the original
	loaded, err := manager.Load(convID)
	if err != nil {
		t.Fatalf("Failed to load conversation: %v", err)
	}

	if loaded.ID != convID {
		t.Errorf("Expected ID '%s', got '%s'", convID, loaded.ID)
	}

	if manager.GetActive().ID != convID {
		t.Error("Expected loaded conversation to be active")
	}

	if len(loaded.Messages) != 3 { // system + user + assistant
		t.Errorf("Expected 3 messages, got %d", len(loaded.Messages))
	}
}

func TestManagerLoadNonExistent(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	_, err := manager.Load("nonexistent")
	if err == nil {
		t.Error("Expected error loading non-existent conversation")
	}
}

func TestManagerRename(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	manager.New()
	err := manager.Rename("My Custom Title")

	if err != nil {
		t.Fatalf("Failed to rename: %v", err)
	}

	if manager.GetActive().Title != "My Custom Title" {
		t.Errorf("Expected title 'My Custom Title', got '%s'", manager.GetActive().Title)
	}

	// Verify it was saved
	loaded, _ := manager.store.Load(manager.GetActive().ID)
	if loaded.Title != "My Custom Title" {
		t.Error("Expected renamed title to be saved")
	}
}

func TestManagerRenameWithoutActiveConversation(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	err := manager.Rename("Title")
	if err == nil {
		t.Error("Expected error when renaming without active conversation")
	}
}

func TestManagerListConversations(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	// Create multiple conversations
	manager.New()
	manager.AddUserMessage("First")

	manager.New()
	manager.AddUserMessage("Second")

	summaries, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}

	if len(summaries) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(summaries))
	}
}

func TestManagerDelete(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	conv := manager.New()
	convID := conv.ID
	manager.AddUserMessage("Hello")

	err := manager.Delete(convID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Active should be nil after deleting active conversation
	if manager.GetActive() != nil {
		t.Error("Expected active conversation to be nil after deleting it")
	}

	// Should not be in list
	summaries, _ := manager.List()
	for _, s := range summaries {
		if s.ID == convID {
			t.Error("Deleted conversation should not appear in list")
		}
	}
}

func TestManagerDeleteNonActive(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	// Create first conversation
	conv1 := manager.New()
	conv1ID := conv1.ID

	// Create second conversation (now active)
	conv2 := manager.New()

	// Delete the first (non-active)
	err := manager.Delete(conv1ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Active should still be conv2
	if manager.GetActive().ID != conv2.ID {
		t.Error("Active conversation should not change when deleting non-active")
	}
}

func TestManagerGenerateTitle(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	// Setup mock client to return a specific title
	mockClient := &MockClient{
		ChatCompletionFunc: func(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error) {
			return &llm.Response{Content: "Greeting Conversation"}, nil
		},
	}
	manager.client = mockClient

	manager.New()
	manager.AddUserMessage("Hello!")

	err := manager.GenerateTitle(context.Background())
	if err != nil {
		t.Fatalf("Failed to generate title: %v", err)
	}

	if manager.GetActive().Title != "Greeting Conversation" {
		t.Errorf("Expected title 'Greeting Conversation', got '%s'", manager.GetActive().Title)
	}
}

func TestManagerGenerateTitleSkipsIfAlreadySet(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	callCount := 0
	mockClient := &MockClient{
		ChatCompletionFunc: func(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error) {
			callCount++
			return &llm.Response{Content: "New Title"}, nil
		},
	}
	manager.client = mockClient

	manager.New()
	manager.Rename("Custom Title")
	manager.AddUserMessage("Hello!")

	err := manager.GenerateTitle(context.Background())
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Should not have called LLM
	if callCount > 0 {
		t.Error("Should not call LLM when title is already set")
	}

	// Title should remain unchanged
	if manager.GetActive().Title != "Custom Title" {
		t.Errorf("Title should remain 'Custom Title', got '%s'", manager.GetActive().Title)
	}
}
