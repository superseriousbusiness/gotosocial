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

package bundb_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type RelationshipTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *RelationshipTestSuite) TestGetBlockBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 block models are equal.
	isEqual := func(b1, b2 gtsmodel.Block) bool {
		// Clear populated sub-models.
		b1.Account = nil
		b2.Account = nil
		b1.TargetAccount = nil
		b2.TargetAccount = nil

		// Clear database-set fields.
		b1.CreatedAt = time.Time{}
		b2.CreatedAt = time.Time{}
		b1.UpdatedAt = time.Time{}
		b2.UpdatedAt = time.Time{}

		return reflect.DeepEqual(b1, b2)
	}

	var testBlocks []*gtsmodel.Block

	for _, account1 := range suite.testAccounts {
		for _, account2 := range suite.testAccounts {
			if account1.ID == account2.ID {
				// don't block *yourself* ...
				continue
			}

			// Create new account block.
			block := &gtsmodel.Block{
				ID:              id.NewULID(),
				URI:             "http://127.0.0.1:8080/" + id.NewULID(),
				AccountID:       account1.ID,
				TargetAccountID: account2.ID,
			}

			// Attempt to place the block in database (if not already).
			if err := suite.db.PutBlock(ctx, block); err != nil {
				if err != db.ErrAlreadyExists {
					// Unrecoverable database error.
					t.Fatalf("error creating block: %v", err)
				}

				// Fetch existing block from database between accounts.
				block, _ = suite.db.GetBlock(ctx, account1.ID, account2.ID)
				continue
			}

			// Append generated block to test cases.
			testBlocks = append(testBlocks, block)
		}
	}

	for _, block := range testBlocks {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.Block, error){
			"id": func() (*gtsmodel.Block, error) {
				return suite.db.GetBlockByID(ctx, block.ID)
			},

			"uri": func() (*gtsmodel.Block, error) {
				return suite.db.GetBlockByURI(ctx, block.URI)
			},

			"origin_target": func() (*gtsmodel.Block, error) {
				return suite.db.GetBlock(ctx, block.AccountID, block.TargetAccountID)
			},
		} {

			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkBlock, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received block data.
			if !isEqual(*checkBlock, *block) {
				t.Errorf("block does not contain expected data: %+v", checkBlock)
				continue
			}

			// Check that block origin account populated.
			if checkBlock.Account == nil || checkBlock.Account.ID != block.AccountID {
				t.Errorf("block origin account not correctly populated for: %+v", checkBlock)
				continue
			}

			// Check that block target account populated.
			if checkBlock.TargetAccount == nil || checkBlock.TargetAccount.ID != block.TargetAccountID {
				t.Errorf("block target account not correctly populated for: %+v", checkBlock)
				continue
			}
		}
	}
}

func (suite *RelationshipTestSuite) TestGetFollowBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 follow models are equal.
	isEqual := func(f1, f2 gtsmodel.Follow) bool {
		// Clear populated sub-models.
		f1.Account = nil
		f2.Account = nil
		f1.TargetAccount = nil
		f2.TargetAccount = nil

		// Clear database-set fields.
		f1.CreatedAt = time.Time{}
		f2.CreatedAt = time.Time{}
		f1.UpdatedAt = time.Time{}
		f2.UpdatedAt = time.Time{}

		return reflect.DeepEqual(f1, f2)
	}

	var testFollows []*gtsmodel.Follow

	for _, account1 := range suite.testAccounts {
		for _, account2 := range suite.testAccounts {
			if account1.ID == account2.ID {
				// don't follow *yourself* ...
				continue
			}

			// Create new account follow.
			follow := &gtsmodel.Follow{
				ID:              id.NewULID(),
				URI:             "http://127.0.0.1:8080/" + id.NewULID(),
				AccountID:       account1.ID,
				TargetAccountID: account2.ID,
			}

			// Attempt to place the follow in database (if not already).
			if err := suite.db.PutFollow(ctx, follow); err != nil {
				if err != db.ErrAlreadyExists {
					// Unrecoverable database error.
					t.Fatalf("error creating follow: %v", err)
				}

				// Fetch existing follow from database between accounts.
				follow, _ = suite.db.GetFollow(ctx, account1.ID, account2.ID)
				continue
			}

			// Append generated follow to test cases.
			testFollows = append(testFollows, follow)
		}
	}

	for _, follow := range testFollows {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.Follow, error){
			"id": func() (*gtsmodel.Follow, error) {
				return suite.db.GetFollowByID(ctx, follow.ID)
			},

			"uri": func() (*gtsmodel.Follow, error) {
				return suite.db.GetFollowByURI(ctx, follow.URI)
			},

			"origin_target": func() (*gtsmodel.Follow, error) {
				return suite.db.GetFollow(ctx, follow.AccountID, follow.TargetAccountID)
			},
		} {
			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkFollow, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received follow data.
			if !isEqual(*checkFollow, *follow) {
				t.Errorf("follow does not contain expected data: %+v", checkFollow)
				continue
			}

			// Check that follow origin account populated.
			if checkFollow.Account == nil || checkFollow.Account.ID != follow.AccountID {
				t.Errorf("follow origin account not correctly populated for: %+v", checkFollow)
				continue
			}

			// Check that follow target account populated.
			if checkFollow.TargetAccount == nil || checkFollow.TargetAccount.ID != follow.TargetAccountID {
				t.Errorf("follow target account not correctly populated for: %+v", checkFollow)
				continue
			}
		}
	}
}

