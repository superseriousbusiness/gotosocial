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

package processing_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FromClientAPITestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *FromClientAPITestSuite) TestProcessStreamNewStatus() {
	ctx := context.Background()

	// let's say that the admin account posts a new status: it should end up in the
	// timeline of any account that follows it and has a stream open
	postingAccount := suite.testAccounts["admin_account"]
	receivingAccount := suite.testAccounts["local_account_1"]

	// open a home timeline stream for zork
	wssStream, errWithCode := suite.processor.OpenStreamForAccount(ctx, receivingAccount, stream.TimelineHome)
	suite.NoError(errWithCode)

	// open another stream for zork, but for a different timeline;
	// this shouldn't get stuff streamed into it, since it's for the public timeline
	irrelevantStream, errWithCode := suite.processor.OpenStreamForAccount(ctx, receivingAccount, stream.TimelinePublic)
	suite.NoError(errWithCode)

	// make a new status from admin account
	newStatus := &gtsmodel.Status{
		ID:                       "01FN4B2F88TF9676DYNXWE1WSS",
		URI:                      "http://localhost:8080/users/admin/statuses/01FN4B2F88TF9676DYNXWE1WSS",
		URL:                      "http://localhost:8080/@admin/statuses/01FN4B2F88TF9676DYNXWE1WSS",
		Content:                  "this status should stream :)",
		AttachmentIDs:            []string{},
		TagIDs:                   []string{},
		MentionIDs:               []string{},
		EmojiIDs:                 []string{},
		CreatedAt:                testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		UpdatedAt:                testrig.TimeMustParse("2021-10-20T11:36:45Z"),
		Local:                    true,
		AccountURI:               "http://localhost:8080/users/admin",
		AccountID:                "01F8MH17FWEB39HZJ76B6VXSKF",
		InReplyToID:              "",
		BoostOfID:                "",
		ContentWarning:           "",
		Visibility:               gtsmodel.VisibilityFollowersOnly,
		Sensitive:                false,
		Language:                 "en",
		CreatedWithApplicationID: "01F8MGXQRHYF5QPMTMXP78QC2F",
		Federated:                false, // set federated as false for this one, since we're not testing federation stuff now
		Boostable:                true,
		Replyable:                true,
		Likeable:                 true,
		ActivityStreamsType:      ap.ObjectNote,
	}

	// put the status in the db first, to mimic what would have already happened earlier up the flow
	err := suite.db.PutStatus(ctx, newStatus)
	suite.NoError(err)

	// process the new status
	err = suite.processor.ProcessFromClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityCreate,
		GTSModel:       newStatus,
		OriginAccount:  postingAccount,
	})
	suite.NoError(err)

	// zork's stream should have the newly created status in it now
	msg := <-wssStream.Messages
	suite.Equal(stream.EventTypeUpdate, msg.Event)
	suite.NotEmpty(msg.Payload)
	suite.EqualValues([]string{stream.TimelineHome}, msg.Stream)
	statusStreamed := &model.Status{}
	err = json.Unmarshal([]byte(msg.Payload), statusStreamed)
	suite.NoError(err)
	suite.Equal("01FN4B2F88TF9676DYNXWE1WSS", statusStreamed.ID)
	suite.Equal("this status should stream :)", statusStreamed.Content)

	// and stream should now be empty
	suite.Empty(wssStream.Messages)

	// the irrelevant messages stream should also be empty
	suite.Empty(irrelevantStream.Messages)
}

func TestFromClientAPITestSuite(t *testing.T) {
	suite.Run(t, &FromClientAPITestSuite{})
}
