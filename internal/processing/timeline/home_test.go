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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type HomeTestSuite struct {
	TimelineStandardTestSuite
}

func (suite *HomeTestSuite) TearDownTest() {
	suite.TimelineStandardTestSuite.TearDownTest()
}

// A timeline containing a status hidden due to filtering should return other statuses with no error.
func (suite *HomeTestSuite) TestHomeTimelineGetHideFiltered() {
	var (
		ctx                 = context.Background()
		requester           = suite.testAccounts["local_account_1"]
		maxID               = ""
		sinceID             = ""
		minID               = "01F8MHAAY43M6RJ473VQFCVH36" // 1 before filteredStatus
		limit               = 40
		local               = false
		filteredStatus      = suite.testStatuses["admin_account_status_2"]
		filteredStatusFound = false
		filterID            = id.NewULID()
		filter              = &gtsmodel.Filter{
			ID:        filterID,
			AccountID: requester.ID,
			Title:     "timeline filtering test",
			Action:    gtsmodel.FilterActionHide,
			Statuses: []*gtsmodel.FilterStatus{
				{
					ID:        id.NewULID(),
					AccountID: requester.ID,
					FilterID:  filterID,
					StatusID:  filteredStatus.ID,
				},
			},
			ContextHome:          util.Ptr(true),
			ContextNotifications: util.Ptr(false),
			ContextPublic:        util.Ptr(false),
			ContextThread:        util.Ptr(false),
			ContextAccount:       util.Ptr(false),
		}
	)

	// Fetch the timeline to make sure the status we're going to filter is in that section of it.
	resp, errWithCode := suite.timeline.HomeTimelineGet(
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

	// Clear the timeline to drop all cached statuses.
	suite.state.Caches.Timelines.Home.Clear(requester.ID)

	// Create a filter to hide one status on the timeline.
	if err := suite.db.PutFilter(ctx, filter); err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch the timeline again with the filter in place.
	resp, errWithCode = suite.timeline.HomeTimelineGet(
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

func TestHomeTestSuite(t *testing.T) {
	suite.Run(t, new(HomeTestSuite))
}