func (suite *RelationshipTestSuite) TestGetFollowRequestBy() {
	t := suite.T()

	// Create a new context for this test.
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Sentinel error to mark avoiding a test case.
	sentinelErr := errors.New("sentinel")

	// isEqual checks if 2 follow request models are equal.
	isEqual := func(f1, f2 gtsmodel.FollowRequest) bool {
		// Clear populated sub-models.
		f1.Account = nil
		f2.Account = nil
		f1.TargetAccount = nil
		f2.TargetAccount = nil

		// Clear database-set fields.
		f1.CreatedAt = time.Time{}
		f2.CreatedAt = time.Time{}
		f1.UpdatedAt = time.Time{}
		f2.UpdatedAt = time.Time{}

		return reflect.DeepEqual(f1, f2)
	}

	var testFollowReqs []*gtsmodel.FollowRequest

	for _, account1 := range suite.testAccounts {
		for _, account2 := range suite.testAccounts {
			if account1.ID == account2.ID {
				// don't follow *yourself* ...
				continue
			}

			// Create new account follow request.
			followReq := &gtsmodel.FollowRequest{
				ID:              id.NewULID(),
				URI:             "http://127.0.0.1:8080/" + id.NewULID(),
				AccountID:       account1.ID,
				TargetAccountID: account2.ID,
			}

			// Attempt to place the follow in database (if not already).
			if err := suite.db.PutFollowRequest(ctx, followReq); err != nil {
				if err != db.ErrAlreadyExists {
					// Unrecoverable database error.
					t.Fatalf("error creating follow request: %v", err)
				}

				// Fetch existing follow request from database between accounts.
				followReq, _ = suite.db.GetFollowRequest(ctx, account1.ID, account2.ID)
				continue
			}

			// Append generated follow request to test cases.
			testFollowReqs = append(testFollowReqs, followReq)
		}
	}

	for _, followReq := range testFollowReqs {
		for lookup, dbfunc := range map[string]func() (*gtsmodel.FollowRequest, error){
			"id": func() (*gtsmodel.FollowRequest, error) {
				return suite.db.GetFollowRequestByID(ctx, followReq.ID)
			},

			"uri": func() (*gtsmodel.FollowRequest, error) {
				return suite.db.GetFollowRequestByURI(ctx, followReq.URI)
			},

			"origin_target": func() (*gtsmodel.FollowRequest, error) {
				return suite.db.GetFollowRequest(ctx, followReq.AccountID, followReq.TargetAccountID)
			},
		} {

			// Clear database caches.
			suite.state.Caches.Init()

			t.Logf("checking database lookup %q", lookup)

			// Perform database function.
			checkFollowReq, err := dbfunc()
			if err != nil {
				if err == sentinelErr {
					continue
				}

				t.Errorf("error encountered for database lookup %q: %v", lookup, err)
				continue
			}

			// Check received follow request data.
			if !isEqual(*checkFollowReq, *followReq) {
				t.Errorf("follow request does not contain expected data: %+v", checkFollowReq)
				continue
			}

			// Check that follow request origin account populated.
			if checkFollowReq.Account == nil || checkFollowReq.Account.ID != followReq.AccountID {
				t.Errorf("follow request origin account not correctly populated for: %+v", checkFollowReq)
				continue
			}

			// Check that follow request target account populated.
			if checkFollowReq.TargetAccount == nil || checkFollowReq.TargetAccount.ID != followReq.TargetAccountID {
				t.Errorf("follow request target account not correctly populated for: %+v", checkFollowReq)
				continue
			}
		}
	}
}

