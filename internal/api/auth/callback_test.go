package auth

import (
	"testing"

	"github.com/superseriousbusiness/gotosocial/testrig"
)

func TestNotInAdminGroup(t *testing.T) {
	testrig.InitTestConfig()
	// Test if user has allowed group but not in admin group
	groups := []string{"group1", "group2", "allowedRole"}

	got := adminGroup(groups)
	want := false

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestInAdminGroup(t *testing.T) {
	testrig.InitTestConfig()
	// Test if user has admin group
	groups := []string{"group1", "group2", "adminRole"}

	got := adminGroup(groups)
	want := true

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestNotInAllowedGroup(t *testing.T) {
	testrig.InitTestConfig()
	// Test if user has admin group but not in allowed group
	groups := []string{"group1", "group2", "adminRole"}

	got := allowedGroup(groups)
	want := false

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestInAllowedGroup(t *testing.T) {
	testrig.InitTestConfig()
	// Test if user has allowed group
	groups := []string{"group1", "group2", "allowedRole"}

	got := allowedGroup(groups)
	want := true

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}
