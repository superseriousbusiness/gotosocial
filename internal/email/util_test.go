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

package email_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/email"
)

type UtilTestSuite struct {
	EmailTestSuite
}

func (suite *UtilTestSuite) TestTemplateConfirm() {
	confirmData := email.ConfirmData{
		Username:     "test",
		InstanceURL:  "https://example.org",
		InstanceName: "Test Instance",
		ConfirmLink:  "https://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa",
	}

	suite.sender.SendConfirmEmail("user@example.org", confirmData)
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nSubject: GoToSocial Email Confirmation\r\n\r\nHello test!\r\n\r\nYou are receiving this mail because you've requested an account on https://example.org.\r\n\r\nWe just need to confirm that this is your email address. To confirm your email, paste the following in your browser's address bar:\r\n\r\nhttps://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *UtilTestSuite) TestTemplateReset() {
	resetData := email.ResetData{
		Username:     "test",
		InstanceURL:  "https://example.org",
		InstanceName: "Test Instance",
		ResetLink:    "https://example.org/reset_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa",
	}

	suite.sender.SendResetEmail("user@example.org", resetData)
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nSubject: GoToSocial Password Reset\r\n\r\nHello test!\r\n\r\nYou are receiving this mail because a password reset has been requested for your account on https://example.org.\r\n\r\nTo reset your password, paste the following in your browser's address bar:\r\n\r\nhttps://example.org/reset_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org.\r\n\r\n", suite.sentEmails["user@example.org"])
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, &UtilTestSuite{})
}
