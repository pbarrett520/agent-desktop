package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/user"
	"sync"
	"time"

	"agent-desktop/internal/agent"
	"agent-desktop/internal/audit"
	"agent-desktop/internal/azure"
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

	// sessionID identifies this app run in the audit trail.
	sessionID string

	// pendingProposals holds az_propose proposals awaiting a user decision,
	// keyed by Proposal.ID.
	pendingProposalsMu sync.Mutex
	pendingProposals   map[string]tools.Proposal
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		pendingProposals: make(map[string]tools.Proposal),
	}
}

// newSessionID generates a short random identifier for the audit trail.
func newSessionID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return "session-" + hex.EncodeToString(buf)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.sessionID = newSessionID()
	tools.SetSessionID(a.sessionID)

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

	mode := config.ModeCloudOps
	if a.config != nil && a.config.Mode != "" {
		mode = a.config.Mode
	}
	systemPrompt := agent.GetSystemPrompt(mode)
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
			// Reinitialize conversation manager with the new client
			a.initConversationManager()
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

// FetchModels fetches available models from an OpenAI-compatible endpoint.
// This works without a fully configured client — only endpoint and optional
// API key are needed. Returns chat models only (embedding models filtered out).
func (a *App) FetchModels(endpoint string, apiKey string) ([]llm.ModelInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return llm.FetchModels(ctx, endpoint, apiKey)
}

// CheckEndpointHealth checks if an endpoint is reachable without requiring
// a model to be configured. Useful for showing connection status for LM Studio.
func (a *App) CheckEndpointHealth(endpoint string, apiKey string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return llm.CheckEndpointHealth(ctx, endpoint, apiKey)
}

// ============================================================================
// Azure Context Methods
// ============================================================================

// GetAzureContext detects az CLI installation, login state, and the active
// tenant/subscription/user for the consultant's own authenticated session.
func (a *App) GetAzureContext() *azure.Context {
	return azure.GetAzureContext()
}

// ListSubscriptions returns all subscriptions visible to the current az login.
func (a *App) ListSubscriptions() ([]azure.Subscription, error) {
	return azure.ListSubscriptions()
}

// SetSubscription switches the active subscription for the local az session.
func (a *App) SetSubscription(id string) error {
	return azure.SetSubscription(id)
}

// ListResourceGroups returns resource groups and resource counts for the
// subscription overview dashboard.
func (a *App) ListResourceGroups() ([]azure.ResourceGroupSummary, error) {
	return azure.ListResourceGroups()
}

// ListVMPowerStates returns VM power states for the subscription overview dashboard.
func (a *App) ListVMPowerStates() ([]azure.VMPowerState, error) {
	return azure.ListVMPowerStates()
}

// GetMonthlyCost returns this month's cost to date for subscriptionID, or an
// error if the costmanagement extension isn't available — the frontend
// degrades gracefully in that case.
func (a *App) GetMonthlyCost(subscriptionID string) (*azure.CostSummary, error) {
	return azure.GetMonthlyCost(subscriptionID)
}

// ============================================================================
// Audit Methods
// ============================================================================

// GetAuditLog returns every recorded audit event, oldest first. The
// frontend Audit panel sorts/filters client-side.
func (a *App) GetAuditLog() ([]audit.Event, error) {
	path, err := audit.DefaultPath()
	if err != nil {
		return nil, err
	}
	return audit.ReadAll(path)
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

// currentMode returns the configured agent mode, defaulting to cloud ops.
func (a *App) currentMode() string {
	if a.config != nil && a.config.Mode != "" {
		return a.config.Mode
	}
	return config.ModeCloudOps
}

// maxStepsFromConfig derives a step budget from the configured execution timeout.
func (a *App) maxStepsFromConfig() int {
	maxSteps := 20
	if a.config != nil && a.config.ExecutionTimeout > 0 {
		maxSteps = a.config.ExecutionTimeout / 3
		if maxSteps < 10 {
			maxSteps = 10
		}
		if maxSteps > 50 {
			maxSteps = 50
		}
	}
	return maxSteps
}

// runContinuation drives a ContinueConversation channel to completion,
// syncing conversation state and emitting frontend events. Shared by
// SendMessage and ResolveProposal so both entry points behave identically.
func (a *App) runContinuation(messages []llm.Message) {
	for step := range agent.ContinueConversation(a.agentCtx, a.client, messages, a.maxStepsFromConfig(), a.currentMode()) {
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
		if step.Type == agent.StepTypeAwaitingApproval {
			var proposal tools.Proposal
			if err := json.Unmarshal([]byte(step.Content), &proposal); err != nil {
				runtime.EventsEmit(a.ctx, "agent:error", "Failed to parse proposal: "+err.Error())
				return
			}
			a.pendingProposalsMu.Lock()
			a.pendingProposals[proposal.ID] = proposal
			a.pendingProposalsMu.Unlock()
			runtime.EventsEmit(a.ctx, "agent:awaiting_approval", proposal)
			return
		}
	}
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

		a.runContinuation(a.convManager.GetMessages())
	}()
}

// ResolveProposal handles the user's Approve/Deny decision on a pending
// az_propose proposal. On approve, it executes the command and streams the
// result back into the conversation; on deny, the agent is told and must
// adapt. Either way, the agent loop resumes afterward.
func (a *App) ResolveProposal(proposalID string, approved bool) {
	if a.client == nil || a.convManager == nil {
		runtime.EventsEmit(a.ctx, "agent:error", "Agent not ready to resolve proposal")
		return
	}

	a.pendingProposalsMu.Lock()
	proposal, ok := a.pendingProposals[proposalID]
	if ok {
		delete(a.pendingProposals, proposalID)
	}
	a.pendingProposalsMu.Unlock()

	if !ok {
		runtime.EventsEmit(a.ctx, "agent:error", "Unknown or already-resolved proposal: "+proposalID)
		return
	}

	decidedBy := currentOSUser()

	var followUp string
	if approved {
		result := tools.ExecuteApprovedProposal(proposal.Command, proposal.Tier, decidedBy)
		if result.Success {
			followUp = fmt.Sprintf("[System] Command approved by %s and executed:\n%s\n\nOutput:\n%s", decidedBy, proposal.Command, result.Output)
		} else {
			followUp = fmt.Sprintf("[System] Command approved by %s but failed to execute:\n%s\n\nError:\n%s", decidedBy, proposal.Command, result.Error)
		}
	} else {
		tools.RecordDenial(proposal.Command, proposal.Tier, decidedBy)
		followUp = fmt.Sprintf("[System] Command denied by %s:\n%s\n\nDo not attempt this operation again without further instructions from the user.", decidedBy, proposal.Command)
	}

	if err := a.convManager.AddUserMessage(followUp); err != nil {
		runtime.EventsEmit(a.ctx, "agent:error", "Failed to record decision: "+err.Error())
		return
	}

	if a.agentCancel != nil {
		a.agentCancel()
	}
	a.agentCtx, a.agentCancel = context.WithCancel(context.Background())

	go a.runContinuation(a.convManager.GetMessages())
}

// currentOSUser best-efforts the local OS username for audit/approval
// attribution in this single-user desktop app.
func currentOSUser() string {
	u, err := user.Current()
	if err != nil || u.Username == "" {
		return "user"
	}
	return u.Username
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

		for step := range agent.RunLoop(a.agentCtx, a.client, task, taskContext, maxSteps, a.currentMode()) {
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
