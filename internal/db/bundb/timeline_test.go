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

package bundb_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type TimelineTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *TimelineTestSuite) publicCount() int {
	var publicCount int
	for _, status := range suite.testStatuses {
		if status.Visibility == gtsmodel.VisibilityPublic &&
			status.BoostOfID == "" &&
			!util.PtrOrZero(status.PendingApproval) {
			publicCount++
		}
	}
	return publicCount
}

func (suite *TimelineTestSuite) localCount() int {
	var localCount int
	for _, status := range suite.testStatuses {
		if status.Visibility == gtsmodel.VisibilityPublic &&
			status.BoostOfID == "" &&
			!util.PtrOrZero(status.PendingApproval) &&
			util.PtrOrValue(status.Local, true) {
			localCount++
		}
	}
	return localCount
}

func (suite *TimelineTestSuite) checkStatuses(statuses []*gtsmodel.Status, maxID string, minID string, expectedOrder paging.Order, expectedLength int) {
	if l := len(statuses); l != expectedLength {
		suite.FailNowf("", "expected %d statuses in slice, got %d", expectedLength, l)
	} else if l == 0 {
		// Can't test empty slice.
		return
	}

	if expectedOrder.Ascending() {
		// Check ordering + bounds of statuses.
		lowest := statuses[0].ID
		for _, status := range statuses {
			id := status.ID

			if id >= maxID {
				suite.FailNowf("", "%s greater than maxID %s", id, maxID)
			}

			if id <= minID {
				suite.FailNowf("", "%s smaller than minID %s", id, minID)
			}

			if id < lowest {
				suite.FailNowf("", "statuses in slice were not ordered lowest -> highest ID")
			}

			lowest = id
		}
	} else {
		// Check ordering + bounds of statuses.
		highest := statuses[0].ID
		for _, status := range statuses {
			id := status.ID

			if id >= maxID {
				suite.FailNowf("", "%s greater than maxID %s", id, maxID)
			}

			if id <= minID {
				suite.FailNowf("", "%s smaller than minID %s", id, minID)
			}

			if id > highest {
				suite.FailNowf("", "statuses in slice were not ordered highest -> lowest ID")
			}

			highest = id
		}
	}
}

func (suite *TimelineTestSuite) TestGetPublicTimeline() {
	ctx := suite.T().Context()

	page := toPage("", "", "", 20)

	s, err := suite.db.GetPublicTimeline(ctx, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), suite.publicCount())
}

func (suite *TimelineTestSuite) TestGetPublicTimelineLocal() {
	ctx := suite.T().Context()

	page := toPage("", "", "", 20)

	s, err := suite.db.GetLocalTimeline(ctx, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), suite.localCount())
}

