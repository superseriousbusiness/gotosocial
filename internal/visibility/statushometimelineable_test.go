/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package visibility_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusStatusHometimelineableTestSuite struct {
	FilterStandardTestSuite
}

func (suite *StatusStatusHometimelineableTestSuite) TestOwnStatusHometimelineable() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHometimelineable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHometimelineableTestSuite) TestFollowingStatusHometimelineable() {
	testStatus := suite.testStatuses["local_account_2_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHometimelineable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHometimelineableTestSuite) TestNotFollowingStatusHometimelineable() {
	testStatus := suite.testStatuses["remote_account_1_status_1"]
	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHometimelineable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.False(timelineable)
}

func (suite *StatusStatusHometimelineableTestSuite) TestStatusTooNewNotTimelineable() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["local_account_1_status_1"]

	var err error
	testStatus.ID, err = id.NewULIDFromTime(time.Now().Add(10 * time.Minute))
	if err != nil {
		suite.FailNow(err.Error())
	}

	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHometimelineable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.False(timelineable)
}

func (suite *StatusStatusHometimelineableTestSuite) TestStatusNotTooNewTimelineable() {
	testStatus := &gtsmodel.Status{}
	*testStatus = *suite.testStatuses["local_account_1_status_1"]

	var err error
	testStatus.ID, err = id.NewULIDFromTime(time.Now().Add(4 * time.Minute))
	if err != nil {
		suite.FailNow(err.Error())
	}

	testAccount := suite.testAccounts["local_account_1"]
	ctx := context.Background()

	timelineable, err := suite.filter.StatusHometimelineable(ctx, testStatus, testAccount)
	suite.NoError(err)

	suite.True(timelineable)
}

func (suite *StatusStatusHometimelineableTestSuite) TestChainReplyFollowersOnly() {
	ctx := context.Background()

	// This scenario makes sure that we don't timeline a status which is a followers-only
	// reply to a followers-only status TO A FOLLOWERS-ONLY STATUS owned by someone the
	// timeline owner account doesn't follow.
	//
	// In other words, remote_account_1 posts a followers-only status, which local_account_1 replies to;
	// THEN, local_account_1 replies to their own reply. We don't want this last status to appear
	// in the timeline of local_account_2, even though they follow local_account_1, because they
	// *don't* follow remote_account_1.
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
		Local:                    testrig.FalseBool(),
		AccountURI:               "http://fossbros-anonymous.io/users/foss_satan",
		AccountID:                originalStatusParent.ID,
		InReplyToID:              "",
		InReplyToAccountID:       "",
		InReplyToURI:             "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, originalStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	originalStatusTimelineable, err := suite.filter.StatusHometimelineable(ctx, originalStatus, timelineOwnerAccount)
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
		Local:                    testrig.FalseBool(),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              originalStatus.ID,
		InReplyToAccountID:       originalStatusParent.ID,
		InReplyToURI:             originalStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, firstReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	firstReplyStatusTimelineable, err := suite.filter.StatusHometimelineable(ctx, firstReplyStatus, timelineOwnerAccount)
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
		Local:                    testrig.FalseBool(),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              firstReplyStatus.ID,
		InReplyToAccountID:       replyingAccount.ID,
		InReplyToURI:             firstReplyStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, secondReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// this status should ALSO not be hometimelineable for local_account_2
	secondReplyStatusTimelineable, err := suite.filter.StatusHometimelineable(ctx, secondReplyStatus, timelineOwnerAccount)
	suite.NoError(err)
	suite.False(secondReplyStatusTimelineable)
}

func (suite *StatusStatusHometimelineableTestSuite) TestChainReplyPublicAndUnlocked() {
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
		Local:                    testrig.FalseBool(),
		AccountURI:               "http://fossbros-anonymous.io/users/foss_satan",
		AccountID:                originalStatusParent.ID,
		InReplyToID:              "",
		InReplyToAccountID:       "",
		InReplyToURI:             "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityUnlocked,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, originalStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	originalStatusTimelineable, err := suite.filter.StatusHometimelineable(ctx, originalStatus, timelineOwnerAccount)
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
		Local:                    testrig.FalseBool(),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              originalStatus.ID,
		InReplyToAccountID:       originalStatusParent.ID,
		InReplyToURI:             originalStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityPublic,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, firstReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}
	// this status should not be hometimelineable for local_account_2
	firstReplyStatusTimelineable, err := suite.filter.StatusHometimelineable(ctx, firstReplyStatus, timelineOwnerAccount)
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
		Local:                    testrig.FalseBool(),
		AccountURI:               "http://localhost:8080/users/the_mighty_zork",
		AccountID:                replyingAccount.ID,
		InReplyToID:              firstReplyStatus.ID,
		InReplyToAccountID:       replyingAccount.ID,
		InReplyToURI:             firstReplyStatus.URI,
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityUnlocked,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "",
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
	if err := suite.db.PutStatus(ctx, secondReplyStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// this status should ALSO not be hometimelineable for local_account_2
	secondReplyStatusTimelineable, err := suite.filter.StatusHometimelineable(ctx, secondReplyStatus, timelineOwnerAccount)
	suite.NoError(err)
	suite.False(secondReplyStatusTimelineable)
}

func TestStatusHometimelineableTestSuite(t *testing.T) {
	suite.Run(t, new(StatusStatusHometimelineableTestSuite))
}
