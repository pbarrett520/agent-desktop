package tools

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"agent-desktop/internal/audit"
	"agent-desktop/internal/azure"
	"agent-desktop/internal/safety"
)

// Proposal is a MUTATE or DESTRUCTIVE az command awaiting human approval.
// AzPropose returns one serialized as JSON in ToolResult.Output; the frontend
// renders it as an approval card and the backend later executes it (or not)
// via ExecuteApprovedProposal.
type Proposal struct {
	ID           string `json:"id"`
	Command      string `json:"command"`
	Explanation  string `json:"explanation"`
	RollbackHint string `json:"rollback_hint"`
	Tier         string `json:"tier"`
	Sensitive    bool   `json:"sensitive"`
}

// currentSessionID is stamped onto every audit record. Set once at app
// startup via SetSessionID.
var currentSessionID string

// SetSessionID sets the session identifier attached to audit records
// written by this package.
func SetSessionID(id string) {
	currentSessionID = id
}

// generateProposalID returns a short random hex identifier.
func generateProposalID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("prop-%d", time.Now().UnixNano())
	}
	return "prop-" + hex.EncodeToString(buf)
}

// hasOutputFlag reports whether command already specifies an --output/-o flag.
func hasOutputFlag(command string) bool {
	return strings.Contains(command, "--output") || strings.Contains(command, " -o ") || strings.HasSuffix(strings.TrimSpace(command), " -o")
}

// ensureJSONOutput appends "--output json" if the command doesn't already
// specify an output format.
func ensureJSONOutput(command string) string {
	if hasOutputFlag(command) {
		return command
	}
	return strings.TrimRight(command, " ") + " --output json"
}

// currentSubscriptionID best-efforts the active subscription ID for audit
// records. Failures are non-fatal — audit entries just omit it.
func currentSubscriptionID() string {
	ctx := azure.GetAzureContext()
	if ctx == nil || !ctx.LoggedIn {
		return ""
	}
	return ctx.SubscriptionID
}

// AzQuery executes a READ-tier az command directly and records it in the
// audit trail. If the command doesn't classify as READ, it refuses and
// tells the agent to use az_propose instead.
func AzQuery(command string) ToolResult {
	classification := safety.ClassifyAzCommand(command)
	if classification.Tier != safety.Read {
		return ToolResult{
			Success: false,
			Error: fmt.Sprintf(
				"az_query only executes READ-tier commands. This command classified as %s (%s). Use az_propose to request approval instead.",
				classification.Tier, classification.Reason,
			),
		}
	}

	cmd := ensureJSONOutput(command)
	start := time.Now()
	result := runShell(cmd, GetSession().CWD, 60)
	duration := time.Since(start)

	exitCode := 0
	if !result.Success {
		exitCode = -1
	}

	audit.Record(audit.Event{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		SessionID:      currentSessionID,
		SubscriptionID: currentSubscriptionID(),
		Tier:           classification.Tier.String(),
		Command:        cmd,
		ProposedBy:     "agent",
		Decision:       "auto",
		ExitCode:       exitCode,
		DurationMs:     duration.Milliseconds(),
		OutputHash:     audit.HashOutput(result.Output),
	})

	return result
}

// AzPropose classifies command and returns a Proposal for the frontend to
// render as an approval card. It never executes anything.
func AzPropose(command string, explanation string, rollbackHint string) ToolResult {
	classification := safety.ClassifyAzCommand(command)

	p := Proposal{
		ID:           generateProposalID(),
		Command:      command,
		Explanation:  explanation,
		RollbackHint: rollbackHint,
		Tier:         classification.Tier.String(),
		Sensitive:    classification.Sensitive,
	}

	data, err := json.Marshal(p)
	if err != nil {
		return ToolResult{Success: false, Error: "failed to build proposal: " + err.Error()}
	}

	return ToolResult{Success: true, Output: string(data)}
}

// ExecuteApprovedProposal runs a previously-proposed command after a human
// has approved it, appending --output json if not already present, and
// writes the audit record. Callers are responsible for having verified the
// proposal was actually approved by a human before calling this.
func ExecuteApprovedProposal(command string, tier string, approvedBy string) ToolResult {
	cmd := ensureJSONOutput(command)
	start := time.Now()
	result := runShell(cmd, GetSession().CWD, 120)
	duration := time.Since(start)

	exitCode := 0
	if !result.Success {
		exitCode = -1
	}

	audit.Record(audit.Event{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		SessionID:      currentSessionID,
		SubscriptionID: currentSubscriptionID(),
		Tier:           tier,
		Command:        cmd,
		ProposedBy:     "agent",
		Decision:       "approved:" + approvedBy,
		ExitCode:       exitCode,
		DurationMs:     duration.Milliseconds(),
		OutputHash:     audit.HashOutput(result.Output),
	})

	return result
}

// RecordDenial writes an audit record for a proposal the user denied. No
// command is executed.
func RecordDenial(command string, tier string, deniedBy string) {
	audit.Record(audit.Event{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		SessionID:      currentSessionID,
		SubscriptionID: currentSubscriptionID(),
		Tier:           tier,
		Command:        command,
		ProposedBy:     "agent",
		Decision:       "denied:" + deniedBy,
		ExitCode:       0,
		DurationMs:     0,
		OutputHash:     "",
	})
}
