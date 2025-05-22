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
	"testing"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/stretchr/testify/suite"
)

type MoveTestSuite struct {
	AccountStandardTestSuite
}

func (suite *MoveTestSuite) TestMoveAccountOK() {
	ctx := suite.T().Context()

	// Copy zork.
	requestingAcct := new(gtsmodel.Account)
	*requestingAcct = *suite.testAccounts["local_account_1"]

	// Copy admin.
	targetAcct := new(gtsmodel.Account)
	*targetAcct = *suite.testAccounts["admin_account"]

	// Update admin to alias back to zork.
	targetAcct.AlsoKnownAsURIs = []string{requestingAcct.URI}
	if err := suite.state.DB.UpdateAccount(
		ctx,
		targetAcct,
		"also_known_as_uris",
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Trigger move from zork to admin.
	if err := suite.accountProcessor.MoveSelf(
		ctx,
		&apiutil.Auth{
			Token:       oauth.DBTokenToToken(suite.testTokens["local_account_1"]),
			Application: suite.testApplications["local_account_1"],
			User:        suite.testUsers["local_account_1"],
			Account:     requestingAcct,
		},
		&apimodel.AccountMoveRequest{
			Password:   "password",
			MovedToURI: targetAcct.URI,
		},
	); err != nil {
		suite.FailNow(err.Error())
	}

	// There should be a message going to the worker.
	cMsg, _ := suite.getClientMsg(5 * time.Second)
	move, ok := cMsg.GTSModel.(*gtsmodel.Move)
	if !ok {
		suite.FailNow("", "could not cast %T to *gtsmodel.Move", move)
	}
	now := time.Now()
	suite.WithinDuration(now, move.CreatedAt, 5*time.Second)
	suite.WithinDuration(now, move.UpdatedAt, 5*time.Second)
	suite.WithinDuration(now, move.AttemptedAt, 5*time.Second)
	suite.Zero(move.SucceededAt)
	suite.NotZero(move.ID)
	suite.Equal(requestingAcct.URI, move.OriginURI)
	suite.NotNil(move.Origin)
	suite.Equal(targetAcct.URI, move.TargetURI)
	suite.NotNil(move.Target)
	suite.NotZero(move.URI)

	// Move should be in the database now.
	move, err := suite.state.DB.GetMoveByOriginTarget(
		ctx,
		requestingAcct.URI,
		targetAcct.URI,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(move)

	// Origin account should have move ID and move to URI set.
	suite.Equal(move.ID, requestingAcct.MoveID)
	suite.Equal(targetAcct.URI, requestingAcct.MovedToURI)
}

func (suite *MoveTestSuite) TestMoveAccountNotAliased() {
	ctx := suite.T().Context()

	// Copy zork.
	requestingAcct := new(gtsmodel.Account)
	*requestingAcct = *suite.testAccounts["local_account_1"]

	// Don't copy admin.
	targetAcct := suite.testAccounts["admin_account"]

	// Trigger move from zork to admin.
	//
	// Move should fail since admin is
	// not aliased back to zork.
	err := suite.accountProcessor.MoveSelf(
		ctx,
		&apiutil.Auth{
			Token:       oauth.DBTokenToToken(suite.testTokens["local_account_1"]),
			Application: suite.testApplications["local_account_1"],
			User:        suite.testUsers["local_account_1"],
			Account:     requestingAcct,
		},
		&apimodel.AccountMoveRequest{
			Password:   "password",
			MovedToURI: targetAcct.URI,
		},
	)
	suite.EqualError(err, "target account http://localhost:8080/users/admin is not aliased to this account via alsoKnownAs; if you just changed it, please wait a few minutes and try the Move again")
}

func (suite *MoveTestSuite) TestMoveAccountBadPassword() {
	ctx := suite.T().Context()

	// Copy zork.
	requestingAcct := new(gtsmodel.Account)
	*requestingAcct = *suite.testAccounts["local_account_1"]

	// Don't copy admin.
	targetAcct := suite.testAccounts["admin_account"]

	// Trigger move from zork to admin.
	//
	// Move should fail since admin is
	// not aliased back to zork.
	err := suite.accountProcessor.MoveSelf(
		ctx,
		&apiutil.Auth{
			Token:       oauth.DBTokenToToken(suite.testTokens["local_account_1"]),
			Application: suite.testApplications["local_account_1"],
			User:        suite.testUsers["local_account_1"],
			Account:     requestingAcct,
		},
		&apimodel.AccountMoveRequest{
			Password:   "boobies",
			MovedToURI: targetAcct.URI,
		},
	)
	suite.EqualError(err, "invalid password provided in Move request")
}

func TestMoveTestSuite(t *testing.T) {
	suite.Run(t, new(MoveTestSuite))
}
