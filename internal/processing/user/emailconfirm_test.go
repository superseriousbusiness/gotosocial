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
	emailShould := fmt.Sprintf("To: some.email@example.org\r\nSubject: GoToSocial Email Confirmation\r\n\r\nHello the_mighty_zork!\r\n\r\nYou are receiving this mail because you've requested an account on http://localhost:8080.\r\n\r\nWe just need to confirm that this is your email address. To confirm your email, paste the following in your browser's address bar:\r\n\r\nhttp://localhost:8080/confirm_email?token=%s\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of http://localhost:8080\r\n\r\n", token)
	suite.Equal(emailShould, email)

	// confirmationSentAt should be recent
	suite.WithinDuration(time.Now(), user.ConfirmationSentAt, 1*time.Minute)
}

func (suite *EmailConfirmTestSuite) TestConfirmEmail() {
	ctx := context.Background()

	user := suite.testUsers["local_account_1"]

	// set a bunch of stuff on the user as though zork hasn't been confirmed yet, but has had an email sent 5 minutes ago
	updatingColumns := []string{"unconfirmed_email", "email", "confirmed_at", "confirmation_sent_at", "confirmation_token"}
	user.UnconfirmedEmail = "some.email@example.org"
	user.Email = ""
	user.ConfirmedAt = time.Time{}
	user.ConfirmationSentAt = time.Now().Add(-5 * time.Minute)
	user.ConfirmationToken = "1d1aa44b-afa4-49c8-ac4b-eceb61715cc6"

	err := suite.db.UpdateByID(ctx, user, user.ID, updatingColumns...)
	suite.NoError(err)

	// confirm with the token set above
	updatedUser, errWithCode := suite.user.ConfirmEmail(ctx, "1d1aa44b-afa4-49c8-ac4b-eceb61715cc6")
	suite.NoError(errWithCode)

	// email should now be confirmed and token cleared
	suite.Equal("some.email@example.org", updatedUser.Email)
	suite.Empty(updatedUser.UnconfirmedEmail)
	suite.Empty(updatedUser.ConfirmationToken)
	suite.WithinDuration(updatedUser.ConfirmedAt, time.Now(), 1*time.Minute)
	suite.WithinDuration(updatedUser.UpdatedAt, time.Now(), 1*time.Minute)
}

func (suite *EmailConfirmTestSuite) TestConfirmEmailOldToken() {
	ctx := context.Background()

	user := suite.testUsers["local_account_1"]

	// set a bunch of stuff on the user as though zork hasn't been confirmed yet, but has had an email sent 8 days ago
	updatingColumns := []string{"unconfirmed_email", "email", "confirmed_at", "confirmation_sent_at", "confirmation_token"}
	user.UnconfirmedEmail = "some.email@example.org"
	user.Email = ""
	user.ConfirmedAt = time.Time{}
	user.ConfirmationSentAt = time.Now().Add(-192 * time.Hour)
	user.ConfirmationToken = "1d1aa44b-afa4-49c8-ac4b-eceb61715cc6"

	err := suite.db.UpdateByID(ctx, user, user.ID, updatingColumns...)
	suite.NoError(err)

	// confirm with the token set above
	updatedUser, errWithCode := suite.user.ConfirmEmail(ctx, "1d1aa44b-afa4-49c8-ac4b-eceb61715cc6")
	suite.Nil(updatedUser)
	suite.EqualError(errWithCode, "ConfirmEmail: confirmation token expired")
}

func TestEmailConfirmTestSuite(t *testing.T) {
	suite.Run(t, &EmailConfirmTestSuite{})
}
