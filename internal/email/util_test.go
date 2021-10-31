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
	suite.Equal("Subject: GoToSocial Email Confirmation\r\nFrom: GoToSocial <test@example.org>\r\nTo: user@example.org\r\nMIME-version: 1.0;\nContent-Type: text/html;\r\n<!DOCTYPE html>\n<html>\n    </head>\n    <body>\n        <div>\n            <h1>\n                Hello test!\n            </h1>\n        </div>\n        <div>\n            <p>\n                You are receiving this mail because you've requested an account on <a href=\"https://example.org\">Test Instance</a>.\n            </p>\n            <p>\n                We just need to confirm that this is your email address. To confirm your email, <a href=\"https://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\">click here</a> or paste the following in your browser's address bar:\n            </p>\n            <p>\n                <code>\n                    https://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\n                </code>\n            </p>\n        </div>\n        <div>\n            <p>\n                If you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of <a href=\"https://example.org\">Test Instance</a>.\n            </p>\n        </div>\n    </body>\n</html>\r\n", suite.sentEmails["user@example.org"])
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
	suite.Equal("Subject: GoToSocial Password Reset\r\nFrom: GoToSocial <test@example.org>\r\nTo: user@example.org\r\nMIME-version: 1.0;\nContent-Type: text/html;\r\n<!DOCTYPE html>\n<html>\n    </head>\n    <body>\n        <div>\n            <h1>\n                Hello test!\n            </h1>\n        </div>\n        <div>\n            <p>\n                You are receiving this mail because a password reset has been requested for your account on <a href=\"https://example.org\">Test Instance</a>.\n            </p>\n            <p>\n                To reset your password, <a href=\"https://example.org/reset_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\">click here</a> or paste the following in your browser's address bar:\n            </p>\n            <p>\n                <code>\n                    https://example.org/reset_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\n                </code>\n            </p>\n        </div>\n        <div>\n            <p>\n                If you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of <a href=\"https://example.org\">Test Instance</a>.\n            </p>\n        </div>\n    </body>\n</html>\r\n", suite.sentEmails["user@example.org"])
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, &UtilTestSuite{})
}
