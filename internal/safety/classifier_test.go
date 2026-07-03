package safety

import "testing"

func TestClassifyAzCommand(t *testing.T) {
	cases := []struct {
		name      string
		cmd       string
		wantTier  Tier
		sensitive bool
	}{
		// READ tier
		{"vm list", "az vm list", Read, false},
		{"group show", "az group show --name rg-prod", Read, false},
		{"account show", "az account show", Read, false},
		{"account list", "az account list --output json", Read, false},
		{"graph query", "az graph query -q \"Resources | limit 5\"", Read, false},
		{"resource list", "az resource list --resource-group rg-prod", Read, false},
		{"vm get-instance-view", "az vm get-instance-view --name vm1 --resource-group rg1", Read, false},
		{"describe verb", "az monitor metrics describe", Read, false},
		{"top verb", "az monitor top-metrics", Read, false},
		{"help flag on delete", "az group delete --help", Read, false},
		{"keyvault secret show is read but sensitive", "az keyvault secret show --name db-pass --vault-name kv1", Read, true},
		{"keyvault key list is read but sensitive", "az keyvault key list --vault-name kv1", Read, true},

		// MUTATE tier
		{"vm create", "az vm create --name vm1 --resource-group rg1", Mutate, false},
		{"webapp update", "az webapp update --name app1", Mutate, false},
		{"vm start", "az vm start --name vm1 --resource-group rg1", Mutate, false},
		{"vm stop", "az vm stop --name vm1 --resource-group rg1", Mutate, false},
		{"vm restart", "az vm restart --name vm1 --resource-group rg1", Mutate, false},
		{"webapp deploy", "az webapp deploy --name app1", Mutate, false},
		{"vmss scale", "az vmss scale --name vmss1 --new-capacity 3", Mutate, false},
		{"resource tag", "az resource tag --tags env=prod --ids /sub/x", Mutate, false},
		{"account set", "az account set --subscription 00000000-0000-0000-0000-000000000000", Mutate, false},
		{"unknown verb defaults to mutate", "az vm frobnicate --name vm1", Mutate, false},
		{"unrecognized tool entirely", "az widget spin", Mutate, false},

		// DESTRUCTIVE tier
		{"group delete", "az group delete --name rg-old --yes", Destructive, false},
		{"vm delete", "az vm delete --name vm1 --resource-group rg1 --yes", Destructive, false},
		{"keyvault purge", "az keyvault purge --name kv1", Destructive, false},
		{"webapp delete", "az webapp delete --name app1", Destructive, false},
		{"role assignment create", "az role assignment create --assignee foo@bar.com --role Owner", Destructive, false},
		{"role assignment delete", "az role assignment delete --assignee foo@bar.com --role Owner", Destructive, false},
		{"role definition delete", "az role definition delete --name custom-role", Destructive, false},
		{"keyvault set-policy", "az keyvault set-policy --name kv1 --object-id abc --secret-permissions get", Destructive, false},
		{"ad app create touches directory", "az ad app create --display-name app1", Destructive, false},
		{"ad user delete", "az ad user delete --id abc", Destructive, false},
		{"role assignment reset via revoke wording", "az role assignment delete --scope /subscriptions/x --role Contributor", Destructive, false},
		{"deployment cancel", "az deployment group cancel --name dep1 --resource-group rg1", Destructive, false},
		{"vm reset password", "az vm user reset-ssh --name vm1", Destructive, false},

		// Chained / piped commands: classify by the most dangerous segment
		{"read then destructive chain", "az vm list && az group delete --name rg-old --yes", Destructive, false},
		{"destructive then read chain", "az group delete --name rg-old --yes; az vm list", Destructive, false},
		{"read piped to grep is still read", "az vm list --output json | grep running", Read, false},
		{"mutate then read semicolon", "az vm start --name vm1; az vm list", Mutate, false},
		{"read and sensitive read chain stays sensitive", "az vm list && az keyvault secret show --name x --vault-name kv1", Read, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyAzCommand(tc.cmd)
			if got.Tier != tc.wantTier {
				t.Errorf("ClassifyAzCommand(%q) tier = %s, want %s (reason: %s)", tc.cmd, got.Tier, tc.wantTier, got.Reason)
			}
			if got.Sensitive != tc.sensitive {
				t.Errorf("ClassifyAzCommand(%q) sensitive = %v, want %v", tc.cmd, got.Sensitive, tc.sensitive)
			}
			if got.Reason == "" {
				t.Errorf("ClassifyAzCommand(%q) should always have a reason", tc.cmd)
			}
		})
	}
}

func TestClassifyAzCommand_EmptyCommand(t *testing.T) {
	got := ClassifyAzCommand("")
	if got.Tier != Mutate {
		t.Errorf("empty command should fail closed to Mutate, got %s", got.Tier)
	}
}

func TestTier_String(t *testing.T) {
	cases := map[Tier]string{
		Read:        "READ",
		Mutate:      "MUTATE",
		Destructive: "DESTRUCTIVE",
	}
	for tier, want := range cases {
		if got := tier.String(); got != want {
			t.Errorf("Tier(%d).String() = %q, want %q", tier, got, want)
		}
	}
}
