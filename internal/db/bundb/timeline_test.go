/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package bundb_test

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

type TimelineTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *TimelineTestSuite) TestGetPublicTimeline() {
	ctx := context.Background()

	s, err := suite.db.GetPublicTimeline(ctx, "", "", "", 20, false)
	suite.NoError(err)

	suite.Len(s, 6)
}

func (suite *TimelineTestSuite) TestGetPublicTimelineWithFutureStatus() {
	ctx := context.Background()

	futureStatus := getFutureStatus()
	err := suite.db.PutStatus(ctx, futureStatus)
	suite.NoError(err)

	s, err := suite.db.GetPublicTimeline(ctx, "", "", "", 20, false)
	suite.NoError(err)

	suite.NotContains(s, futureStatus)
	suite.Len(s, 6)
}

func (suite *TimelineTestSuite) TestGetHomeTimeline() {
	ctx := context.Background()

	viewingAccount := suite.testAccounts["local_account_1"]

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", "", 20, false)
	suite.NoError(err)

	suite.Len(s, 16)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineWithFutureStatus() {
	ctx := context.Background()

	viewingAccount := suite.testAccounts["local_account_1"]

	futureStatus := getFutureStatus()
	err := suite.db.PutStatus(ctx, futureStatus)
	suite.NoError(err)

	s, err := suite.db.GetHomeTimeline(context.Background(), viewingAccount.ID, "", "", "", 20, false)
	suite.NoError(err)

	suite.NotContains(s, futureStatus)
	suite.Len(s, 16)
}

func getFutureStatus() *gtsmodel.Status {
	theDistantFuture := time.Now().Add(876600 * time.Hour)
	id, err := id.NewULIDFromTime(theDistantFuture)
	if err != nil {
		panic(err)
	}

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
		UpdatedAt:                theDistantFuture,
		Local:                    testrig.TrueBool(),
		AccountURI:               "http://localhost:8080/users/admin",
		AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
		InReplyToID:              "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityPublic,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
		Pinned:                   testrig.FalseBool(),
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
	}
}

func TestTimelineTestSuite(t *testing.T) {
	suite.Run(t, new(TimelineTestSuite))
}
