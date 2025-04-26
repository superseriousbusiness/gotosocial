package auth

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/testrig"
)

func TestAdminGroup(t *testing.T) {
	testrig.InitTestConfig()
	for _, test := range []struct {
		name     string
		groups   []string
		expected bool
	}{
		{name: "not in admin group", groups: []string{"group1", "group2", "allowedRole"}, expected: false},
		{name: "in admin group", groups: []string{"group1", "group2", "adminRole"}, expected: true},
	} {
		test := test // loopvar capture
		t.Run(test.name, func(t *testing.T) {
			if got := adminGroup(test.groups); got != test.expected {
				t.Fatalf("got: %t, wanted: %t", got, test.expected)
			}
		})
	}
}

func TestAllowedGroup(t *testing.T) {
	testrig.InitTestConfig()
	for _, test := range []struct {
		name     string
		groups   []string
		expected bool
	}{
		{name: "not in allowed group", groups: []string{"group1", "group2", "adminRole"}, expected: false},
		{name: "in allowed group", groups: []string{"group1", "group2", "allowedRole"}, expected: true},
	} {
		test := test // loopvar capture
		t.Run(test.name, func(t *testing.T) {
			if got := allowedGroup(test.groups); got != test.expected {
				t.Fatalf("got: %t, wanted: %t", got, test.expected)
			}
		})
	}
}
