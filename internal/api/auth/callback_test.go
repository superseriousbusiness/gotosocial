package auth

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
)

type CallbackTestSuite struct {
	// standard suite interfaces
	suite.Suite
}

func (suite *CallbackTestSuite) TestAllowedByClaims() {
	tests := []struct {
		name   string
		claims *oidc.Claims
		rules  []config.OIDCRequirement
		pass   bool
	}{
		{
			name:   "fail closed",
			claims: &oidc.Claims{},
			rules:  nil,
			pass:   false,
		},
		{
			name: "single claim and no matching rules",
			claims: &oidc.Claims{Attrs: map[string][]string{
				"apple": {"pie"},
			}},
			rules: nil,
			pass:  false,
		},
		{
			name: "single claim and matching rule",
			claims: &oidc.Claims{Attrs: map[string][]string{
				"apple": {"pie"},
			}},
			rules: []config.OIDCRequirement{
				{Claim: "apple", Value: "pie"},
			},
			pass: true,
		},
		{
			name: "single claim and multiple rules",
			claims: &oidc.Claims{Attrs: map[string][]string{
				"apple": {"pie"},
			}},
			rules: []config.OIDCRequirement{
				{Claim: "apple", Value: "pie"},
				{Claim: "blueberry", Value: "muffin"},
			},
			pass: false,
		},
		{
			name: "multiple claims and single matching rule",
			claims: &oidc.Claims{Attrs: map[string][]string{
				"apple":     {"pie"},
				"blueberry": {"muffin"},
			}},
			rules: []config.OIDCRequirement{
				{Claim: "blueberry", Value: "muffin"},
			},
			pass: true,
		},
		{
			name: "multiple claims and multiple matching rules",
			claims: &oidc.Claims{Attrs: map[string][]string{
				"apple":     {"pie"},
				"blueberry": {"muffin"},
			}},
			rules: []config.OIDCRequirement{
				{Claim: "apple", Value: "pie"},
				{Claim: "blueberry", Value: "muffin"},
			},
			pass: true,
		},
		{
			name: "multiple claims and only partial matching rule",
			claims: &oidc.Claims{Attrs: map[string][]string{
				"apple":     {"pie"},
				"blueberry": {"muffin"},
			}},
			rules: []config.OIDCRequirement{
				{Claim: "apple", Value: "pie"},
				{Claim: "blueberry", Value: "cobbler"},
			},
			pass: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		suite.Run(tt.name, func() {
			res := allowedByClaims(tt.claims, tt.rules)
			if tt.pass != res {
				suite.FailNowf("rule evaluation incorrect", "expected: %t, got: %t", tt.pass, res)
			}
		})
	}
}

func TestCallbackTestSuite(t *testing.T) {
	suite.Run(t, &CallbackTestSuite{})
}