func (suite *TimelineTestSuite) TestGetHomeTimeline() {
	var (
		ctx            = suite.T().Context()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	page := toPage("", "", "", 20)

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 20)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineIgnoreExclusive() {
	var (
		ctx            = suite.T().Context()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	// local_account_1_list_1 contains both admin_account
	// and local_account_2. If we mark this list as exclusive,
	// and remove the list entry for admin account, we should
	// only get statuses from zork and turtle in the timeline.
	list := new(gtsmodel.List)
	*list = *suite.testLists["local_account_1_list_1"]
	list.Exclusive = util.Ptr(true)
	if err := suite.db.UpdateList(ctx, list, "exclusive"); err != nil {
		suite.FailNow(err.Error())
	}

	page := toPage("", "", "", 20)

	// First try with list just set to exclusive.
	// We should only get zork's own statuses.
	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 9)

	// Remove admin account from the exclusive list.
	listEntry := suite.testListEntries["local_account_1_list_1_entry_2"]
	if err := suite.db.DeleteListEntry(ctx, listEntry.ListID, listEntry.FollowID); err != nil {
		suite.FailNow(err.Error())
	}

	// Zork should only see their own
	// statuses and admin's statuses now.
	s, err = suite.db.GetHomeTimeline(ctx, viewingAccount.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 13)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineNoFollowing() {
	var (
		ctx            = suite.T().Context()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	// Remove all of viewingAccount's follows.
	follows, err := suite.state.DB.GetAccountFollows(
		gtscontext.SetBarebones(ctx),
		viewingAccount.ID,
		nil, // select all
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	for _, f := range follows {
		if err := suite.state.DB.DeleteFollowByID(ctx, f.ID); err != nil {
			suite.FailNow(err.Error())
		}
	}

	page := toPage("", "", "", 20)

	// Query should work fine; though far
	// fewer statuses will be returned ofc.
	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 9)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineBackToFront() {
	var (
		ctx            = suite.T().Context()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	page := toPage("", "", id.Lowest, 5)

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 5)
	suite.Equal("01F8MHAYFKS4KMXF8K5Y1C0KRN", s[len(s)-1].ID)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[0].ID)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineFromHighest() {
	var (
		ctx            = suite.T().Context()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	page := toPage(id.Highest, "", "", 5)

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 5)
	suite.Equal("01JDPZEZ77X1NX0TY9M10BK1HM", s[0].ID)
	suite.Equal("01HEN2RZ8BG29Y5Z9VJC73HZW7", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetListTimelineNoParams() {
	var (
		ctx  = suite.T().Context()
		list = suite.testLists["local_account_1_list_1"]
	)

	page := toPage("", "", "", 20)

	s, err := suite.db.GetListTimeline(ctx, list.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 13)
}

func (suite *TimelineTestSuite) TestGetListTimelineMaxID() {
	var (
		ctx  = suite.T().Context()
		list = suite.testLists["local_account_1_list_1"]
	)

	page := toPage(id.Highest, "", "", 5)

	s, err := suite.db.GetListTimeline(ctx, list.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 5)
	suite.Equal("01JDPZEZ77X1NX0TY9M10BK1HM", s[0].ID)
	suite.Equal("01FN3VJGFH10KR7S2PB0GFJZYG", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetListTimelineMinID() {
	var (
		ctx  = suite.T().Context()
		list = suite.testLists["local_account_1_list_1"]
	)

	page := toPage("", "", id.Lowest, 5)

	s, err := suite.db.GetListTimeline(ctx, list.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 5)
	suite.Equal("01F8MHC8VWDRBQR0N1BATDDEM5", s[len(s)-1].ID)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[0].ID)
}

func (suite *TimelineTestSuite) TestGetListTimelineMinIDPagingUp() {
	var (
		ctx  = suite.T().Context()
		list = suite.testLists["local_account_1_list_1"]
	)

	page := toPage("", "", "01F8MHC8VWDRBQR0N1BATDDEM5", 5)

	s, err := suite.db.GetListTimeline(ctx, list.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, "01F8MHC8VWDRBQR0N1BATDDEM5", page.Order(), 5)
	suite.Equal("01G20ZM733MGN8J344T4ZDDFY1", s[len(s)-1].ID)
	suite.Equal("01F8MHCP5P2NWYQ416SBA0XSEV", s[0].ID)
}

func (suite *TimelineTestSuite) TestGetTagTimelineNoParams() {
	var (
		ctx = suite.T().Context()
		tag = suite.testTags["welcome"]
	)

	page := toPage("", "", "", 1)

	s, err := suite.db.GetTagTimeline(ctx, tag.ID, page)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, page.Order(), 1)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[0].ID)
}

func TestTimelineTestSuite(t *testing.T) {
	suite.Run(t, new(TimelineTestSuite))
}

// toPage is a helper function to wrap a series of paging arguments in paging.Page{}.
func toPage(maxID, sinceID, minID string, limit int) *paging.Page {
	var pg paging.Page
	pg.Limit = limit

	if maxID != "" {
		pg.Max = paging.MaxID(maxID)
	}

	if sinceID != "" || minID != "" {
		pg.Min = paging.EitherMinID(minID, sinceID)
	}

	return &pg
}