func (suite *RelationshipTestSuite) TestIsBlocked() {
	ctx := context.Background()

	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID

	// no blocks exist between account 1 and account 2
	blocked, err := suite.db.IsBlocked(ctx, account1, account2)
	suite.NoError(err)
	suite.False(blocked)

	blocked, err = suite.db.IsBlocked(ctx, account2, account1)
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
	blocked, err = suite.db.IsBlocked(ctx, account1, account2)
	suite.NoError(err)
	suite.True(blocked)

	// account 2 doesn't block account 1
	blocked, err = suite.db.IsBlocked(ctx, account2, account1)
	suite.NoError(err)
	suite.False(blocked)

	// a block exists in either direction between the two
	blocked, err = suite.db.IsEitherBlocked(ctx, account1, account2)
	suite.NoError(err)
	suite.True(blocked)
	blocked, err = suite.db.IsEitherBlocked(ctx, account2, account1)
	suite.NoError(err)
	suite.True(blocked)
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

func (suite *RelationshipTestSuite) TestDeleteAccountBlocks() {
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
	err = suite.db.DeleteAccountBlocks(ctx, account1)
	suite.NoError(err)

	// block should be gone
	block, err = suite.db.GetBlock(ctx, account1, account2)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(block)
}

func (suite *RelationshipTestSuite) TestDeleteAccountMutes() {
	ctx := context.Background()

	// Add a mute.
	accountID1 := suite.testAccounts["local_account_1"].ID
	accountID2 := suite.testAccounts["local_account_2"].ID
	muteID := "01HZGZ3F3C7S1TTPE8F9VPZDCB"
	err := suite.db.PutMute(ctx, &gtsmodel.UserMute{
		ID:              muteID,
		AccountID:       accountID1,
		TargetAccountID: accountID2,
	})
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Make sure the mute is in the DB.
	mute, err := suite.db.GetMute(ctx, accountID1, accountID2)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if suite.NotNil(mute) {
		suite.Equal(muteID, mute.ID)
	}

	// Delete all mutes owned by that account.
	err = suite.db.DeleteAccountMutes(ctx, accountID1)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Mute should be gone.
	mute, err = suite.db.GetMute(ctx, accountID1, accountID2)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(mute)
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
	isFollowing, err := suite.db.IsFollowing(context.Background(), requestingAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.True(isFollowing)
}

func (suite *RelationshipTestSuite) TestIsFollowingNo() {
	requestingAccount := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]
	isFollowing, err := suite.db.IsFollowing(context.Background(), requestingAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.False(isFollowing)
}

func (suite *RelationshipTestSuite) TestIsMutualFollowing() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]
	isMutualFollowing, err := suite.db.IsMutualFollowing(context.Background(), requestingAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.True(isMutualFollowing)
}

func (suite *RelationshipTestSuite) TestIsMutualFollowingNo() {
	requestingAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["local_account_2"]
	isMutualFollowing, err := suite.db.IsMutualFollowing(context.Background(), requestingAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.True(isMutualFollowing)
}

func (suite *RelationshipTestSuite) TestAcceptFollowRequestOK() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	// Fetch relationship before follow request.
	relationship, err := suite.db.GetRelationship(ctx, account.ID, targetAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(relationship.Following)
	suite.False(relationship.Requested)

	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
	}

	if err := suite.db.PutFollowRequest(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	// Fetch relationship while follow requested.
	relationship, err = suite.db.GetRelationship(ctx, account.ID, targetAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(relationship.Following)
	suite.True(relationship.Requested)

	// Check the other way around too; local_account_2
	// should have requested_by true for admin now.
	inverse, err := suite.db.GetRelationship(ctx, targetAccount.ID, account.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(inverse.RequestedBy)

	followRequestNotification := &gtsmodel.Notification{
		ID:               "01GV8MY1Q9KX2ZSWN4FAQ3V1PB",
		OriginAccountID:  account.ID,
		TargetAccountID:  targetAccount.ID,
		NotificationType: gtsmodel.NotificationFollowRequest,
	}

	if err := suite.db.PutNotification(ctx, followRequestNotification); err != nil {
		suite.FailNow(err.Error())
	}

	follow, err := suite.db.AcceptFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(follow)
	suite.Equal(followRequest.URI, follow.URI)

	// Ensure notification is deleted.
	notification, err := suite.db.GetNotificationByID(ctx, followRequestNotification.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(notification)

	// Fetch relationship while followed.
	relationship, err = suite.db.GetRelationship(ctx, account.ID, targetAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(relationship.Following)
	suite.False(relationship.Requested)
}

func (suite *RelationshipTestSuite) TestAcceptFollowRequestNoNotification() {
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

	// Unlike the above test, don't create a notification.
	// Follow request accept should still produce no error.

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

	if err := suite.db.PutFollowRequest(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	followRequestNotification := &gtsmodel.Notification{
		ID:               "01GV8MY1Q9KX2ZSWN4FAQ3V1PB",
		OriginAccountID:  account.ID,
		TargetAccountID:  targetAccount.ID,
		NotificationType: gtsmodel.NotificationFollowRequest,
	}

	if err := suite.db.Put(ctx, followRequestNotification); err != nil {
		suite.FailNow(err.Error())
	}

	err := suite.db.RejectFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.NoError(err)

	// Ensure notification is deleted.
	notification, err := suite.db.GetNotificationByID(ctx, followRequestNotification.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(notification)
}

func (suite *RelationshipTestSuite) TestRejectFollowRequestNotExisting() {
	ctx := context.Background()
	account := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	err := suite.db.RejectFollowRequest(ctx, account.ID, targetAccount.ID)
	suite.NoError(err)
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

	followRequests, err := suite.db.GetAccountFollowRequests(ctx, targetAccount.ID, nil)
	suite.NoError(err)
	suite.Len(followRequests, 1)
}

func (suite *RelationshipTestSuite) TestGetAccountFollows() {
	account := suite.testAccounts["local_account_1"]
	follows, err := suite.db.GetAccountFollows(context.Background(), account.ID, nil)
	suite.NoError(err)
	suite.Len(follows, 2)
}

func (suite *RelationshipTestSuite) TestGetAccountFollowers() {
	account := suite.testAccounts["local_account_1"]
	follows, err := suite.db.GetAccountFollowers(context.Background(), account.ID, nil)
	suite.NoError(err)
	suite.Len(follows, 2)
}

func (suite *RelationshipTestSuite) TestUnfollowExisting() {
	originAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	follow, err := suite.db.GetFollow(context.Background(), originAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(follow)
	followID := follow.ID

	// We should have lists that this follow is a part of.
	lists, err := suite.db.GetListsContainingFollowID(context.Background(), followID)
	suite.NoError(err)
	suite.NotEmpty(lists)

	err = suite.db.DeleteFollowByID(context.Background(), followID)
	suite.NoError(err)

	follow, err = suite.db.GetFollow(context.Background(), originAccount.ID, targetAccount.ID)
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(follow)

	// Lists containing this follow should return empty too.
	lists, err = suite.db.GetListsContainingFollowID(context.Background(), followID)
	suite.NoError(err)
	suite.Empty(lists)
}

func (suite *RelationshipTestSuite) TestGetFollowNotExisting() {
	originAccount := suite.testAccounts["local_account_1"]
	targetAccountID := "01GTVD9N484CZ6AM90PGGNY7GQ"

	follow, err := suite.db.GetFollow(context.Background(), originAccount.ID, targetAccountID)
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(follow)
}

func (suite *RelationshipTestSuite) TestDeleteFollow() {
	ctx := context.Background()
	originAccount := suite.testAccounts["local_account_1"]
	targetAccount := suite.testAccounts["admin_account"]

	err := suite.db.DeleteFollow(ctx, originAccount.ID, targetAccount.ID)
	suite.NoError(err)

	follow, err := suite.db.GetFollow(ctx, originAccount.ID, targetAccount.ID)
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(follow)
}

func (suite *RelationshipTestSuite) TestUnfollowRequestExisting() {
	ctx := context.Background()
	originAccount := suite.testAccounts["admin_account"]
	targetAccount := suite.testAccounts["local_account_2"]

	followRequest := &gtsmodel.FollowRequest{
		ID:              "01GEF753FWHCHRDWR0QEHBXM8W",
		URI:             "http://localhost:8080/weeeeeeeeeeeeeeeee",
		AccountID:       originAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	if err := suite.db.PutFollowRequest(ctx, followRequest); err != nil {
		suite.FailNow(err.Error())
	}

	followRequest, err := suite.db.GetFollowRequest(context.Background(), originAccount.ID, targetAccount.ID)
	suite.NoError(err)
	suite.NotNil(followRequest)

	err = suite.db.DeleteFollowRequestByID(context.Background(), followRequest.ID)
	suite.NoError(err)

	followRequest, err = suite.db.GetFollowRequest(context.Background(), originAccount.ID, targetAccount.ID)
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(followRequest)
}

func (suite *RelationshipTestSuite) TestUnfollowRequestNotExisting() {
	originAccount := suite.testAccounts["local_account_1"]
	targetAccountID := "01GTVD9N484CZ6AM90PGGNY7GQ"

	followRequest, err := suite.db.GetFollowRequest(context.Background(), originAccount.ID, targetAccountID)
	suite.EqualError(err, db.ErrNoEntries.Error())
	suite.Nil(followRequest)
}

func (suite *RelationshipTestSuite) TestUpdateFollow() {
	ctx := context.Background()

	follow := &gtsmodel.Follow{}
	*follow = *suite.testFollows["local_account_1_admin_account"]

	follow.Notify = util.Ptr(true)
	if err := suite.db.UpdateFollow(ctx, follow, "notify"); err != nil {
		suite.FailNow(err.Error())
	}

	dbFollow, err := suite.db.GetFollowByID(ctx, follow.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(*dbFollow.Notify)

	relationship, err := suite.db.GetRelationship(ctx, follow.AccountID, follow.TargetAccountID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.True(relationship.Notifying)
}

func (suite *RelationshipTestSuite) TestGetNote() {
	ctx := context.Background()

	// Retrieve a fixture note
	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID
	expectedNote := suite.testAccountNotes["local_account_2_note_on_1"]
	note, err := suite.db.GetNote(ctx, account2, account1)
	suite.NoError(err)
	suite.NotNil(note)
	suite.Equal(expectedNote.ID, note.ID)
	suite.Equal(expectedNote.Comment, note.Comment)
}

func (suite *RelationshipTestSuite) TestPutNote() {
	ctx := context.Background()

	// put a note in
	account1 := suite.testAccounts["local_account_1"].ID
	account2 := suite.testAccounts["local_account_2"].ID
	err := suite.db.PutNote(ctx, &gtsmodel.AccountNote{
		ID:              "01H539R2NA0M83JX15Y5RWKE97",
		AccountID:       account1,
		TargetAccountID: account2,
		Comment:         "foo",
	})
	suite.NoError(err)

	// make sure the note is in the db
	note, err := suite.db.GetNote(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(note)
	suite.Equal("01H539R2NA0M83JX15Y5RWKE97", note.ID)
	suite.Equal("foo", note.Comment)

	// update the note
	note.Comment = "bar"
	err = suite.db.PutNote(ctx, note)
	suite.NoError(err)

	// make sure the comment changes
	note, err = suite.db.GetNote(ctx, account1, account2)
	suite.NoError(err)
	suite.NotNil(note)
	suite.Equal("bar", note.Comment)
}

func TestRelationshipTestSuite(t *testing.T) {
	suite.Run(t, new(RelationshipTestSuite))
}
