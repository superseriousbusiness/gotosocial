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

package streaming_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type NotificationTestSuite struct {
	StreamingTestSuite
}

func (suite *NotificationTestSuite) TestStreamNotification() {
	account := suite.testAccounts["local_account_1"]

	openStream, errWithCode := suite.streamingProcessor.OpenStreamForAccount(context.Background(), account, "user")
	suite.NoError(errWithCode)

	followAccount := suite.testAccounts["remote_account_1"]
	followAccountAPIModel, err := testrig.NewTestTypeConverter(suite.db).AccountToAPIAccountPublic(context.Background(), followAccount)
	suite.NoError(err)

	notification := &apimodel.Notification{
		ID:        "01FH57SJCMDWQGEAJ0X08CE3WV",
		Type:      "follow",
		CreatedAt: "2021-10-04T08:52:36.000Z",
		Account:   followAccountAPIModel,
	}

	err = suite.streamingProcessor.StreamNotificationToAccount(notification, account)
	suite.NoError(err)

	msg := <-openStream.Messages
	suite.Equal(`{"id":"01FH57SJCMDWQGEAJ0X08CE3WV","type":"follow","created_at":"2021-10-04T08:52:36.000Z","account":{"id":"01F8MH5ZK5VRH73AKHQM6Y9VNX","username":"foss_satan","acct":"foss_satan@fossbros-anonymous.io","display_name":"big gerald","locked":false,"bot":false,"created_at":"2021-09-26T10:52:36.000Z","note":"i post about like, i dunno, stuff, or whatever!!!!","url":"http://fossbros-anonymous.io/@foss_satan","avatar":"","avatar_static":"","header":"http://localhost:8080/assets/default_header.png","header_static":"http://localhost:8080/assets/default_header.png","followers_count":0,"following_count":0,"statuses_count":1,"last_status_at":"2021-09-20T10:40:37.000Z","emojis":[],"fields":[]}}`, msg.Payload)
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, &NotificationTestSuite{})
}
