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

package validate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func happyNotification() *gtsmodel.Notification {
	return &gtsmodel.Notification{
		ID:               "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:        time.Now(),
		NotificationType: gtsmodel.NotificationFave,
		OriginAccountID:  "01FE96MAE58MXCE5C4SSMEMCEK",
		OriginAccount:    nil,
		TargetAccountID:  "01FE96MXRHWZHKC0WH5FT82H1A",
		TargetAccount:    nil,
		StatusID:         "01FE96NBPNJNY26730FT6GZTFE",
		Status:           nil,
	}
}

type NotificationValidateTestSuite struct {
	suite.Suite
}

func (suite *NotificationValidateTestSuite) TestValidateNotificationHappyPath() {
	// no problem here
	n := happyNotification()
	err := validate.Struct(n)
	suite.NoError(err)
}

func (suite *NotificationValidateTestSuite) TestValidateNotificationBadID() {
	n := happyNotification()

	n.ID = ""
	err := validate.Struct(n)
	suite.EqualError(err, "Key: 'Notification.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	n.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(n)
	suite.EqualError(err, "Key: 'Notification.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *NotificationValidateTestSuite) TestValidateNotificationStatusID() {
	n := happyNotification()

	n.StatusID = ""
	err := validate.Struct(n)
	suite.EqualError(err, "Key: 'Notification.StatusID' Error:Field validation for 'StatusID' failed on the 'required_if' tag")

	n.StatusID = "9HZJ76B6VXSKF"
	err = validate.Struct(n)
	suite.EqualError(err, "Key: 'Notification.StatusID' Error:Field validation for 'StatusID' failed on the 'ulid' tag")

	n.StatusID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!!!!!!!!!!!!"
	err = validate.Struct(n)
	suite.EqualError(err, "Key: 'Notification.StatusID' Error:Field validation for 'StatusID' failed on the 'ulid' tag")

	n.StatusID = ""
	n.NotificationType = gtsmodel.NotificationFollowRequest
	err = validate.Struct(n)
	suite.NoError(err)
}

func (suite *NotificationValidateTestSuite) TestValidateNotificationNoCreatedAt() {
	n := happyNotification()

	n.CreatedAt = time.Time{}
	err := validate.Struct(n)
	suite.NoError(err)
}

func TestNotificationValidateTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationValidateTestSuite))
}
