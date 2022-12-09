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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Relationship contains functions for getting or modifying the relationship between two accounts.
type Relationship interface {
	// IsBlocked checks whether account 1 has a block in place against account2.
	// If eitherDirection is true, then the function returns true if account1 blocks account2, OR if account2 blocks account1.
	IsBlocked(ctx context.Context, account1 string, account2 string, eitherDirection bool) (bool, Error)

	// GetBlock returns the block from account1 targeting account2, if it exists, or an error if it doesn't.
	//
	// Because this is slower than Blocked, only use it if you need the actual Block struct for some reason,
	// not if you're just checking for the existence of a block.
	GetBlock(ctx context.Context, account1 string, account2 string) (*gtsmodel.Block, Error)

	// PutBlock attempts to place the given account block in the database.
	PutBlock(ctx context.Context, block *gtsmodel.Block) Error

	// DeleteBlockByID removes block with given ID from the database.
	DeleteBlockByID(ctx context.Context, id string) Error

	// DeleteBlockByURI removes block with given AP URI from the database.
	DeleteBlockByURI(ctx context.Context, uri string) Error

	// DeleteBlocksByOriginAccountID removes any blocks with accountID equal to originAccountID.
	DeleteBlocksByOriginAccountID(ctx context.Context, originAccountID string) Error

	// DeleteBlocksByTargetAccountID removes any blocks with given targetAccountID.
	DeleteBlocksByTargetAccountID(ctx context.Context, targetAccountID string) Error

	// GetRelationship retrieves the relationship of the targetAccount to the requestingAccount.
	GetRelationship(ctx context.Context, requestingAccount string, targetAccount string) (*gtsmodel.Relationship, Error)

	// IsFollowing returns true if sourceAccount follows target account, or an error if something goes wrong while finding out.
	IsFollowing(ctx context.Context, sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, Error)

	// IsFollowRequested returns true if sourceAccount has requested to follow target account, or an error if something goes wrong while finding out.
	IsFollowRequested(ctx context.Context, sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, Error)

	// IsMutualFollowing returns true if account1 and account2 both follow each other, or an error if something goes wrong while finding out.
	IsMutualFollowing(ctx context.Context, account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, Error)

	// AcceptFollowRequest moves a follow request in the database from the follow_requests table to the follows table.
	// In other words, it should create the follow, and delete the existing follow request.
	//
	// It will return the newly created follow for further processing.
	AcceptFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.Follow, Error)

	// RejectFollowRequest fetches a follow request from the database, and then deletes it.
	//
	// The deleted follow request will be returned so that further processing can be done on it.
	RejectFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.FollowRequest, Error)

	// GetAccountFollowRequests returns all follow requests targeting the given account.
	GetAccountFollowRequests(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, Error)

	// GetAccountFollows returns a slice of follows owned by the given accountID.
	GetAccountFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, Error)

	// CountAccountFollows returns the amount of accounts that the given accountID is following.
	//
	// If localOnly is set to true, then only follows from *this instance* will be returned.
	CountAccountFollows(ctx context.Context, accountID string, localOnly bool) (int, Error)

	// GetAccountFollowedBy fetches follows that target given accountID.
	//
	// If localOnly is set to true, then only follows from *this instance* will be returned.
	GetAccountFollowedBy(ctx context.Context, accountID string, localOnly bool) ([]*gtsmodel.Follow, Error)

	// CountAccountFollowedBy returns the amounts that the given ID is followed by.
	CountAccountFollowedBy(ctx context.Context, accountID string, localOnly bool) (int, Error)
}
