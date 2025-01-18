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
	"context"
	"testing"
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type TimelineTestSuite struct {
	BunDBStandardTestSuite
}

func getFutureStatus() *gtsmodel.Status {
	theDistantFuture := time.Now().Add(876600 * time.Hour)
	id := id.NewULIDFromTime(theDistantFuture)

	return &gtsmodel.Status{
		ID:                       id,
		URI:                      "http://localhost:8080/users/admin/statuses/" + id,
		URL:                      "http://localhost:8080/@admin/statuses/" + id,
		Content:                  "it's the future, wooooooooooooooooooooooooooooooooo",
		Text:                     "it's the future, wooooooooooooooooooooooooooooooooo",
		AttachmentIDs:            []string{},
		TagIDs:                   []string{},
		MentionIDs:               []string{},
		EmojiIDs:                 []string{},
		CreatedAt:                theDistantFuture,
		Local:                    util.Ptr(true),
		AccountURI:               "http://localhost:8080/users/admin",
		AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
		InReplyToID:              "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityPublic,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
		Federated:                util.Ptr(true),
		InteractionPolicy:        gtsmodel.DefaultInteractionPolicyPublic(),
		ActivityStreamsType:      ap.ObjectNote,
	}
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

func (suite *TimelineTestSuite) checkStatuses(statuses []*gtsmodel.Status, maxID string, minID string, expectedLength int) {
	if l := len(statuses); l != expectedLength {
		suite.FailNowf("", "expected %d statuses in slice, got %d", expectedLength, l)
	} else if l == 0 {
		// Can't test empty slice.
		return
	}

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

func (suite *TimelineTestSuite) TestGetPublicTimeline() {
	ctx := context.Background()

	s, err := suite.db.GetPublicTimeline(ctx, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.T().Log(kv.Field{
		K: "statuses", V: s,
	})

	suite.checkStatuses(s, id.Highest, id.Lowest, suite.publicCount())
}

func (suite *TimelineTestSuite) TestGetPublicTimelineWithFutureStatus() {
	ctx := context.Background()

	// Insert a status set far in the
	// future, it shouldn't be retrieved.
	futureStatus := getFutureStatus()
	if err := suite.db.PutStatus(ctx, futureStatus); err != nil {
		suite.FailNow(err.Error())
	}

	s, err := suite.db.GetPublicTimeline(ctx, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotContains(s, futureStatus)
	suite.checkStatuses(s, id.Highest, id.Lowest, suite.publicCount())
}

func (suite *TimelineTestSuite) TestGetHomeTimeline() {
	var (
		ctx            = context.Background()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 20)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineIgnoreExclusive() {
	var (
		ctx            = context.Background()
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

	// First try with list just set to exclusive.
	// We should only get zork's own statuses.
	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(s, id.Highest, id.Lowest, 9)

	// Remove admin account from the exclusive list.
	listEntry := suite.testListEntries["local_account_1_list_1_entry_2"]
	if err := suite.db.DeleteListEntry(ctx, listEntry.ListID, listEntry.FollowID); err != nil {
		suite.FailNow(err.Error())
	}

	// Zork should only see their own
	// statuses and admin's statuses now.
	s, err = suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.checkStatuses(s, id.Highest, id.Lowest, 13)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineNoFollowing() {
	var (
		ctx            = context.Background()
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

	// Query should work fine; though far
	// fewer statuses will be returned ofc.
	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 9)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineWithFutureStatus() {
	var (
		ctx            = context.Background()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	// Insert a status set far in the
	// future, it shouldn't be retrieved.
	futureStatus := getFutureStatus()
	if err := suite.db.PutStatus(ctx, futureStatus); err != nil {
		suite.FailNow(err.Error())
	}

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", "", 20, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotContains(s, futureStatus)
	suite.checkStatuses(s, id.Highest, id.Lowest, 20)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineBackToFront() {
	var (
		ctx            = context.Background()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", id.Lowest, 5, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 5)
	suite.Equal("01F8MHAYFKS4KMXF8K5Y1C0KRN", s[0].ID)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineFromHighest() {
	var (
		ctx            = context.Background()
		viewingAccount = suite.testAccounts["local_account_1"]
	)

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, id.Highest, "", "", 5, false)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 5)
	suite.Equal("01JDPZEZ77X1NX0TY9M10BK1HM", s[0].ID)
	suite.Equal("01HEN2RZ8BG29Y5Z9VJC73HZW7", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetListTimelineNoParams() {
	var (
		ctx  = context.Background()
		list = suite.testLists["local_account_1_list_1"]
	)

	s, err := suite.db.GetListTimeline(ctx, list.ID, "", "", "", 20)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 13)
}

func (suite *TimelineTestSuite) TestGetListTimelineMaxID() {
	var (
		ctx  = context.Background()
		list = suite.testLists["local_account_1_list_1"]
	)

	s, err := suite.db.GetListTimeline(ctx, list.ID, id.Highest, "", "", 5)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 5)
	suite.Equal("01JDPZEZ77X1NX0TY9M10BK1HM", s[0].ID)
	suite.Equal("01FN3VJGFH10KR7S2PB0GFJZYG", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetListTimelineMinID() {
	var (
		ctx  = context.Background()
		list = suite.testLists["local_account_1_list_1"]
	)

	s, err := suite.db.GetListTimeline(ctx, list.ID, "", "", id.Lowest, 5)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 5)
	suite.Equal("01F8MHC8VWDRBQR0N1BATDDEM5", s[0].ID)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetListTimelineMinIDPagingUp() {
	var (
		ctx  = context.Background()
		list = suite.testLists["local_account_1_list_1"]
	)

	s, err := suite.db.GetListTimeline(ctx, list.ID, "", "", "01F8MHC8VWDRBQR0N1BATDDEM5", 5)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, "01F8MHC8VWDRBQR0N1BATDDEM5", 5)
	suite.Equal("01G20ZM733MGN8J344T4ZDDFY1", s[0].ID)
	suite.Equal("01F8MHCP5P2NWYQ416SBA0XSEV", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetTagTimelineNoParams() {
	var (
		ctx = context.Background()
		tag = suite.testTags["welcome"]
	)

	s, err := suite.db.GetTagTimeline(ctx, tag.ID, "", "", "", 1)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.checkStatuses(s, id.Highest, id.Lowest, 1)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[0].ID)
}

func TestTimelineTestSuite(t *testing.T) {
	suite.Run(t, new(TimelineTestSuite))
}
