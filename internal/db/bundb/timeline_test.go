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
	var count int

	for _, status := range suite.testStatuses {
		if status.Visibility == gtsmodel.VisibilityPublic &&
			status.BoostOfID == "" {
			count++
		}
	}

	ctx := context.Background()
	s, err := suite.db.GetPublicTimeline(ctx, "", "", "", 20, false)
	suite.NoError(err)

	suite.Len(s, count)
}

func (suite *TimelineTestSuite) TestGetPublicTimelineWithFutureStatus() {
	var count int

	for _, status := range suite.testStatuses {
		if status.Visibility == gtsmodel.VisibilityPublic &&
			status.BoostOfID == "" {
			count++
		}
	}

	ctx := context.Background()

	futureStatus := getFutureStatus()
	err := suite.db.PutStatus(ctx, futureStatus)
	suite.NoError(err)

	s, err := suite.db.GetPublicTimeline(ctx, "", "", "", 20, false)
	suite.NoError(err)

	suite.NotContains(s, futureStatus)
	suite.Len(s, count)
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

func (suite *TimelineTestSuite) TestGetHomeTimelineBackToFront() {
	ctx := context.Background()

	viewingAccount := suite.testAccounts["local_account_1"]

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, "", "", id.Lowest, 5, false)
	suite.NoError(err)

	suite.Len(s, 5)
	suite.Equal("01F8MHAYFKS4KMXF8K5Y1C0KRN", s[0].ID)
	suite.Equal("01F8MH75CBF9JFX4ZAD54N0W0R", s[len(s)-1].ID)
}

func (suite *TimelineTestSuite) TestGetHomeTimelineFromHighest() {
	ctx := context.Background()

	viewingAccount := suite.testAccounts["local_account_1"]

	s, err := suite.db.GetHomeTimeline(ctx, viewingAccount.ID, id.Highest, "", "", 5, false)
	suite.NoError(err)

	suite.Len(s, 5)
	suite.Equal("01G36SF3V6Y6V5BF9P4R7PQG7G", s[0].ID)
	suite.Equal("01FCTA44PW9H1TB328S9AQXKDS", s[len(s)-1].ID)
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
