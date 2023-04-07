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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Relationship contains functions for getting or modifying the relationship between two accounts.
type Relationship interface {
	// IsBlocked checks whether source account has a block in place against target.
	IsBlocked(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, Error)

	// IsEitherBlocked checks whether there is a block in place between either of account1 and account2.
	IsEitherBlocked(ctx context.Context, accountID1 string, accountID2 string) (bool, error)

	// GetBlockByID fetches block with given ID from the database.
	GetBlockByID(ctx context.Context, id string) (*gtsmodel.Block, error)

	// GetBlockByURI fetches block with given AP URI from the database.
	GetBlockByURI(ctx context.Context, uri string) (*gtsmodel.Block, error)

	// GetBlock returns the block from account1 targeting account2, if it exists, or an error if it doesn't.
	GetBlock(ctx context.Context, account1 string, account2 string) (*gtsmodel.Block, error)

	// PutBlock attempts to place the given account block in the database.
	PutBlock(ctx context.Context, block *gtsmodel.Block) error

	// DeleteBlockByID removes block with given ID from the database.
	DeleteBlockByID(ctx context.Context, id string) error

	// DeleteBlockByURI removes block with given AP URI from the database.
	DeleteBlockByURI(ctx context.Context, uri string) error

	// DeleteAccountBlocks will delete all database blocks to / from the given account ID.
	DeleteAccountBlocks(ctx context.Context, accountID string) error

	// GetRelationship retrieves the relationship of the targetAccount to the requestingAccount.
	GetRelationship(ctx context.Context, requestingAccount string, targetAccount string) (*gtsmodel.Relationship, Error)

	// GetFollowByID fetches follow with given ID from the database.
	GetFollowByID(ctx context.Context, id string) (*gtsmodel.Follow, error)

	// GetFollowByURI fetches follow with given AP URI from the database.
	GetFollowByURI(ctx context.Context, uri string) (*gtsmodel.Follow, error)

	// GetFollow retrieves a follow if it exists between source and target accounts.
	GetFollow(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.Follow, error)

	// GetFollowRequestByID fetches follow request with given ID from the database.
	GetFollowRequestByID(ctx context.Context, id string) (*gtsmodel.FollowRequest, error)

	// GetFollowRequestByURI fetches follow request with given AP URI from the database.
	GetFollowRequestByURI(ctx context.Context, uri string) (*gtsmodel.FollowRequest, error)

	// GetFollowRequest retrieves a follow request if it exists between source and target accounts.
	GetFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.FollowRequest, error)

	// IsFollowing returns true if sourceAccount follows target account, or an error if something goes wrong while finding out.
	IsFollowing(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, Error)

	// IsMutualFollowing returns true if account1 and account2 both follow each other, or an error if something goes wrong while finding out.
	IsMutualFollowing(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, Error)

	// IsFollowRequested returns true if sourceAccount has requested to follow target account, or an error if something goes wrong while finding out.
	IsFollowRequested(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, Error)

	// PutFollow attempts to place the given account follow in the database.
	PutFollow(ctx context.Context, follow *gtsmodel.Follow) error

	// UpdateFollow updates one follow by ID.
	UpdateFollow(ctx context.Context, follow *gtsmodel.Follow, columns ...string) error

	// PutFollowRequest attempts to place the given account follow request in the database.
	PutFollowRequest(ctx context.Context, follow *gtsmodel.FollowRequest) error

	// UpdateFollowRequest updates one follow request by ID.
	UpdateFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest, columns ...string) error

	// DeleteFollowByID deletes a follow from the database with the given ID.
	DeleteFollowByID(ctx context.Context, id string) error

	// DeleteFollowByURI deletes a follow from the database with the given URI.
	DeleteFollowByURI(ctx context.Context, uri string) error

	// DeleteFollowRequestByID deletes a follow request from the database with the given ID.
	DeleteFollowRequestByID(ctx context.Context, id string) error

	// DeleteFollowRequestByURI deletes a follow request from the database with the given URI.
	DeleteFollowRequestByURI(ctx context.Context, uri string) error

	// DeleteAccountFollows will delete all database follows to / from the given account ID.
	DeleteAccountFollows(ctx context.Context, accountID string) error

	// DeleteAccountFollowRequests will delete all database follow requests to / from the given account ID.
	DeleteAccountFollowRequests(ctx context.Context, accountID string) error

	// AcceptFollowRequest moves a follow request in the database from the follow_requests table to the follows table.
	// In other words, it should create the follow, and delete the existing follow request.
	//
	// It will return the newly created follow for further processing.
	AcceptFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.Follow, Error)

	// RejectFollowRequest fetches a follow request from the database, and then deletes it.
	RejectFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) Error

	// GetAccountFollows returns a slice of follows owned by the given accountID.
	GetAccountFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error)

	// GetAccountLocalFollows returns a slice of follows owned by the given accountID, only including follows from this instance.
	GetAccountLocalFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error)

	// CountAccountFollows returns the amount of accounts that the given accountID is following.
	CountAccountFollows(ctx context.Context, accountID string) (int, error)

	// CountAccountLocalFollows returns the amount of accounts that the given accountID is following, only including follows from this instance.
	CountAccountLocalFollows(ctx context.Context, accountID string) (int, error)

	// GetAccountFollowers fetches follows that target given accountID.
	GetAccountFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error)

	// GetAccountLocalFollowers fetches follows that target given accountID, only including follows from this instance.
	GetAccountLocalFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error)

	// CountAccountFollowers returns the amounts that the given ID is followed by.
	CountAccountFollowers(ctx context.Context, accountID string) (int, error)

	// CountAccountLocalFollowers returns the amounts that the given ID is followed by, only including follows from this instance.
	CountAccountLocalFollowers(ctx context.Context, accountID string) (int, error)

	// GetAccountFollowRequests returns all follow requests targeting the given account.
	GetAccountFollowRequests(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error)

	// GetAccountFollowRequesting returns all follow requests originating from the given account.
	GetAccountFollowRequesting(ctx context.Context, accountID string) ([]*gtsmodel.FollowRequest, error)

	// CountAccountFollowRequests returns number of follow requests targeting the given account.
	CountAccountFollowRequests(ctx context.Context, accountID string) (int, error)

	// CountAccountFollowerRequests returns number of follow requests originating from the given account.
	CountAccountFollowRequesting(ctx context.Context, accountID string) (int, error)
}
