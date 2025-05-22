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

package admin_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type AdminRejectTestSuite struct {
	AdminStandardTestSuite
}

func (suite *AdminRejectTestSuite) TestReject() {
	var (
		ctx            = suite.T().Context()
		adminAcct      = suite.testAccounts["admin_account"]
		targetAcct     = suite.testAccounts["unconfirmed_account"]
		targetUser     = suite.testUsers["unconfirmed_account"]
		privateComment = "It's a no from me chief."
		sendEmail      = true
		message        = "Too stinky."
	)

	acct, errWithCode := suite.adminProcessor.SignupReject(
		ctx,
		adminAcct,
		targetAcct.ID,
		privateComment,
		sendEmail,
		message,
	)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}
	suite.NotNil(acct)
	suite.False(acct.Approved)

	// Wait for processor to
	// handle side effects.
	var (
		deniedUser *gtsmodel.DeniedUser
		err        error
	)
	if !testrig.WaitFor(func() bool {
		deniedUser, err = suite.state.DB.GetDeniedUserByID(ctx, targetUser.ID)
		return deniedUser != nil && err == nil
	}) {
		suite.FailNow("waiting for denied user")
	}

	// Ensure fields as expected.
	suite.Equal(targetUser.ID, deniedUser.ID)
	suite.Equal(targetUser.UnconfirmedEmail, deniedUser.Email)
	suite.Equal(targetAcct.Username, deniedUser.Username)
	suite.Equal(targetUser.SignUpIP, deniedUser.SignUpIP)
	suite.Equal(targetUser.InviteID, deniedUser.InviteID)
	suite.Equal(targetUser.Locale, deniedUser.Locale)
	suite.Equal(targetUser.CreatedByApplicationID, deniedUser.CreatedByApplicationID)
	suite.Equal(targetUser.Reason, deniedUser.SignUpReason)
	suite.Equal(privateComment, deniedUser.PrivateComment)
	suite.Equal(sendEmail, *deniedUser.SendEmail)
	suite.Equal(message, deniedUser.Message)

	// Should be no user entry for
	// this denied request now.
	_, err = suite.state.DB.GetUserByID(ctx, targetUser.ID)
	suite.ErrorIs(db.ErrNoEntries, err)

	// Should be no account entry for
	// this denied request now.
	_, err = suite.state.DB.GetAccountByID(ctx, targetAcct.ID)
	suite.ErrorIs(db.ErrNoEntries, err)
}

func (suite *AdminRejectTestSuite) TestRejectRemote() {
	var (
		ctx            = suite.T().Context()
		adminAcct      = suite.testAccounts["admin_account"]
		targetAcct     = suite.testAccounts["remote_account_1"]
		privateComment = "It's a no from me chief."
		sendEmail      = true
		message        = "Too stinky."
	)

	// Try to reject a remote account.
	_, err := suite.adminProcessor.SignupReject(
		ctx,
		adminAcct,
		targetAcct.ID,
		privateComment,
		sendEmail,
		message,
	)
	suite.EqualError(err, "user for account 01F8MH5ZK5VRH73AKHQM6Y9VNX not found")
}

func (suite *AdminRejectTestSuite) TestRejectApproved() {
	var (
		ctx            = suite.T().Context()
		adminAcct      = suite.testAccounts["admin_account"]
		targetAcct     = suite.testAccounts["local_account_1"]
		privateComment = "It's a no from me chief."
		sendEmail      = true
		message        = "Too stinky."
	)

	// Try to reject an already-approved account.
	_, err := suite.adminProcessor.SignupReject(
		ctx,
		adminAcct,
		targetAcct.ID,
		privateComment,
		sendEmail,
		message,
	)
	suite.EqualError(err, "account 01F8MH1H7YV1Z7D2C8K2730QBF has already been approved")
}

func TestAdminRejectTestSuite(t *testing.T) {
	suite.Run(t, new(AdminRejectTestSuite))
}
