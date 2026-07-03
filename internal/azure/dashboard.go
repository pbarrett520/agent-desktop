package azure

import (
	"encoding/json"
)

// ResourceGroupSummary is one row of the subscription overview's resource
// group table.
type ResourceGroupSummary struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceCount int    `json:"resource_count"`
}

// VMPowerState is one row of the subscription overview's VM power state table.
type VMPowerState struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resource_group"`
	PowerState    string `json:"power_state"`
}

// CostSummary is this month's cost, when the costmanagement extension is
// available. Degrades gracefully (nil, error) when it isn't.
type CostSummary struct {
	Currency    string  `json:"currency"`
	AmountToDate float64 `json:"amount_to_date"`
}

type azResourceGroupEntry struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type azResourceEntry struct {
	ResourceGroup string `json:"resourceGroup"`
}

// ListResourceGroups returns every resource group in the active subscription
// along with how many resources each one contains.
func ListResourceGroups() ([]ResourceGroupSummary, error) {
	groupsOut, err := runAz("group", "list", "--output", "json")
	if err != nil {
		return nil, err
	}
	var groups []azResourceGroupEntry
	if err := json.Unmarshal(groupsOut, &groups); err != nil {
		return nil, err
	}

	counts := map[string]int{}
	resourcesOut, err := runAz("resource", "list", "--output", "json")
	if err == nil {
		var resources []azResourceEntry
		if json.Unmarshal(resourcesOut, &resources) == nil {
			for _, r := range resources {
				counts[r.ResourceGroup]++
			}
		}
	}
	// A failure to list resources still yields the resource group list with
	// zero counts, rather than failing the whole dashboard.

	summaries := make([]ResourceGroupSummary, 0, len(groups))
	for _, g := range groups {
		summaries = append(summaries, ResourceGroupSummary{
			Name:          g.Name,
			Location:      g.Location,
			ResourceCount: counts[g.Name],
		})
	}
	return summaries, nil
}

type azVMPowerStateEntry struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
	PowerState    string `json:"powerState"`
}

// ListVMPowerStates returns the power state of every VM in the active
// subscription.
func ListVMPowerStates() ([]VMPowerState, error) {
	out, err := runAz("vm", "list", "-d", "--query", "[].{name:name, resourceGroup:resourceGroup, powerState:powerState}", "--output", "json")
	if err != nil {
		return nil, err
	}
	var entries []azVMPowerStateEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, err
	}

	states := make([]VMPowerState, 0, len(entries))
	for _, e := range entries {
		states = append(states, VMPowerState{Name: e.Name, ResourceGroup: e.ResourceGroup, PowerState: e.PowerState})
	}
	return states, nil
}

type azCostQueryResponse struct {
	Properties struct {
		Rows [][]interface{} `json:"rows"`
	} `json:"properties"`
}

// GetMonthlyCost returns this month's cost to date, if the costmanagement
// az extension is installed and the account has access. Returns an error
// (not a panic) when unavailable so the caller can degrade gracefully.
func GetMonthlyCost(subscriptionID string) (*CostSummary, error) {
	out, err := runAz(
		"costmanagement", "query",
		"--type", "ActualCost",
		"--timeframe", "MonthToDate",
		"--scope", "/subscriptions/"+subscriptionID,
		"--output", "json",
	)
	if err != nil {
		return nil, err
	}

	var resp azCostQueryResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, err
	}
	if len(resp.Properties.Rows) == 0 {
		return &CostSummary{Currency: "", AmountToDate: 0}, nil
	}

	row := resp.Properties.Rows[0]
	amount, _ := row[0].(float64)
	currency := ""
	if len(row) > 1 {
		currency, _ = row[len(row)-1].(string)
	}

	return &CostSummary{Currency: currency, AmountToDate: amount}, nil
}
