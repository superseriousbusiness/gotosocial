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

package account_test

import (
	"net"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type AccountDeleteTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountDeleteTestSuite) TestAccountDeleteLocal() {
	ctx := suite.T().Context()

	// Keep a reference around to the original account
	// and user, before the delete was processed.
	ogAccount := suite.testAccounts["local_account_1"]
	ogUser := suite.testUsers["local_account_1"]

	testAccount := &gtsmodel.Account{}
	*testAccount = *ogAccount

	suspensionOrigin := "01GWVP2A8J38Q2J2FDZ6TS8AQG"
	if err := suite.accountProcessor.Delete(ctx, testAccount, suspensionOrigin); err != nil {
		suite.FailNow(err.Error())
	}

	updatedAccount, err := suite.db.GetAccountByID(ctx, ogAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.WithinDuration(time.Now(), updatedAccount.UpdatedAt, 1*time.Minute)
	suite.Zero(updatedAccount.FetchedAt)
	suite.Zero(updatedAccount.AvatarMediaAttachmentID)
	suite.Zero(updatedAccount.AvatarRemoteURL)
	suite.Zero(updatedAccount.HeaderMediaAttachmentID)
	suite.Zero(updatedAccount.HeaderRemoteURL)
	suite.Zero(updatedAccount.DisplayName)
	suite.Nil(updatedAccount.EmojiIDs)
	suite.Nil(updatedAccount.Emojis)
	suite.Nil(updatedAccount.Fields)
	suite.Zero(updatedAccount.Note)
	suite.Zero(updatedAccount.NoteRaw)
	suite.Zero(updatedAccount.MemorializedAt)
	suite.Empty(updatedAccount.AlsoKnownAsURIs)
	suite.False(*updatedAccount.Discoverable)
	suite.WithinDuration(time.Now(), updatedAccount.SuspendedAt, 1*time.Minute)
	suite.Equal(suspensionOrigin, updatedAccount.SuspensionOrigin)

	updatedUser, err := suite.db.GetUserByAccountID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.WithinDuration(time.Now(), updatedUser.UpdatedAt, 1*time.Minute)
	suite.NotEqual(updatedUser.EncryptedPassword, ogUser.EncryptedPassword)
	suite.Equal(net.IPv4zero, updatedUser.SignUpIP)
	suite.Zero(updatedUser.Locale)
	suite.Zero(updatedUser.CreatedByApplicationID)
	suite.Zero(updatedUser.LastEmailedAt)
	suite.Zero(updatedUser.ConfirmationToken)
	suite.Zero(updatedUser.ConfirmationSentAt)
	suite.Zero(updatedUser.ResetPasswordToken)
	suite.Zero(updatedUser.ResetPasswordSentAt)
}

func TestAccountDeleteTestSuite(t *testing.T) {
	suite.Run(t, new(AccountDeleteTestSuite))
}
