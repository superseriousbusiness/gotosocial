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

package timeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PublicTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *PublicTestSuite) TestPublicTimelineGet() {
	var (
		ctx       = context.Background()
		requester = suite.testAccounts["local_account_1"]
		maxID     = ""
		sinceID   = ""
		minID     = ""
		limit     = 10
		local     = false
	)

	resp, errWithCode := suite.timeline.PublicTimelineGet(
		ctx,
		requester,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)

	// We should have some statuses,
	// and paging headers should be set.
	suite.NoError(errWithCode)
	suite.NotEmpty(resp.Items)
	suite.NotEmpty(resp.LinkHeader)
	suite.NotEmpty(resp.NextLink)
	suite.NotEmpty(resp.PrevLink)
}

func (suite *PublicTestSuite) TestPublicTimelineGetNotEmpty() {
	var (
		ctx       = context.Background()
		requester = suite.testAccounts["local_account_1"]
		// Select 1 *just above* a status we know should
		// not be in the public timeline -- a public
		// reply to one of admin's statuses.
		maxID   = "01HE7XJ1CG84TBKH5V9XKBVGF6"
		sinceID = ""
		minID   = ""
		limit   = 1
		local   = false
	)

	resp, errWithCode := suite.timeline.PublicTimelineGet(
		ctx,
		requester,
		maxID,
		sinceID,
		minID,
		limit,
		local,
	)

	// We should have a status even though
	// some other statuses were filtered out.
	suite.NoError(errWithCode)
	suite.Len(resp.Items, 1)
	suite.Equal(`<http://localhost:8080/api/v1/timelines/public?limit=1&max_id=01F8MHCP5P2NWYQ416SBA0XSEV&local=false>; rel="next", <http://localhost:8080/api/v1/timelines/public?limit=1&min_id=01HE7XJ1CG84TBKH5V9XKBVGF5&local=false>; rel="prev"`, resp.LinkHeader)
	suite.Equal(`http://localhost:8080/api/v1/timelines/public?limit=1&max_id=01F8MHCP5P2NWYQ416SBA0XSEV&local=false`, resp.NextLink)
	suite.Equal(`http://localhost:8080/api/v1/timelines/public?limit=1&min_id=01HE7XJ1CG84TBKH5V9XKBVGF5&local=false`, resp.PrevLink)
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
