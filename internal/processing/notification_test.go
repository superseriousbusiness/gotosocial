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

package processing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type NotificationTestSuite struct {
	ProcessingStandardTestSuite
}

// get a notification where someone has liked our status
func (suite *NotificationTestSuite) TestGetNotifications() {
	receivingAccount := suite.testAccounts["local_account_1"]
	notifsResponse, err := suite.processor.NotificationsGet(context.Background(), suite.testAutheds["local_account_1"], []string{}, 10, "", "")
	suite.NoError(err)
	suite.Len(notifsResponse.Items, 1)
	notif, ok := notifsResponse.Items[0].(*apimodel.Notification)
	if !ok {
		panic("notif in response wasn't *apimodel.Notification")
	}

	suite.NotNil(notif.Status)
	suite.NotNil(notif.Status)
	suite.NotNil(notif.Status.Account)
	suite.Equal(receivingAccount.ID, notif.Status.Account.ID)
	suite.Equal(`<http://localhost:8080/api/v1/notifications?limit=10&max_id=01F8Q0ANPTWW10DAKTX7BRPBJP>; rel="next", <http://localhost:8080/api/v1/notifications?limit=10&since_id=01F8Q0ANPTWW10DAKTX7BRPBJP>; rel="prev"`, notifsResponse.LinkHeader)
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, &NotificationTestSuite{})
}
