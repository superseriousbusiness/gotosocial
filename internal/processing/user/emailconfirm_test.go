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

package user_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type EmailConfirmTestSuite struct {
	UserStandardTestSuite
}

func (suite *EmailConfirmTestSuite) TestSendConfirmEmail() {
	user := suite.testUsers["local_account_1"]

	// set a bunch of stuff on the user as though zork hasn't been confirmed (perish the thought)
	user.UnconfirmedEmail = "some.email@example.org"
	user.Email = ""
	user.ConfirmedAt = time.Time{}
	user.ConfirmationSentAt = time.Time{}
	user.ConfirmationToken = ""

	err := suite.user.SendConfirmEmail(context.Background(), user, "the_mighty_zork")
	suite.NoError(err)

	// zork should have an email now
	suite.Len(suite.sentEmails, 1)
	email, ok := suite.sentEmails["some.email@example.org"]
	suite.True(ok)

	// a token should be set on zork
	token := user.ConfirmationToken
	suite.NotEmpty(token)

	// email should contain the token
	emailShould := fmt.Sprintf("Subject: GoToSocial Email Confirmation\r\nFrom: GoToSocial <test@example.org>\r\nTo: some.email@example.org\r\nMIME-version: 1.0;\nContent-Type: text/html;\r\n<!doctype html><html></head><body><div><h1>Hello the_mighty_zork!</h1></div><div><p>You are receiving this mail because you've requested an account on <a href=\"http://localhost:8080\">localhost:8080</a>.</p><p>We just need to confirm that this is your email address. To confirm your email, <a href=\"http://localhost:8080/confirm_email?token=%s\">click here</a> or paste the following in your browser's address bar:</p><p><code>http://localhost:8080/confirm_email?token=%s</code></p></div><div><p>If you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of <a href=\"http://localhost:8080\">localhost:8080</a>.</p></div></body></html>\r\n", token, token)
	suite.Equal(emailShould, email)

	// confirmationSentAt should be recent
	suite.WithinDuration(time.Now(), user.ConfirmationSentAt, 1*time.Minute)
}

func TestEmailConfirmTestSuite(t *testing.T) {
	suite.Run(t, &EmailConfirmTestSuite{})
}
