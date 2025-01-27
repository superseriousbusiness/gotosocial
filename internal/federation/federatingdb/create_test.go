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

package federatingdb_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type CreateTestSuite struct {
	FederatingDBTestSuite
}

func (suite *CreateTestSuite) TestCreateNote() {
	receivingAccount := suite.testAccounts["local_account_1"]
	requestingAccount := suite.testAccounts["remote_account_1"]

	ctx := createTestContext(receivingAccount, requestingAccount)

	create := suite.testActivities["dm_for_zork"].Activity
	objProp := create.GetActivityStreamsObject()
	note := objProp.At(0).GetType().(ap.Statusable)

	err := suite.federatingDB.Create(ctx, create)
	suite.NoError(err)

	// should be a message heading to the processor now, which we can intercept here
	msg, _ := suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ObjectNote, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)
	suite.Equal(note, msg.APObject)
}

func (suite *CreateTestSuite) TestCreateNoteForward() {
	receivingAccount := suite.testAccounts["local_account_1"]
	requestingAccount := suite.testAccounts["remote_account_1"]

	ctx := createTestContext(receivingAccount, requestingAccount)

	create := suite.testActivities["forwarded_message"].Activity

	// ensure a follow exists between requesting
	// and receiving account, this ensures the forward
	// will be seen as "relevant" and not get dropped.
	err := suite.db.PutFollow(ctx, &gtsmodel.Follow{
		ID:              id.NewULID(),
		URI:             "https://this.is.a.url",
		AccountID:       receivingAccount.ID,
		TargetAccountID: requestingAccount.ID,
		ShowReblogs:     util.Ptr(true),
		Notify:          util.Ptr(false),
	})
	suite.NoError(err)

	err = suite.federatingDB.Create(ctx, create)
	suite.NoError(err)

	// should be a message heading to the processor now, which we can intercept here
	msg, _ := suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ObjectNote, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)

	// nothing should be set as the model since this is a forward
	suite.Nil(msg.APObject)

	// but we should have a uri set
	suite.Equal("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1", msg.APIRI.String())
}

func (suite *CreateTestSuite) TestCreateFlag1() {
	reportedAccount := suite.testAccounts["local_account_1"]
	reportingAccount := suite.testAccounts["remote_account_1"]
	reportedStatus := suite.testStatuses["local_account_1_status_1"]

	raw := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "` + reportingAccount.URI + `",
  "content": "Note: ` + reportedStatus.URL + `\n-----\nban this sick filth ⛔",
  "id": "http://fossbros-anonymous.io/db22128d-884e-4358-9935-6a7c3940535d",
  "object": "` + reportedAccount.URI + `",
  "type": "Flag"
}`

	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		suite.FailNow(err.Error())
	}

	t, err := streams.ToType(context.Background(), m)
	if err != nil {
		suite.FailNow(err.Error())
	}

	flag := t.(vocab.ActivityStreamsFlag)

	ctx := createTestContext(reportedAccount, reportingAccount)
	if err := suite.federatingDB.Flag(ctx, flag); err != nil {
		suite.FailNow(err.Error())
	}

	// should be a message heading to the processor now, which we can intercept here
	msg, _ := suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ActivityFlag, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)

	// shiny new report should be defined on the message
	suite.NotNil(msg.GTSModel)
	report := msg.GTSModel.(*gtsmodel.Report)

	// report should be in the database
	if _, err := suite.db.GetReportByID(context.Background(), report.ID); err != nil {
		suite.FailNow(err.Error())
	}
}

func TestCreateTestSuite(t *testing.T) {
	suite.Run(t, &CreateTestSuite{})
}
