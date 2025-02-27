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

package dereferencing_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InstanceTestSuite struct {
	DereferencerStandardTestSuite
}

func (suite *InstanceTestSuite) TestDerefInstance() {
	type testCase struct {
		instanceIRI      *url.URL
		expectedSoftware string
	}

	for _, tc := range []testCase{
		{
			// Fossbros anonymous doesn't shield their nodeinfo or
			// well-known or anything so we should be able to fetch.
			instanceIRI:      testrig.URLMustParse("https://fossbros-anonymous.io"),
			expectedSoftware: "Hellsoft 6.6.6",
		},
		{
			// Furtive nerds forbids /nodeinfo using
			// robots.txt so we should get bare minimum only.
			//
			// Debug-level logs should show something like:
			//
			//   - "can't fetch /nodeinfo/2.1: robots.txt disallows it"
			instanceIRI:      testrig.URLMustParse("https://furtive-nerds.example.org"),
			expectedSoftware: "",
		},
		{
			// Robotic furtive nerds forbids *everything* using
			// robots.txt so we should get bare minimum only.
			//
			// Debug-level logs should show something like:
			//
			//   - "can't fetch api/v1/instance: robots.txt disallows it"
			//   - "can't fetch .well-known/nodeinfo: robots.txt disallows it"
			instanceIRI:      testrig.URLMustParse("https://robotic.furtive-nerds.example.org"),
			expectedSoftware: "",
		},
		{
			// Really furtive nerds forbids .well-known/nodeinfo using
			// X-Robots-Tagheaders, so we should get bare minimum only.
			//
			// Debug-level logs should show something like:
			//
			//   - "can't use fetched .well-known/nodeinfo: robots tags disallows it"
			instanceIRI:      testrig.URLMustParse("https://really.furtive-nerds.example.org"),
			expectedSoftware: "",
		},
	} {
		instance, err := suite.dereferencer.GetRemoteInstance(
			gtscontext.SetFastFail(context.Background()),
			suite.testAccounts["admin_account"].Username,
			tc.instanceIRI,
		)
		if err != nil {
			suite.FailNow(err.Error())
		}

		suite.Equal(tc.expectedSoftware, instance.Version)
	}
}

func TestInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(InstanceTestSuite))
}
