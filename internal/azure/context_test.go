package azure

import "testing"

func TestGetAzureContext_HandlesMissingCLI(t *testing.T) {
	// This test runs in CI/dev environments where az may or may not be
	// installed. It only asserts the function never panics and returns a
	// well-formed Context either way.
	ctx := GetAzureContext()
	if ctx == nil {
		t.Fatal("GetAzureContext() returned nil")
	}
	if !ctx.Installed && ctx.Error == "" {
		t.Error("expected a human-readable Error when az is not installed")
	}
	if ctx.Installed && !ctx.LoggedIn && ctx.Error == "" {
		t.Error("expected a human-readable Error when not logged in")
	}
}

func TestSubscription_FieldsRoundTrip(t *testing.T) {
	s := Subscription{
		ID:        "00000000-0000-0000-0000-000000000000",
		Name:      "Contoso Dev",
		TenantID:  "11111111-1111-1111-1111-111111111111",
		State:     "Enabled",
		IsDefault: true,
	}
	if s.ID == "" || s.Name == "" || s.TenantID == "" {
		t.Fatal("Subscription fields should be settable")
	}
}
