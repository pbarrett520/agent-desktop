// Package azure provides bindings around the user's own authenticated az CLI
// session. This package never stores or proxies credentials — it only shells
// out to the az CLI that the consultant has already logged into themselves.
package azure

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// Context describes the current state of the local az CLI session.
type Context struct {
	Installed        bool   `json:"installed"`
	LoggedIn         bool   `json:"logged_in"`
	TenantID         string `json:"tenant_id,omitempty"`
	SubscriptionID   string `json:"subscription_id,omitempty"`
	SubscriptionName string `json:"subscription_name,omitempty"`
	User             string `json:"user,omitempty"`
	// Error holds a human-readable reason when Installed or LoggedIn is false,
	// e.g. "run `az login` to authenticate". It is not a raw error dump.
	Error string `json:"error,omitempty"`
}

// Subscription represents one entry from `az account list`.
type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenant_id"`
	State     string `json:"state"`
	IsDefault bool   `json:"is_default"`
}

// azAccountShow mirrors the fields we read from `az account show --output json`.
type azAccountShow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TenantID string `json:"tenantId"`
	User     struct {
		Name string `json:"name"`
	} `json:"user"`
}

// azAccountListEntry mirrors one entry of `az account list --output json`.
type azAccountListEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId"`
	State     string `json:"state"`
	IsDefault bool   `json:"isDefault"`
}

const azCommandTimeout = 10 * time.Second

// runAz runs the az CLI with the given args and returns combined stdout.
func runAz(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), azCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "az", args...)
	return cmd.Output()
}

// GetAzureContext detects whether az is installed, whether there's an active
// login, and which tenant/subscription/user is active.
func GetAzureContext() *Context {
	if _, err := exec.LookPath("az"); err != nil {
		return &Context{
			Installed: false,
			Error:     "az CLI not found. Install it and run `az login`.",
		}
	}

	out, err := runAz("account", "show", "--output", "json")
	if err != nil {
		return &Context{
			Installed: true,
			LoggedIn:  false,
			Error:     "Not logged in. Run `az login` to authenticate.",
		}
	}

	var acc azAccountShow
	if err := json.Unmarshal(out, &acc); err != nil {
		return &Context{
			Installed: true,
			LoggedIn:  false,
			Error:     "Could not parse `az account show` output. Run `az login` to re-authenticate.",
		}
	}

	return &Context{
		Installed:        true,
		LoggedIn:         true,
		TenantID:         acc.TenantID,
		SubscriptionID:   acc.ID,
		SubscriptionName: acc.Name,
		User:             acc.User.Name,
	}
}

// ListSubscriptions wraps `az account list` and returns all subscriptions
// visible to the current login.
func ListSubscriptions() ([]Subscription, error) {
	out, err := runAz("account", "list", "--output", "json")
	if err != nil {
		return nil, err
	}

	var entries []azAccountListEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, err
	}

	subs := make([]Subscription, 0, len(entries))
	for _, e := range entries {
		subs = append(subs, Subscription{
			ID:        e.ID,
			Name:      e.Name,
			TenantID:  e.TenantID,
			State:     e.State,
			IsDefault: e.IsDefault,
		})
	}
	return subs, nil
}

// SetSubscription wraps `az account set --subscription <id>` to switch the
// active subscription for the local az session.
func SetSubscription(id string) error {
	id = strings.TrimSpace(id)
	ctx, cancel := context.WithTimeout(context.Background(), azCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "az", "account", "set", "--subscription", id)
	return cmd.Run()
}
