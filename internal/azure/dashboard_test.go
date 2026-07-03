package azure

import "testing"

// These tests don't assume az is installed; they only assert the functions
// fail gracefully (return an error, not a panic) rather than assuming success.

func TestListResourceGroups_DoesNotPanic(t *testing.T) {
	_, err := ListResourceGroups()
	_ = err // may or may not error depending on whether az/login is present in this environment
}

func TestListVMPowerStates_DoesNotPanic(t *testing.T) {
	_, err := ListVMPowerStates()
	_ = err
}

func TestGetMonthlyCost_DegradesGracefullyWhenUnavailable(t *testing.T) {
	summary, err := GetMonthlyCost("00000000-0000-0000-0000-000000000000")
	if err != nil && summary != nil {
		t.Error("on error, summary should be nil")
	}
}
