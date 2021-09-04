/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"testing"

	"github.com/stretchr/testify/suite"
)

type NotificationTestSuite struct {
	ProcessingStandardTestSuite
}

// get a notification where someone has liked our status
func (suite *NotificationTestSuite) TestGetNotifications() {
	receivingAccount := suite.testAccounts["local_account_1"]
	notifs, err := suite.processor.NotificationsGet(context.Background(), suite.testAutheds["local_account_1"], 10, "", "")
	suite.NoError(err)
	suite.Len(notifs, 1)
	notif := notifs[0]
	suite.NotNil(notif.Status)
	suite.NotNil(notif.Status)
	suite.NotNil(notif.Status.Account)
	suite.Equal(receivingAccount.ID, notif.Status.Account.ID)
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, &NotificationTestSuite{})
}
