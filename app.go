package main

import (
	"context"

	"agent-desktop/internal/agent"
	"agent-desktop/internal/config"
	"agent-desktop/internal/conversation"
	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config *config.Config
	client *llm.Client

	// Conversation state
	convManager *conversation.Manager

	// Agent state
	agentCancel context.CancelFunc
	agentCtx    context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{ExecutionTimeout: 60}
	}
	a.config = cfg

	// Initialize LLM client if configured
	if cfg.IsConfigured() {
		client, err := llm.NewClient(cfg)
		if err == nil {
			a.client = client
		}
	}

	// Initialize conversation manager
	a.initConversationManager()
}

// initConversationManager initializes or reinitializes the conversation manager.
func (a *App) initConversationManager() {
	storePath, err := conversation.GetDefaultStorePath()
	if err != nil {
		// Fallback to temp directory if home dir fails
		storePath = "./conversations"
	}

	store, err := conversation.NewStore(storePath)
	if err != nil {
		// Log error but don't fail startup
		return
	}

	systemPrompt := agent.GetSystemPrompt()
	a.convManager = conversation.NewManager(store, a.client, systemPrompt)
}

// ============================================================================
// Configuration Methods
// ============================================================================

// GetConfig returns the current configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(cfg *config.Config) error {
	if err := cfg.Save(); err != nil {
		return err
	}
	a.config = cfg

	// Reinitialize client with new config
	if cfg.IsConfigured() {
		client, err := llm.NewClient(cfg)
		if err == nil {
			a.client = client
		}
	}

	return nil
}

// IsConfigured returns true if the app is configured with LLM credentials
func (a *App) IsConfigured() bool {
	return a.config != nil && a.config.IsConfigured()
}

// TestConnection tests the LLM connection
func (a *App) TestConnection() (bool, string) {
	if a.config == nil {
		return false, "No configuration loaded"
	}
	return llm.TestConnection(a.config)
}

// ============================================================================
// Session Methods
// ============================================================================

// GetSessionInfo returns information about the current shell session
func (a *App) GetSessionInfo() map[string]interface{} {
	return tools.GetSessionInfo()
}

// ResetSession resets the shell session
func (a *App) ResetSession() {
	tools.ResetSession()
}

// ============================================================================
// Conversation Methods
// ============================================================================

// NewConversation creates a new conversation and makes it active.
func (a *App) NewConversation() *conversation.Conversation {
	if a.convManager == nil {
		return nil
	}
	return a.convManager.New()
}

// LoadConversation loads an existing conversation by ID.
func (a *App) LoadConversation(id string) (*conversation.Conversation, error) {
	if a.convManager == nil {
		return nil, nil
	}
	return a.convManager.Load(id)
}

// ListConversations returns summaries of all saved conversations.
func (a *App) ListConversations() ([]conversation.Summary, error) {
	if a.convManager == nil {
		return nil, nil
	}
	return a.convManager.List()
}

// DeleteConversation removes a conversation by ID.
func (a *App) DeleteConversation(id string) error {
	if a.convManager == nil {
		return nil
	}
	return a.convManager.Delete(id)
}

// RenameConversation sets a custom title for a conversation.
func (a *App) RenameConversation(id string, title string) error {
	if a.convManager == nil {
		return nil
	}

	// Load the conversation if it's not active
	active := a.convManager.GetActive()
	if active == nil || active.ID != id {
		_, err := a.convManager.Load(id)
		if err != nil {
			return err
		}
	}

	return a.convManager.Rename(title)
}

// GetActiveConversation returns the currently active conversation.
func (a *App) GetActiveConversation() *conversation.Conversation {
	if a.convManager == nil {
		return nil
	}
	return a.convManager.GetActive()
}

