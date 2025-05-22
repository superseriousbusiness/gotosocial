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
	"encoding/json"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"github.com/stretchr/testify/suite"
)

type UpdateTestSuite struct {
	FederatingDBTestSuite
}

func (suite *UpdateTestSuite) TestUpdateNewMention() {
	var (
		ctx            = suite.T().Context()
		update         = suite.testActivities["remote_account_2_status_1_update"]
		receivingAcct  = suite.testAccounts["local_account_1"]
		requestingAcct = suite.testAccounts["remote_account_2"]
	)

	ctx = gtscontext.SetReceivingAccount(ctx, receivingAcct)
	ctx = gtscontext.SetRequestingAccount(ctx, requestingAcct)

	m, err := ap.Serialize(update.Activity)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.T().Logf("Update:\n%s\n", string(b))

	note := update.Activity.GetActivityStreamsObject().At(0).GetActivityStreamsNote()
	if err := suite.federatingDB.Update(ctx, note); err != nil {
		suite.FailNow(err.Error())
	}

	// Should be a message heading to the processor.
	msg, ok := suite.getFederatorMsg(5 * time.Second)
	if !ok {
		suite.FailNow("no federator message after 5s")
	}

	suite.Equal(ap.ObjectNote, msg.APObjectType)
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.NotNil(msg.APObject)
}

func TestUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateTestSuite))
}
