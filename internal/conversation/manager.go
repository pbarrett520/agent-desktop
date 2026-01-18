package conversation

import (
	"context"
	"errors"
	"strings"

	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"
)

// Client interface for LLM calls (allows mocking in tests)
type Client interface {
	ChatCompletion(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error)
}

// Manager handles active conversation state and operations.
type Manager struct {
	store        *Store
	client       Client
	active       *Conversation
	systemPrompt string
}

// NewManager creates a new conversation manager.
func NewManager(store *Store, client Client, systemPrompt string) *Manager {
	return &Manager{
		store:        store,
		client:       client,
		systemPrompt: systemPrompt,
	}
}

// New creates a new conversation, resets the tools session, and makes it active.
func (m *Manager) New() *Conversation {
	// Reset tools session for new conversation
	tools.ResetSession()

	conv := New()

	// Add system prompt as first message
	conv.AddMessage(llm.Message{
		Role:    "system",
		Content: m.systemPrompt,
	})

	m.active = conv

	// Auto-save
	m.store.Save(conv)

	return conv
}

// Load retrieves a conversation by ID, resets the tools session, and makes it active.
func (m *Manager) Load(id string) (*Conversation, error) {
	conv, err := m.store.Load(id)
	if err != nil {
		return nil, err
	}

	// Reset tools session when loading a different conversation
	tools.ResetSession()

	m.active = conv
	return conv, nil
}

// GetActive returns the currently active conversation, or nil if none.
func (m *Manager) GetActive() *Conversation {
	return m.active
}

// AddUserMessage adds a user message to the active conversation and auto-saves.
func (m *Manager) AddUserMessage(content string) error {
	if m.active == nil {
		return errors.New("no active conversation")
	}

	m.active.AddMessage(llm.Message{
		Role:    "user",
		Content: content,
	})

	return m.store.Save(m.active)
}

// AddAssistantMessage adds an assistant message to the active conversation and auto-saves.
func (m *Manager) AddAssistantMessage(msg llm.Message) error {
	if m.active == nil {
		return errors.New("no active conversation")
	}

	m.active.AddMessage(msg)
	return m.store.Save(m.active)
}

// AddToolMessage adds a tool result message to the active conversation and auto-saves.
func (m *Manager) AddToolMessage(toolCallID, content string) error {
	if m.active == nil {
		return errors.New("no active conversation")
	}

	m.active.AddMessage(llm.Message{
		Role:       "tool",
		Content:    content,
		ToolCallID: toolCallID,
	})

	return m.store.Save(m.active)
}

// GetMessages returns a copy of the current conversation messages.
// This is safe to pass to the agent loop without risking mutation.
func (m *Manager) GetMessages() []llm.Message {
	if m.active == nil {
		return nil
	}

	// Return a copy
	messages := make([]llm.Message, len(m.active.Messages))
	copy(messages, m.active.Messages)
	return messages
}

// Rename sets a custom title for the active conversation and saves.
func (m *Manager) Rename(title string) error {
	if m.active == nil {
		return errors.New("no active conversation")
	}

	m.active.Title = title
	m.active.UpdatedAt = m.active.UpdatedAt // Keep the same timestamp for rename
	return m.store.Save(m.active)
}

// List returns summaries of all conversations.
func (m *Manager) List() ([]Summary, error) {
	return m.store.List()
}

// Delete removes a conversation by ID.
// If deleting the active conversation, active is set to nil.
func (m *Manager) Delete(id string) error {
	err := m.store.Delete(id)
	if err != nil {
		return err
	}

	// If we deleted the active conversation, clear it
	if m.active != nil && m.active.ID == id {
		m.active = nil
	}

	return nil
}

// GenerateTitle uses the LLM to generate a title based on the first user message.
// If the conversation already has a non-default title, this is a no-op.
func (m *Manager) GenerateTitle(ctx context.Context) error {
	if m.active == nil {
		return errors.New("no active conversation")
	}

	// Skip if no LLM client configured
	if m.client == nil {
		return nil
	}

	// Skip if title is already set (not default)
	if m.active.Title != "" && m.active.Title != "New Conversation" {
		return nil
	}

	// Find first user message
	var firstUserMessage string
	for _, msg := range m.active.Messages {
		if msg.Role == "user" {
			firstUserMessage = msg.Content
			break
		}
	}

	if firstUserMessage == "" {
		return nil // No user message yet
	}

	// Call LLM to generate title
	prompt := []llm.Message{
		{
			Role:    "system",
			Content: "Generate a short title (3-6 words) for this conversation based on the user's first message. Reply with only the title, no quotes or extra text.",
		},
		{
			Role:    "user",
			Content: firstUserMessage,
		},
	}

	resp, err := m.client.ChatCompletion(ctx, prompt, nil)
	if err != nil {
		return err
	}

	// Clean up the title
	title := strings.TrimSpace(resp.Content)
	title = strings.Trim(title, "\"'") // Remove quotes if present

	m.active.Title = title
	return m.store.Save(m.active)
}

// Save explicitly saves the active conversation.
func (m *Manager) Save() error {
	if m.active == nil {
		return errors.New("no active conversation")
	}
	return m.store.Save(m.active)
}

// GetStore returns the underlying store (for testing purposes).
func (m *Manager) GetStore() *Store {
	return m.store
}