// SendMessage sends a message to the active conversation and runs the agent.
// This is the main method for multi-turn chat.
func (a *App) SendMessage(message string, taskContext string) {
	if a.client == nil {
		runtime.EventsEmit(a.ctx, "agent:error", "LLM not configured")
		return
	}

	if a.convManager == nil {
		runtime.EventsEmit(a.ctx, "agent:error", "Conversation manager not initialized")
		return
	}

	// Ensure we have an active conversation
	if a.convManager.GetActive() == nil {
		a.convManager.New()
	}

	// Cancel any existing agent run
	if a.agentCancel != nil {
		a.agentCancel()
	}

	// Create new context for this run
	a.agentCtx, a.agentCancel = context.WithCancel(context.Background())

	go func() {
		// Build message content with optional context
		content := message
		if taskContext != "" {
			content = message + "\n\nContext: " + taskContext
		}

		// Add user message to conversation
		if err := a.convManager.AddUserMessage(content); err != nil {
			runtime.EventsEmit(a.ctx, "agent:error", "Failed to add message: "+err.Error())
			return
		}

		// Get messages for the agent
		messages := a.convManager.GetMessages()

		maxSteps := 20
		if a.config.ExecutionTimeout > 0 {
			maxSteps = a.config.ExecutionTimeout / 3
			if maxSteps < 10 {
				maxSteps = 10
			}
			if maxSteps > 50 {
				maxSteps = 50
			}
		}

		// Run conversation continuation
		for step := range agent.ContinueConversation(a.agentCtx, a.client, messages, maxSteps) {
			// Emit step to frontend
			runtime.EventsEmit(a.ctx, "agent:step", step)

			// Update conversation with new messages if present
			if step.Messages != nil {
				// Find and add new messages since last sync
				currentMsgs := a.convManager.GetMessages()
				for i := len(currentMsgs); i < len(step.Messages); i++ {
					msg := step.Messages[i]
					if msg.Role == "assistant" {
						a.convManager.AddAssistantMessage(msg)
					} else if msg.Role == "tool" {
						a.convManager.AddToolMessage(msg.ToolCallID, msg.Content)
					}
				}
			}

			// Handle completion states
			if step.Type == agent.StepTypeComplete {
				// Generate title if this is the first completion
				go a.convManager.GenerateTitle(context.Background())
				runtime.EventsEmit(a.ctx, "agent:complete", step.Content)
				return
			}
			if step.Type == agent.StepTypeAssistantMessage {
				// Conversational response - also triggers title generation
				go a.convManager.GenerateTitle(context.Background())
				runtime.EventsEmit(a.ctx, "agent:message", step.Content)
				return
			}
			if step.Type == agent.StepTypeError {
				runtime.EventsEmit(a.ctx, "agent:error", step.Content)
				return
			}
		}
	}()
}

// ============================================================================
// Agent Methods (Legacy - kept for backward compatibility)
// ============================================================================

// RunAgentTask starts the agent to complete a task
// It emits events to the frontend as the agent progresses
func (a *App) RunAgentTask(task string, taskContext string) {
	if a.client == nil {
		runtime.EventsEmit(a.ctx, "agent:error", "LLM not configured")
		return
	}

	// Cancel any existing agent run
	if a.agentCancel != nil {
		a.agentCancel()
	}

	// Create new context for this run
	a.agentCtx, a.agentCancel = context.WithCancel(context.Background())

	go func() {
		// Reset session for fresh start
		tools.ResetSession()

		maxSteps := 20
		if a.config.ExecutionTimeout > 0 {
			// Use execution timeout as rough guide for max steps
			maxSteps = a.config.ExecutionTimeout / 3
			if maxSteps < 10 {
				maxSteps = 10
			}
			if maxSteps > 50 {
				maxSteps = 50
			}
		}

		for step := range agent.RunLoop(a.agentCtx, a.client, task, taskContext, maxSteps) {
			// Emit step to frontend
			runtime.EventsEmit(a.ctx, "agent:step", step)

			// Check if complete or error
			if step.Type == agent.StepTypeComplete {
				runtime.EventsEmit(a.ctx, "agent:complete", step.Content)
				return
			}
			if step.Type == agent.StepTypeError {
				runtime.EventsEmit(a.ctx, "agent:error", step.Content)
				return
			}
		}
	}()
}

// StopAgent stops the currently running agent
func (a *App) StopAgent() {
	if a.agentCancel != nil {
		a.agentCancel()
		a.agentCancel = nil
	}
}
