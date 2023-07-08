package oidc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OIDCTestSuite struct {
	// standard suite interfaces
	suite.Suite
}

func (suite *OIDCTestSuite) TestOIDCClaimsUnmarshall() {
	tests := []struct {
		name string
		in   string
		err  bool
		res  Claims
	}{
		{
			name: "no unknown claims",
			in:   `{"sub":"a", "email":"b", "email_verified": true, "name": "c", "preferred_username": "c"}`,
			res:  Claims{Sub: "a", Email: "b", EmailVerified: true, Name: "c", PreferredUsername: "c", Attrs: map[string][]string{}},
		},
		{
			name: "unknown claim with string value",
			in:   `{"sub":"a", "email":"b", "email_verified": true, "name": "c", "preferred_username": "c", "groups": "test"}`,
			res: Claims{Sub: "a", Email: "b", EmailVerified: true, Name: "c", PreferredUsername: "c", Attrs: map[string][]string{
				"groups": {"test"},
			}},
		},
		{
			name: "unknown claim with list of string value",
			in:   `{"sub":"a", "email":"b", "email_verified": true, "name": "c", "preferred_username": "c", "groups": ["test1", "test2"]}`,
			res: Claims{Sub: "a", Email: "b", EmailVerified: true, Name: "c", PreferredUsername: "c", Attrs: map[string][]string{
				"groups": {"test1", "test2"},
			}},
		},
		{
			name: "unknown claim with list of int value",
			in:   `{"sub":"a", "email":"b", "email_verified": true, "name": "c", "preferred_username": "c", "groups": [1, 2]}`,
			res:  Claims{Sub: "a", Email: "b", EmailVerified: true, Name: "c", PreferredUsername: "c", Attrs: map[string][]string{}},
		},
	}
	for _, tt := range tests {
		tt := tt
		suite.Run(tt.name, func() {
			var c Claims
			err := json.Unmarshal([]byte(tt.in), &c)
			if !tt.err && err != nil {
				suite.FailNowf("expected success unmarshalling", "got error: %s", err.Error())
			}
			if tt.err && err == nil {
				suite.FailNowf("expected failed unmarshalling", "got no error")
			}
			if !reflect.DeepEqual(tt.res, c) {
				suite.T().Log(c.Attrs)
				suite.FailNow("expected structs to be identical")
			}
		})
	}
}

func TestOIDCTestSuite(t *testing.T) {
	suite.Run(t, &OIDCTestSuite{})
}
