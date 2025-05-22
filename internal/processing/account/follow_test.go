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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type FollowTestSuite struct {
	AccountStandardTestSuite
}

func (suite *FollowTestSuite) TestUpdateExistingFollowChangeBoth() {
	ctx := suite.T().Context()
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	// Change both Reblogs and Notify.
	// Trace logs should show a query similar to this:
	//	UPDATE "follows" AS "follow" SET "show_reblogs" = FALSE, "notify" = TRUE, "updated_at" = '2023-04-09 11:42:39.424705+00:00' WHERE ("follow"."id" = '01F8PY8RHWRQZV038T4E8T9YK8')
	relationship, err := suite.accountProcessor.FollowCreate(ctx, requestingAccount, &apimodel.AccountFollowRequest{
		ID:      targetAccount.ID,
		Reblogs: util.Ptr(false),
		Notify:  util.Ptr(true),
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.False(relationship.ShowingReblogs)
	suite.True(relationship.Notifying)
}

func (suite *FollowTestSuite) TestUpdateExistingFollowChangeNotifyIgnoreReblogs() {
	ctx := suite.T().Context()
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	// Change Notify, ignore Reblogs.
	// Trace logs should show a query similar to this:
	//	UPDATE "follows" AS "follow" SET "notify" = TRUE, "updated_at" = '2023-04-09 11:40:33.827858+00:00' WHERE ("follow"."id" = '01F8PY8RHWRQZV038T4E8T9YK8')
	relationship, err := suite.accountProcessor.FollowCreate(ctx, requestingAccount, &apimodel.AccountFollowRequest{
		ID:     targetAccount.ID,
		Notify: util.Ptr(true),
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.ShowingReblogs)
	suite.True(relationship.Notifying)
}

func (suite *FollowTestSuite) TestUpdateExistingFollowChangeNotifySetReblogs() {
	ctx := suite.T().Context()
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	// Change Notify, set Reblogs to same value as before.
	// Trace logs should show a query similar to this:
	//	UPDATE "follows" AS "follow" SET "notify" = TRUE, "updated_at" = '2023-04-09 11:40:33.827858+00:00' WHERE ("follow"."id" = '01F8PY8RHWRQZV038T4E8T9YK8')
	relationship, err := suite.accountProcessor.FollowCreate(ctx, requestingAccount, &apimodel.AccountFollowRequest{
		ID:      targetAccount.ID,
		Notify:  util.Ptr(true),
		Reblogs: util.Ptr(true),
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.ShowingReblogs)
	suite.True(relationship.Notifying)
}

func (suite *FollowTestSuite) TestUpdateExistingFollowChangeNothing() {
	ctx := suite.T().Context()
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	// Set Notify and Reblogs to same values as before.
	// Trace logs should show no update query.
	relationship, err := suite.accountProcessor.FollowCreate(ctx, requestingAccount, &apimodel.AccountFollowRequest{
		ID:      targetAccount.ID,
		Notify:  util.Ptr(false),
		Reblogs: util.Ptr(true),
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.ShowingReblogs)
	suite.False(relationship.Notifying)
}

func (suite *FollowTestSuite) TestUpdateExistingFollowSetNothing() {
	ctx := suite.T().Context()
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	// Don't set Notify or Reblogs.
	// Trace logs should show no update query.
	relationship, err := suite.accountProcessor.FollowCreate(ctx, requestingAccount, &apimodel.AccountFollowRequest{
		ID: targetAccount.ID,
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.ShowingReblogs)
	suite.False(relationship.Notifying)
}

func (suite *FollowTestSuite) TestFollowRequestLocal() {
	ctx := suite.T().Context()
	requestingAccount := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	// Have admin follow request turtle.
	_, err := suite.accountProcessor.FollowCreate(
		ctx,
		requestingAccount,
		&apimodel.AccountFollowRequest{
			ID:      targetAccount.ID,
			Reblogs: util.Ptr(true),
			Notify:  util.Ptr(false),
		})
	if err != nil {
		suite.FailNow(err.Error())
	}

	// There should be a message going to the worker.
	cMsg, _ := suite.getClientMsg(5 * time.Second)
	suite.Equal(ap.ActivityCreate, cMsg.APActivityType)
	suite.Equal(ap.ActivityFollow, cMsg.APObjectType)
	suite.Equal(requestingAccount.ID, cMsg.Origin.ID)
	suite.Equal(targetAccount.ID, cMsg.Target.ID)
}

func TestFollowTestS(t *testing.T) {
	suite.Run(t, new(FollowTestSuite))
}
