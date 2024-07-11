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

package visibility_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusStatusHomeTimelineableTestSuite struct {
	FilterStandardTestSuite
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestOwnStatusHomeTimelineable() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestFollowingStatusHomeTimelineable() {
	testStatus := suite.testStatuses["local_account_2_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestFollowingBoostedStatusHomeTimelineable() {
	ctx := context.Background()

	testStatus := suite.testStatuses["admin_account_status_4"]
	testAccount := suite.testAccounts["local_account_1"]
	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestFollowingBoostedStatusHomeTimelineableNoReblogs() {
	ctx := context.Background()

	// Update follow to indicate that local_account_1
	// doesn't want to see reblogs by admin_account.
	follow := &gtsmodel.Follow{}
	*follow = *suite.testFollows["local_account_1_admin_account"]
	follow.ShowReblogs = util.Ptr(false)

	if err := suite.db.UpdateFollow(ctx, follow, "show_reblogs"); err != nil {
		suite.FailNow(err.Error())
	}

	testStatus := suite.testStatuses["admin_account_status_4"]
	testAccount := suite.testAccounts["local_account_1"]
	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.False(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestNotFollowingStatusHomeTimelineable() {
	testStatus := suite.testStatuses["remote_account_1_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.False(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestStatusTooNewNotTimelineable() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["local_account_1_status_1"]

	testStatus.CreatedAt = time.Now().Add(25 * time.Hour)

	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.False(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestStatusNotTooNewTimelineable() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["local_account_1_status_1"]

	testStatus.CreatedAt = time.Now().Add(23 * time.Hour)

	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHomeTimelineable(ctx, testAccount, testStatus)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestThread() {
	ctx := context.Background()

	threadParentAccount := suite.testAccounts["local_account_1"]
	timelineOwnerAccount := suite.testAccounts["local_account_2"]
	originalStatus := suite.testStatuses["local_account_1_status_1"]

	// this status should be hometimelineable for local_account_2
	originalStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, originalStatus)
	suite.NoError(err)
	suite.True(originalStatusTimelineable)

	// now a reply from the original status author to their own status
	firstReplyStatus := &gtsmodel.Status{
		ID:                       "01G395ESAYPK9161QSQEZKATJN",
		URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01G395ESAYPK9161QSQEZKATJN",
		URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01G395ESAYPK9161QSQEZKATJN",
		Content:                  "nbnbdy expects dog",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:41:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:41:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                threadParentAccount.ID,
		InReplyToID:              originalStatus.ID,
		InReplyToAccountID:       threadParentAccount.ID,
		InReplyToURI:             originalStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, firstReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// this status should also be hometimelineable for local_account_2
	firstReplyStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, firstReplyStatus)
	suite.NoError(err)
	suite.True(firstReplyStatusTimelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestChainReplyFollowersOnly() {
	ctx := context.Background()

	// This scenario makes sure that we don't timeline a status which is a followers-only
	// reply to a followers-only status TO A FOLLOWERS-ONLY STATUS owned by someone the
	// timeline owner account doesn't follow.
	//
	// In other words, remote_account_1 posts a followers-only status, which local_account_1 replies to;
	// THEN, local_account_1 replies to their own reply. None of these statuses should appear to
	// local_account_2 since they don't follow the original parent.
	//
	// See: https://github.com/superseriousbusiness/gotosocial/issues/501

	originalStatusParent := suite.testAccounts["remote_account_1"]
	replyingAccount := suite.testAccounts["local_account_1"]
	timelineOwnerAccount := suite.testAccounts["local_account_2"]

	// put a followers-only status by remote_account_1 in the db
	originalStatus := &gtsmodel.Status{
		ID:                       "01G3957TS7XE2CMDKFG3MZPWAF",
		URI:                      "http://fossbros-anonymous.io/users/foss_satan/statuses/01G3957TS7XE2CMDKFG3MZPWAF",
		URL:                      "http://fossbros-anonymous.io/@foss_satan/statuses/01G3957TS7XE2CMDKFG3MZPWAF",
		Content:                  "didn't expect dog",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:40:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:40:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://fossbros-anonymous.io/users/foss_satan",
		AccountID:                originalStatusParent.ID,
		InReplyToID:              "",
		InReplyToAccountID:       "",
		InReplyToURI:             "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, originalStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	originalStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, originalStatus)
	suite.NoError(err)
	suite.False(originalStatusTimelineable)

	// now a followers-only reply from zork
	firstReplyStatus := &gtsmodel.Status{
		ID:                       "01G395ESAYPK9161QSQEZKATJN",
		URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01G395ESAYPK9161QSQEZKATJN",
		URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01G395ESAYPK9161QSQEZKATJN",
		Content:                  "nbnbdy expects dog",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:41:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:41:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              originalStatus.ID,
		InReplyToAccountID:       originalStatusParent.ID,
		InReplyToURI:             originalStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, firstReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should be hometimelineable for local_account_2
	firstReplyStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, firstReplyStatus)
	suite.NoError(err)
	suite.False(firstReplyStatusTimelineable)

	// now a followers-only reply from zork to the status they just replied to
	secondReplyStatus := &gtsmodel.Status{
		ID:                       "01G395NZQZGJYRBAES57KYZ7XP",
		URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01G395NZQZGJYRBAES57KYZ7XP",
		URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01G395NZQZGJYRBAES57KYZ7XP",
		Content:                  "*nobody",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:42:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:42:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              firstReplyStatus.ID,
		InReplyToAccountID:       replyingAccount.ID,
		InReplyToURI:             firstReplyStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, secondReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// this status should ALSO not be hometimelineable for local_account_2
	secondReplyStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, secondReplyStatus)
	suite.NoError(err)
	suite.False(secondReplyStatusTimelineable)
}

func (suite *StatusStatusHomeTimelineableTestSuite) TestChainReplyPublicAndUnlocked() {
	ctx := context.Background()

	// This scenario is exactly the same as the above test, but for a mix of unlocked + public posts

	originalStatusParent := suite.testAccounts["remote_account_1"]
	replyingAccount := suite.testAccounts["local_account_1"]
	timelineOwnerAccount := suite.testAccounts["local_account_2"]

	// put an unlocked status by remote_account_1 in the db
	originalStatus := &gtsmodel.Status{
		ID:                       "01G3957TS7XE2CMDKFG3MZPWAF",
		URI:                      "http://fossbros-anonymous.io/users/foss_satan/statuses/01G3957TS7XE2CMDKFG3MZPWAF",
		URL:                      "http://fossbros-anonymous.io/@foss_satan/statuses/01G3957TS7XE2CMDKFG3MZPWAF",
		Content:                  "didn't expect dog",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:40:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:40:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://fossbros-anonymous.io/users/foss_satan",
		AccountID:                originalStatusParent.ID,
		InReplyToID:              "",
		InReplyToAccountID:       "",
		InReplyToURI:             "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityUnlocked,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, originalStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	originalStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, originalStatus)
	suite.NoError(err)
	suite.False(originalStatusTimelineable)

	// now a public reply from zork
	firstReplyStatus := &gtsmodel.Status{
		ID:                       "01G395ESAYPK9161QSQEZKATJN",
		URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01G395ESAYPK9161QSQEZKATJN",
		URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01G395ESAYPK9161QSQEZKATJN",
		Content:                  "nbnbdy expects dog",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:41:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:41:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              originalStatus.ID,
		InReplyToAccountID:       originalStatusParent.ID,
		InReplyToURI:             originalStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityPublic,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, firstReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	firstReplyStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, firstReplyStatus)
	suite.NoError(err)
	suite.False(firstReplyStatusTimelineable)

	// now an unlocked reply from zork to the status they just replied to
	secondReplyStatus := &gtsmodel.Status{
		ID:                       "01G395NZQZGJYRBAES57KYZ7XP",
		URI:                      "http://localhost:8080/users/the_mighty_zork/statuses/01G395NZQZGJYRBAES57KYZ7XP",
		URL:                      "http://localhost:8080/@the_mighty_zork/statuses/01G395NZQZGJYRBAES57KYZ7XP",
		Content:                  "*nobody",
		CreatedAt:                testrig.TimeMustParse("2021-09-20T12:42:37+02:00"),
		UpdatedAt:                testrig.TimeMustParse("2021-09-20T12:42:37+02:00"),
		Local:                    util.Ptr(false),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              firstReplyStatus.ID,
		InReplyToAccountID:       replyingAccount.ID,
		InReplyToURI:             firstReplyStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityUnlocked,
		Sensitive:                util.Ptr(false),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                util.Ptr(true),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, secondReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// this status should ALSO not be hometimelineable for local_account_2
	secondReplyStatusTimelineable, err := suite.filter.StatusHomeTimelineable(ctx, timelineOwnerAccount, secondReplyStatus)
	suite.NoError(err)
	suite.False(secondReplyStatusTimelineable)
}

func TestStatusHomeTimelineableTestSuite(t *testing.T) {
	suite.Run(t, new(StatusStatusHomeTimelineableTestSuite))
}
