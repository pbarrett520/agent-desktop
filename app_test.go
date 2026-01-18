package main

import (
	"context"
	"os"
	"testing"
	"time"

	"agent-desktop/internal/config"
	"agent-desktop/internal/conversation"
	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"
)

// MockLLMClient implements the Client interface for testing
type MockLLMClient struct {
	ChatCompletionFunc func(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error)
}

func (m *MockLLMClient) ChatCompletion(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error) {
	if m.ChatCompletionFunc != nil {
		return m.ChatCompletionFunc(ctx, messages, toolDefs)
	}
	return &llm.Response{Content: "Test response"}, nil
}

func setupTestApp(t *testing.T) (*App, func()) {
	tempDir, err := os.MkdirTemp("", "app_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	app := NewApp()
	app.ctx = context.Background()
	app.config = &config.Config{ExecutionTimeout: 60}

	// Create conversation store
	store, err := conversation.NewStore(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create mock client
	mockClient := &MockLLMClient{}
	app.convManager = conversation.NewManager(store, mockClient, "Test system prompt")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return app, cleanup
}

func TestApp_NewConversation(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	conv := app.NewConversation()

	if conv == nil {
		t.Fatal("Expected conversation to be created")
	}

	if conv.ID == "" {
		t.Error("Expected conversation to have ID")
	}

	// Should be the active conversation
	active := app.GetActiveConversation()
	if active == nil || active.ID != conv.ID {
		t.Error("New conversation should be active")
	}
}

func TestApp_ListConversations(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create some conversations
	app.NewConversation()
	time.Sleep(10 * time.Millisecond)
	app.NewConversation()

	summaries, err := app.ListConversations()
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}

	if len(summaries) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(summaries))
	}
}

func TestApp_LoadConversation(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Create and remember ID
	conv1 := app.NewConversation()
	conv1ID := conv1.ID

	// Create another (switches active)
	app.NewConversation()

	// Load the first
	loaded, err := app.LoadConversation(conv1ID)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if loaded.ID != conv1ID {
		t.Error("Loaded wrong conversation")
	}

	// Should be active
	active := app.GetActiveConversation()
	if active.ID != conv1ID {
		t.Error("Loaded conversation should be active")
	}
}

func TestApp_DeleteConversation(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	conv := app.NewConversation()
	convID := conv.ID

	err := app.DeleteConversation(convID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Should no longer exist
	_, err = app.LoadConversation(convID)
	if err == nil {
		t.Error("Deleted conversation should not be loadable")
	}
}

func TestApp_RenameConversation(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	conv := app.NewConversation()

	err := app.RenameConversation(conv.ID, "My New Title")
	if err != nil {
		t.Fatalf("Failed to rename: %v", err)
	}

	// Reload and verify
	loaded, _ := app.LoadConversation(conv.ID)
	if loaded.Title != "My New Title" {
		t.Errorf("Expected 'My New Title', got '%s'", loaded.Title)
	}
}

func TestApp_GetActiveConversation_ReturnsNilWhenNone(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// No conversation created yet
	// Need to reset the manager to have no active
	app.convManager = conversation.NewManager(
		app.convManager.GetStore(),
		&MockLLMClient{},
		"Test prompt",
	)

	active := app.GetActiveConversation()
	if active != nil {
		t.Error("Expected nil when no active conversation")
	}
}
