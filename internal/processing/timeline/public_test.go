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
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/stretchr/testify/suite"
)

type PublicTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *PublicTestSuite) TestPublicTimelineGet() {
	var (
		ctx       = suite.T().Context()
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
		&paging.Page{
			Min:   paging.EitherMinID(minID, sinceID),
			Max:   paging.MaxID(maxID),
			Limit: limit,
		},
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
		ctx       = suite.T().Context()
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
		&paging.Page{
			Min:   paging.EitherMinID(minID, sinceID),
			Max:   paging.MaxID(maxID),
			Limit: limit,
		},
		local,
	)

	// We should have a status even though
	// some other statuses were filtered out.
	suite.NoError(errWithCode)
	suite.Len(resp.Items, 1)
	suite.Equal(`<http://localhost:8080/api/v1/timelines/public?limit=1&local=false&max_id=01F8MHCP5P2NWYQ416SBA0XSEV>; rel="next", <http://localhost:8080/api/v1/timelines/public?limit=1&local=false&min_id=01HE7XJ1CG84TBKH5V9XKBVGF5>; rel="prev"`, resp.LinkHeader)
	suite.Equal(`http://localhost:8080/api/v1/timelines/public?limit=1&local=false&max_id=01F8MHCP5P2NWYQ416SBA0XSEV`, resp.NextLink)
	suite.Equal(`http://localhost:8080/api/v1/timelines/public?limit=1&local=false&min_id=01HE7XJ1CG84TBKH5V9XKBVGF5`, resp.PrevLink)
}

// A timeline containing a status hidden due to filtering should return other statuses with no error.
func (suite *PublicTestSuite) TestPublicTimelineGetHideFiltered() {
	var (
		ctx                 = suite.T().Context()
		requester           = suite.testAccounts["local_account_1"]
		maxID               = ""
		sinceID             = ""
		minID               = ""
		limit               = 100
		local               = false
		filteredStatus      = suite.testStatuses["admin_account_status_2"]
		filteredStatusFound = false
		filterID            = id.NewULID()
		filterStatusID      = id.NewULID()
		filterStatus        = &gtsmodel.FilterStatus{
			ID:       filterStatusID,
			FilterID: filterID,
			StatusID: filteredStatus.ID,
		}
		filter = &gtsmodel.Filter{
			ID:        filterID,
			AccountID: requester.ID,
			Title:     "timeline filtering test",
			Action:    gtsmodel.FilterActionHide,
			Statuses:  []*gtsmodel.FilterStatus{filterStatus},
			StatusIDs: []string{filterStatusID},
			Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextPublic),
		}
	)

	// Fetch the timeline to make sure the status we're going to filter is in that section of it.
	resp, errWithCode := suite.timeline.PublicTimelineGet(
		ctx,
		requester,
		&paging.Page{
			Min:   paging.EitherMinID(minID, sinceID),
			Max:   paging.MaxID(maxID),
			Limit: limit,
		},
		local,
	)
	suite.NoError(errWithCode)
	for _, item := range resp.Items {
		if item.(*apimodel.Status).ID == filteredStatus.ID {
			filteredStatusFound = true
			break
		}
	}
	if !filteredStatusFound {
		suite.FailNow("precondition failed: status we would filter isn't present in unfiltered timeline")
	}

	// Create the filter status associated with the main filter.
	if err := suite.db.PutFilterStatus(ctx, filterStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// Create a filter to hide one status on the timeline.
	if err := suite.db.PutFilter(ctx, filter); err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch the timeline again with the filter in place.
	resp, errWithCode = suite.timeline.PublicTimelineGet(
		ctx,
		requester,
		&paging.Page{
			Min:   paging.EitherMinID(minID, sinceID),
			Max:   paging.MaxID(maxID),
			Limit: limit,
		},
		local,
	)

	// We should have some statuses even though one status was filtered out.
	suite.NoError(errWithCode)
	suite.NotEmpty(resp.Items)
	// The filtered status should not be there.
	filteredStatusFound = false
	for _, item := range resp.Items {
		if item.(*apimodel.Status).ID == filteredStatus.ID {
			filteredStatusFound = true
			break
		}
	}
	suite.False(filteredStatusFound)
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
