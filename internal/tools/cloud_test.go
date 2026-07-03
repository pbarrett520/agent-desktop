package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAzQuery_RefusesNonReadCommands(t *testing.T) {
	result := AzQuery("az group delete --name rg-old --yes")
	if result.Success {
		t.Fatal("AzQuery should refuse a DESTRUCTIVE command")
	}
	if !strings.Contains(result.Error, "az_propose") {
		t.Errorf("refusal should point to az_propose, got: %s", result.Error)
	}
}

func TestAzQuery_RefusesMutateCommands(t *testing.T) {
	result := AzQuery("az vm start --name vm1 --resource-group rg1")
	if result.Success {
		t.Fatal("AzQuery should refuse a MUTATE command")
	}
}

func TestEnsureJSONOutput_AppendsWhenMissing(t *testing.T) {
	got := ensureJSONOutput("az vm list")
	if !strings.Contains(got, "--output json") {
		t.Errorf("expected --output json to be appended, got: %s", got)
	}
}

func TestEnsureJSONOutput_LeavesExistingFlag(t *testing.T) {
	got := ensureJSONOutput("az vm list --output table")
	if strings.Count(got, "--output") != 1 {
		t.Errorf("should not append a second --output flag, got: %s", got)
	}
}

func TestAzPropose_ReturnsProposalWithoutExecuting(t *testing.T) {
	result := AzPropose("az group delete --name rg-old --yes", "Removes the decommissioned dev resource group", "Resource group cannot be recovered once purged")
	if !result.Success {
		t.Fatalf("AzPropose should always succeed (it only builds a proposal), got error: %s", result.Error)
	}

	var p Proposal
	if err := json.Unmarshal([]byte(result.Output), &p); err != nil {
		t.Fatalf("AzPropose output should be a JSON Proposal, got parse error: %v", err)
	}
	if p.Tier != "DESTRUCTIVE" {
		t.Errorf("expected Tier=DESTRUCTIVE, got %s", p.Tier)
	}
	if p.ID == "" {
		t.Error("expected a non-empty proposal ID")
	}
	if p.Command == "" || p.Explanation == "" || p.RollbackHint == "" {
		t.Error("proposal should echo back command, explanation, and rollback hint")
	}
}

func TestGenerateProposalID_Unique(t *testing.T) {
	ids := map[string]bool{}
	for i := 0; i < 20; i++ {
		id := generateProposalID()
		if ids[id] {
			t.Fatalf("generateProposalID produced a duplicate: %s", id)
		}
		ids[id] = true
	}
}
