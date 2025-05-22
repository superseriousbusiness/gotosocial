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

package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type EmailConfirmTestSuite struct {
	UserStandardTestSuite
}

func (suite *EmailConfirmTestSuite) TestConfirmEmail() {
	ctx := suite.T().Context()

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
	updatedUser, errWithCode := suite.user.EmailConfirm(ctx, "1d1aa44b-afa4-49c8-ac4b-eceb61715cc6")
	suite.NoError(errWithCode)

	// email should now be confirmed and token cleared
	suite.Equal("some.email@example.org", updatedUser.Email)
	suite.Empty(updatedUser.UnconfirmedEmail)
	suite.Empty(updatedUser.ConfirmationToken)
	suite.WithinDuration(updatedUser.ConfirmedAt, time.Now(), 1*time.Minute)
	suite.WithinDuration(updatedUser.UpdatedAt, time.Now(), 1*time.Minute)
}

func (suite *EmailConfirmTestSuite) TestConfirmEmailOldToken() {
	ctx := suite.T().Context()

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
	updatedUser, errWithCode := suite.user.EmailConfirm(ctx, "1d1aa44b-afa4-49c8-ac4b-eceb61715cc6")
	suite.Nil(updatedUser)
	suite.EqualError(errWithCode, "confirmation token expired (older than one week)")
}

func TestEmailConfirmTestSuite(t *testing.T) {
	suite.Run(t, &EmailConfirmTestSuite{})
}
