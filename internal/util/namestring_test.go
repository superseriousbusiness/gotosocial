// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package util_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type NamestringSuite struct {
	suite.Suite
}

func (suite *NamestringSuite) TestExtractWebfingerParts() {
	tests := []struct {
		in, username, domain, err string
	}{
		{in: "acct:stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "acct:stonerkitty.monster@stonerkitty.monster:8080", username: "stonerkitty.monster", domain: "stonerkitty.monster:8080"},
		{in: "acct:@stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "stonerkitty.monster@stonerkitty.monster:8080", username: "stonerkitty.monster", domain: "stonerkitty.monster:8080"},
		{in: "@stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "acct:@@stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "acct:@stonerkitty.monster@@stonerkitty.monster", err: "couldn't match namestring @stonerkitty.monster@@stonerkitty.monster"},
		{in: "@@stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "@stonerkitty.monster@@stonerkitty.monster", err: "couldn't match namestring @stonerkitty.monster@@stonerkitty.monster"},
		{in: "s3:stonerkitty.monster@stonerkitty.monster", err: "unsupported scheme s3 for resource s3:stonerkitty.monster@stonerkitty.monster"},
		{in: "https://stonerkitty.monster/users/stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "https://stonerkitty.monster/users/@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "https://stonerkitty.monster/@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "https://stonerkitty.monster/@@stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "https://stonerkitty.monster:8080/users/stonerkitty.monster", username: "stonerkitty.monster", domain: "stonerkitty.monster:8080"},
		{in: "https://stonerkitty.monster/users/stonerkitty.monster/evil", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "https://stonerkitty.monster/@stonerkitty.monster/evil", username: "stonerkitty.monster", domain: "stonerkitty.monster"},
		{in: "/@stonerkitty.monster", err: "no scheme for resource /@stonerkitty.monster"},
		{in: "/users/stonerkitty.monster", err: "no scheme for resource /users/stonerkitty.monster"},
		{in: "@stonerkitty.monster", err: "failed to extract domain from: @stonerkitty.monster"},
		{in: "users/stonerkitty.monster", err: "couldn't match namestring @users/stonerkitty.monster"},
		{in: "https://stonerkitty.monster/users/", err: "failed to extract username from: https://stonerkitty.monster/users/"},
		{in: "https://stonerkitty.monster/users/@", err: "failed to extract username from: https://stonerkitty.monster/users/@"},
		{in: "https://stonerkitty.monster/@", err: "failed to extract username from: https://stonerkitty.monster/@"},
		{in: "https://stonerkitty.monster/", err: "failed to extract username from: https://stonerkitty.monster/"},
	}

	for _, tt := range tests {
		tt := tt
		suite.Run(tt.in, func() {
			suite.T().Parallel()
			username, domain, err := util.ExtractWebfingerParts(tt.in)
			if tt.err == "" {
				suite.NoError(err)
				suite.Equal(tt.username, username)
				suite.Equal(tt.domain, domain)
			} else {
				if !suite.EqualError(err, tt.err) {
					suite.T().Logf("expected error %s", tt.err)
				}
			}
		})
	}
}

func (suite *NamestringSuite) TestExtractNamestring() {
	tests := []struct {
		in, username, host, err string
	}{
		{in: "@stonerkitty.monster@stonerkitty.monster", username: "stonerkitty.monster", host: "stonerkitty.monster"},
		{in: "@stonerkitty.monster", username: "stonerkitty.monster"},
		{in: "@someone@somewhere", username: "someone", host: "somewhere"},
		{in: "", err: "couldn't match namestring "},
	}

	for _, tt := range tests {
		tt := tt
		suite.Run(tt.in, func() {
			suite.T().Parallel()
			username, host, err := util.ExtractNamestringParts(tt.in)
			if tt.err != "" {
				suite.EqualError(err, tt.err)
			} else {
				suite.NoError(err)
				suite.Equal(tt.username, username)
				suite.Equal(tt.host, host)
			}
		})
	}
}

func TestNamestringSuite(t *testing.T) {
	suite.Run(t, &NamestringSuite{})
}
