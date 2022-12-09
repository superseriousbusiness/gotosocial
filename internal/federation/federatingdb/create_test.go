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

package federatingdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type CreateTestSuite struct {
	FederatingDBTestSuite
}

func (suite *CreateTestSuite) TestCreateNote() {
	receivingAccount := suite.testAccounts["local_account_1"]
	requestingAccount := suite.testAccounts["remote_account_1"]

	ctx := createTestContext(receivingAccount, requestingAccount)

	create := suite.testActivities["dm_for_zork"].Activity

	err := suite.federatingDB.Create(ctx, create)
	suite.NoError(err)

	// should be a message heading to the processor now, which we can intercept here
	msg := <-suite.fromFederator
	suite.Equal(ap.ObjectNote, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)

	// shiny new status should be defined on the message
	suite.NotNil(msg.GTSModel)
	status := msg.GTSModel.(*gtsmodel.Status)

	// status should have some expected values
	suite.Equal(requestingAccount.ID, status.AccountID)
	suite.Equal("hey zork here's a new private note for you", status.Content)

	// status should be in the database
	_, err = suite.db.GetStatusByID(context.Background(), status.ID)
	suite.NoError(err)
}

func (suite *CreateTestSuite) TestCreateNoteForward() {
	receivingAccount := suite.testAccounts["local_account_1"]
	requestingAccount := suite.testAccounts["remote_account_1"]

	ctx := createTestContext(receivingAccount, requestingAccount)

	create := suite.testActivities["forwarded_message"].Activity

	err := suite.federatingDB.Create(ctx, create)
	suite.NoError(err)

	// should be a message heading to the processor now, which we can intercept here
	msg := <-suite.fromFederator
	suite.Equal(ap.ObjectNote, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)

	// nothing should be set as the model since this is a forward
	suite.Nil(msg.GTSModel)

	// but we should have a uri set
	suite.Equal("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1", msg.APIri.String())
}

func TestCreateTestSuite(t *testing.T) {
	suite.Run(t, &CreateTestSuite{})
}
