/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package bundb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type RelationshipTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *RelationshipTestSuite) TestIsBlocked() {
	ctx := context.Background()

	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID

	// no blocks exist between account 1 and account 2
	blocked, err := suite.db.IsBlocked(ctx, account1, account2, false)
	suite.NoError(err)
	suite.False(blocked)

	blocked, err = suite.db.IsBlocked(ctx, account2, account1, false)
	suite.NoError(err)
	suite.False(blocked)

	// have account1 block account2
	if err := suite.db.PutBlock(ctx, &gtsmodel.Block{
		ID:              "01G202BCSXXJZ70BHB5KCAHH8C",
		URI:             "http://localhost:8080/some_block_uri_1",
		AccountID:       account1,
		TargetAccountID: account2,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// account 1 now blocks account 2
	blocked, err = suite.db.IsBlocked(ctx, account1, account2, false)
	suite.NoError(err)
	suite.True(blocked)

	// account 2 doesn't block account 1
	blocked, err = suite.db.IsBlocked(ctx, account2, account1, false)
	suite.NoError(err)
	suite.False(blocked)

	// a block exists in either direction between the two
	blocked, err = suite.db.IsBlocked(ctx, account1, account2, true)
	suite.NoError(err)
	suite.True(blocked)
	blocked, err = suite.db.IsBlocked(ctx, account2, account1, true)
	suite.NoError(err)
	suite.True(blocked)
}

func (suite *RelationshipTestSuite) TestGetBlock() {
	ctx := context.Background()

	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID

	if err := suite.db.PutBlock(ctx, &gtsmodel.Block{
		ID:              "01G202BCSXXJZ70BHB5KCAHH8C",
		URI:             "http://localhost:8080/some_block_uri_1",
		AccountID:       account1,
		TargetAccountID: account2,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	block, err := suite.db.GetBlock(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(block)
	suite.Equal("01G202BCSXXJZ70BHB5KCAHH8C", block.ID)
}

func (suite *RelationshipTestSuite) TestDeleteBlockByID() {
	ctx := context.Background()

	// put a block in first
	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID
	if err := suite.db.PutBlock(ctx, &gtsmodel.Block{
		ID:              "01G202BCSXXJZ70BHB5KCAHH8C",
		URI:             "http://localhost:8080/some_block_uri_1",
		AccountID:       account1,
		TargetAccountID: account2,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// make sure the block is in the db
	block, err := suite.db.GetBlock(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(block)
	suite.Equal("01G202BCSXXJZ70BHB5KCAHH8C", block.ID)

	// delete the block by ID
	err = suite.db.DeleteBlockByID(ctx, "01G202BCSXXJZ70BHB5KCAHH8C")
	suite.NoError(err)

	// block should be gone
	block, err = suite.db.GetBlock(ctx, account1, account2)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(block)
}

func (suite *RelationshipTestSuite) TestDeleteBlockByURI() {
	ctx := context.Background()

	// put a block in first
	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID
	if err := suite.db.PutBlock(ctx, &gtsmodel.Block{
		ID:              "01G202BCSXXJZ70BHB5KCAHH8C",
		URI:             "http://localhost:8080/some_block_uri_1",
		AccountID:       account1,
		TargetAccountID: account2,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// make sure the block is in the db
	block, err := suite.db.GetBlock(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(block)
	suite.Equal("01G202BCSXXJZ70BHB5KCAHH8C", block.ID)

	// delete the block by uri
	err = suite.db.DeleteBlockByURI(ctx, "http://localhost:8080/some_block_uri_1")
	suite.NoError(err)

	// block should be gone
	block, err = suite.db.GetBlock(ctx, account1, account2)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(block)
}

func (suite *RelationshipTestSuite) TestDeleteBlocksByOriginAccountID() {
	ctx := context.Background()

	// put a block in first
	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID
	if err := suite.db.PutBlock(ctx, &gtsmodel.Block{
		ID:              "01G202BCSXXJZ70BHB5KCAHH8C",
		URI:             "http://localhost:8080/some_block_uri_1",
		AccountID:       account1,
		TargetAccountID: account2,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// make sure the block is in the db
	block, err := suite.db.GetBlock(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(block)
	suite.Equal("01G202BCSXXJZ70BHB5KCAHH8C", block.ID)

	// delete the block by originAccountID
	err = suite.db.DeleteBlocksByOriginAccountID(ctx, account1)
	suite.NoError(err)

	// block should be gone
	block, err = suite.db.GetBlock(ctx, account1, account2)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(block)
}

func (suite *RelationshipTestSuite) TestDeleteBlocksByTargetAccountID() {
	ctx := context.Background()

	// put a block in first
	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID
	if err := suite.db.PutBlock(ctx, &gtsmodel.Block{
		ID:              "01G202BCSXXJZ70BHB5KCAHH8C",
		URI:             "http://localhost:8080/some_block_uri_1",
		AccountID:       account1,
		TargetAccountID: account2,
	}); err != nil {
		suite.FailNow(err.Error())
	}

	// make sure the block is in the db
	block, err := suite.db.GetBlock(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(block)
	suite.Equal("01G202BCSXXJZ70BHB5KCAHH8C", block.ID)

	// delete the block by targetAccountID
	err = suite.db.DeleteBlocksByTargetAccountID(ctx, account2)
	suite.NoError(err)

	// block should be gone
	block, err = suite.db.GetBlock(ctx, account1, account2)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(block)
}

func (suite *RelationshipTestSuite) TestGetRelationship() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	relationship, err := suite.db.GetRelationship(context.Background(), requestingAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(relationship)

	suite.True(relationship.Following)
	suite.True(relationship.ShowingReblogs)
	suite.False(relationship.Notifying)
	suite.True(relationship.FollowedBy)
	suite.False(relationship.Blocking)
	suite.False(relationship.BlockedBy)
	suite.False(relationship.Muting)
	suite.False(relationship.MutingNotifications)
	suite.False(relationship.Requested)
	suite.False(relationship.DomainBlocking)
	suite.False(relationship.Endorsed)
	suite.Empty(relationship.Note)
}

func (suite *RelationshipTestSuite) TestIsFollowingYes() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]
	isFollowing, err := suite.db.IsFollowing(context.Background(), requestingAccount, targetAccount)
	suite.NoError(err)
	suite.True(isFollowing)
}

func (suite *RelationshipTestSuite) TestIsFollowingNo() {
	requestingAccount := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]
	isFollowing, err := suite.db.IsFollowing(context.Background(), requestingAccount, targetAccount)
	suite.NoError(err)
	suite.False(isFollowing)
}

func (suite *RelationshipTestSuite) TestIsMutualFollowing() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]
	isMutualFollowing, err := suite.db.IsMutualFollowing(context.Background(), requestingAccount, targetAccount)
	suite.NoError(err)
	suite.True(isMutualFollowing)
}

func (suite *RelationshipTestSuite) TestIsMutualFollowingNo() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]
	isMutualFollowing, err := suite.db.IsMutualFollowing(context.Background(), requestingAccount, targetAccount)
	suite.NoError(err)
	suite.True(isMutualFollowing)
}

func (suite *RelationshipTestSuite) TestAcceptFollowRequestOK() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
	}

	if err := suite.db.Put(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	follow, err := suite.db.AcceptFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(follow)
	suite.Equal(followRequest.URI, follow.URI)
}

func (suite *RelationshipTestSuite) TestAcceptFollowRequestNotExisting() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	follow, err := suite.db.AcceptFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(follow)
}

func (suite *RelationshipTestSuite) TestAcceptFollowRequestFollowAlreadyExists() {
	ctx := context.Background()
	account := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	// follow already exists in the db from local_account_1 -> admin_account
	existingFollow := &gtsmodel.Follow{}
	if err := suite.db.GetByID(ctx, suite.testFollows["local_account_1_admin_account"].ID, existingFollow); err != nil {
		suite.FailNow(err.Error())
	}

	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
	}

	if err := suite.db.Put(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	follow, err := suite.db.AcceptFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(follow)

	// uri should be equal to value of new/overlapping follow request
	suite.NotEqual(followRequest.URI, existingFollow.URI)
	suite.Equal(followRequest.URI, follow.URI)
}

func (suite *RelationshipTestSuite) TestRejectFollowRequestOK() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
	}

	if err := suite.db.Put(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	rejectedFollowRequest, err := suite.db.RejectFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(rejectedFollowRequest)
}

func (suite *RelationshipTestSuite) TestRejectFollowRequestNotExisting() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	rejectedFollowRequest, err := suite.db.RejectFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(rejectedFollowRequest)
}

func (suite *RelationshipTestSuite) TestGetAccountFollowRequests() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
	}

	if err := suite.db.Put(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	followRequests, err := suite.db.GetAccountFollowRequests(ctx, targetAccount.ID)
	suite.NoError(err)
	suite.Len(followRequests, 1)
}

func (suite *RelationshipTestSuite) TestGetAccountFollows() {
	account := suite.testAccounts["local_account_1"]
	follows, err := suite.db.GetAccountFollows(context.Background(), account.ID)
	suite.NoError(err)
	suite.Len(follows, 2)
}

func (suite *RelationshipTestSuite) TestCountAccountFollowsLocalOnly() {
	account := suite.testAccounts["local_account_1"]
	followsCount, err := suite.db.CountAccountFollows(context.Background(), account.ID, true)
	suite.NoError(err)
	suite.Equal(2, followsCount)
}

func (suite *RelationshipTestSuite) TestCountAccountFollows() {
	account := suite.testAccounts["local_account_1"]
	followsCount, err := suite.db.CountAccountFollows(context.Background(), account.ID, false)
	suite.NoError(err)
	suite.Equal(2, followsCount)
}

func (suite *RelationshipTestSuite) TestGetAccountFollowedBy() {
	account := suite.testAccounts["local_account_1"]
	follows, err := suite.db.GetAccountFollowedBy(context.Background(), account.ID, false)
	suite.NoError(err)
	suite.Len(follows, 2)
}

func (suite *RelationshipTestSuite) TestGetAccountFollowedByLocalOnly() {
	account := suite.testAccounts["local_account_1"]
	follows, err := suite.db.GetAccountFollowedBy(context.Background(), account.ID, true)
	suite.NoError(err)
	suite.Len(follows, 2)
}

func (suite *RelationshipTestSuite) TestCountAccountFollowedBy() {
	account := suite.testAccounts["local_account_1"]
	followsCount, err := suite.db.CountAccountFollowedBy(context.Background(), account.ID, false)
	suite.NoError(err)
	suite.Equal(2, followsCount)
}

func (suite *RelationshipTestSuite) TestCountAccountFollowedByLocalOnly() {
	account := suite.testAccounts["local_account_1"]
	followsCount, err := suite.db.CountAccountFollowedBy(context.Background(), account.ID, true)
	suite.NoError(err)
	suite.Equal(2, followsCount)
}

func TestRelationshipTestSuite(t *testing.T) {
	suite.Run(t, new(RelationshipTestSuite))
}
