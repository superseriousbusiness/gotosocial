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
	DeleteAccountFollowRequests(ctx context.Context, id string) error

	// AcceptFollowRequest moves a follow request in the database from the follow_requests table to the follows table.
	// In other words, it should create the follow, and delete the existing follow request.
	//
	// It will return the newly created follow for further processing.
	AcceptFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.Follow, Error)

	// RejectFollowRequest fetches a follow request from the database, and then deletes it.
	//
	// The deleted follow request will be returned so that further processing can be done on it.
	RejectFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.FollowRequest, Error)

	// GetFollows returns a slice of follows owned by the given accountID, and/or
	// targeting the given account id.
	//
	// If accountID is set and targetAccountID isn't, then all follows created by
	// accountID will be returned.
	//
	// If targetAccountID is set and accountID isn't, then all follows targeting
	// targetAccountID will be returned.
	//
	// If both accountID and targetAccountID are set, then only 0 or 1 follows will
	// be in the returned slice.
	GetFollows(ctx context.Context, accountID string, targetAccountID string) ([]*gtsmodel.Follow, Error)

	// GetLocalFollowersIDs returns a list of local account IDs which follow the
	// targetAccountID. The returned IDs are not guaranteed to be ordered in any
	// particular way, so take care.
	GetLocalFollowersIDs(ctx context.Context, targetAccountID string) ([]string, Error)

	// CountFollows is like GetFollows, but just counts rather than returning.
	CountFollows(ctx context.Context, accountID string, targetAccountID string) (int, Error)

	// GetFollowRequests returns a slice of follows requests owned by the given
	// accountID, and/or targeting the given account id.
	//
	// If accountID is set and targetAccountID isn't, then all requests created by
	// accountID will be returned.
	//
	// If targetAccountID is set and accountID isn't, then all requests targeting
	// targetAccountID will be returned.
	//
	// If both accountID and targetAccountID are set, then only 0 or 1 requests will
	// be in the returned slice.
	GetFollowRequests(ctx context.Context, accountID string, targetAccountID string) ([]*gtsmodel.FollowRequest, Error)

	// CountFollowRequests is like GetFollowRequests, but just counts rather than returning.
	CountFollowRequests(ctx context.Context, accountID string, targetAccountID string) (int, Error)

	// Unfollow removes a follow targeting targetAccountID and originating
	// from originAccountID.
	//
	// If a follow was removed this way, the AP URI of the follow will be
	// returned to the caller, so that further processing can take place
	// if necessary.
	//
	// If no follow was removed this way, the returned string will be empty.
	Unfollow(ctx context.Context, originAccountID string, targetAccountID string) (string, Error)

	// UnfollowRequest removes a follow request targeting targetAccountID
	// and originating from originAccountID.
	//
	// If a follow request was removed this way, the AP URI of the follow
	// request will be returned to the caller, so that further processing
	// can take place if necessary.
	//
	// If no follow request was removed this way, the returned string will
	// be empty.
	UnfollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (string, Error)
}
