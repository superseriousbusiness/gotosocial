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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// Relationship contains functions for getting or modifying the relationship between two accounts.
type Relationship interface {
	// IsBlocked checks whether source account has a block in place against target.
	IsBlocked(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error)

	// IsEitherBlocked checks whether there is a block in place between either of account1 and account2.
	IsEitherBlocked(ctx context.Context, accountID1 string, accountID2 string) (bool, error)

	// GetBlockByID fetches block with given ID from the database.
	GetBlockByID(ctx context.Context, id string) (*gtsmodel.Block, error)

	// GetBlockByURI fetches block with given AP URI from the database.
	GetBlockByURI(ctx context.Context, uri string) (*gtsmodel.Block, error)

	// GetBlock returns the block from account1 targeting account2, if it exists, or an error if it doesn't.
	GetBlock(ctx context.Context, account1 string, account2 string) (*gtsmodel.Block, error)

	// PopulateBlock populates the struct pointers on the given block.
	PopulateBlock(ctx context.Context, block *gtsmodel.Block) error

	// PutBlock attempts to place the given account block in the database.
	PutBlock(ctx context.Context, block *gtsmodel.Block) error

	// DeleteBlockByID removes block with given ID from the database.
	DeleteBlockByID(ctx context.Context, id string) error

	// DeleteBlockByURI removes block with given AP URI from the database.
	DeleteBlockByURI(ctx context.Context, uri string) error

	// DeleteAccountBlocks will delete all database blocks to / from the given account ID.
	DeleteAccountBlocks(ctx context.Context, accountID string) error

	// GetRelationship retrieves the relationship of the targetAccount to the requestingAccount.
	GetRelationship(ctx context.Context, requestingAccount string, targetAccount string) (*gtsmodel.Relationship, error)

	// GetFollowByID fetches follow with given ID from the database.
	GetFollowByID(ctx context.Context, id string) (*gtsmodel.Follow, error)

	// GetFollowByURI fetches follow with given AP URI from the database.
	GetFollowByURI(ctx context.Context, uri string) (*gtsmodel.Follow, error)

	// GetFollow retrieves a follow if it exists between source and target accounts.
	GetFollow(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.Follow, error)

	// GetFollowsByIDs fetches all follows from database with given IDs.
	GetFollowsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Follow, error)

	// PopulateFollow populates the struct pointers on the given follow.
	PopulateFollow(ctx context.Context, follow *gtsmodel.Follow) error

	// GetFollowRequestByID fetches follow request with given ID from the database.
	GetFollowRequestByID(ctx context.Context, id string) (*gtsmodel.FollowRequest, error)

	// GetFollowRequestByURI fetches follow request with given AP URI from the database.
	GetFollowRequestByURI(ctx context.Context, uri string) (*gtsmodel.FollowRequest, error)

	// GetFollowRequest retrieves a follow request if it exists between source and target accounts.
	GetFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.FollowRequest, error)

	// PopulateFollowRequest populates the struct pointers on the given follow request.
	PopulateFollowRequest(ctx context.Context, follow *gtsmodel.FollowRequest) error

	// IsFollowing returns true if sourceAccount follows target account, or an error if something goes wrong while finding out.
	IsFollowing(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error)

	// IsMutualFollowing returns true if account1 and account2 both follow each other, or an error if something goes wrong while finding out.
	IsMutualFollowing(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error)

	// IsFollowRequested returns true if sourceAccount has requested to follow target account, or an error if something goes wrong while finding out.
	IsFollowRequested(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error)

	// PutFollow attempts to place the given account follow in the database.
	PutFollow(ctx context.Context, follow *gtsmodel.Follow) error

	// UpdateFollow updates one follow by ID.
	UpdateFollow(ctx context.Context, follow *gtsmodel.Follow, columns ...string) error

	// PutFollowRequest attempts to place the given account follow request in the database.
	PutFollowRequest(ctx context.Context, follow *gtsmodel.FollowRequest) error

	// UpdateFollowRequest updates one follow request by ID.
	UpdateFollowRequest(ctx context.Context, followRequest *gtsmodel.FollowRequest, columns ...string) error

	// DeleteFollow deletes a follow if it exists between source and target accounts.
	DeleteFollow(ctx context.Context, sourceAccountID string, targetAccountID string) error

	// DeleteFollowByID deletes a follow from the database with the given ID.
	DeleteFollowByID(ctx context.Context, id string) error

	// DeleteFollowByURI deletes a follow from the database with the given URI.
	DeleteFollowByURI(ctx context.Context, uri string) error

	// DeleteFollowRequest deletes a follow request if it exists between source and target accounts.
	DeleteFollowRequest(ctx context.Context, sourceAccountID string, targetAccountID string) error

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
	AcceptFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) (*gtsmodel.Follow, error)

	// RejectFollowRequest fetches a follow request from the database, and then deletes it.
	RejectFollowRequest(ctx context.Context, originAccountID string, targetAccountID string) error

	// GetAccountFollows returns a slice of follows owned by the given accountID.
	GetAccountFollows(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Follow, error)

	// GetAccountFollowIDs is like GetAccountFollows, but returns just IDs.
	GetAccountFollowIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error)

	// GetAccountLocalFollows returns a slice of follows owned by the given accountID, only including follows from this instance.
	GetAccountLocalFollows(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error)

	// GetAccountLocalFollowIDs is like GetAccountLocalFollows, but returns just IDs.
	GetAccountLocalFollowIDs(ctx context.Context, accountID string) ([]string, error)

	// GetAccountFollowers fetches follows that target given accountID.
	GetAccountFollowers(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Follow, error)

	// GetAccountFollowerIDs is like GetAccountFollowers, but returns just IDs.
	GetAccountFollowerIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error)

	// GetAccountLocalFollowers fetches follows that target given accountID, only including follows from this instance.
	GetAccountLocalFollowers(ctx context.Context, accountID string) ([]*gtsmodel.Follow, error)

	// GetAccountLocalFollowerIDs is like GetAccountLocalFollowers, but returns just IDs.
	GetAccountLocalFollowerIDs(ctx context.Context, accountID string) ([]string, error)

	// GetAccountFollowRequests returns all follow requests targeting the given account.
	GetAccountFollowRequests(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.FollowRequest, error)

	// GetAccountFollowRequestIDs is like GetAccountFollowRequests, but returns just IDs.
	GetAccountFollowRequestIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error)

	// GetAccountFollowRequesting returns all follow requests originating from the given account.
	GetAccountFollowRequesting(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.FollowRequest, error)

	// GetAccountFollowRequestingIDs is like GetAccountFollowRequesting, but returns just IDs.
	GetAccountFollowRequestingIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error)

	// GetAccountBlocks returns all blocks originating from the given account, with given optional paging parameters.
	GetAccountBlocks(ctx context.Context, accountID string, paging *paging.Page) ([]*gtsmodel.Block, error)

	// GetAccountBlockIDs is like GetAccountBlocks, but returns just IDs.
	GetAccountBlockIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error)

	// CountAccountBlocks counts the number of blocks owned by the given account.
	CountAccountBlocks(ctx context.Context, accountID string) (int, error)

	// GetNote gets a private note from a source account on a target account, if it exists.
	GetNote(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.AccountNote, error)

	// PutNote creates or updates a private note.
	PutNote(ctx context.Context, note *gtsmodel.AccountNote) error

	// PopulateNote populates the struct pointers on the given note.
	PopulateNote(ctx context.Context, note *gtsmodel.AccountNote) error

	// IsMuted checks whether source account has a mute in place against target.
	IsMuted(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error)

	// GetMuteByID fetches mute with given ID from the database.
	GetMuteByID(ctx context.Context, id string) (*gtsmodel.UserMute, error)

	// GetMute returns the mute from account1 targeting account2, if it exists, or an error if it doesn't.
	GetMute(ctx context.Context, account1 string, account2 string) (*gtsmodel.UserMute, error)

	// CountAccountMutes counts the number of mutes owned by the given account.
	CountAccountMutes(ctx context.Context, accountID string) (int, error)

	// PutMute attempts to insert or update the given account mute in the database.
	PutMute(ctx context.Context, mute *gtsmodel.UserMute) error

	// DeleteMuteByID removes mute with given ID from the database.
	DeleteMuteByID(ctx context.Context, id string) error

	// DeleteAccountMutes will delete all database mutes to / from the given account ID.
	DeleteAccountMutes(ctx context.Context, accountID string) error

	// GetAccountMutes returns all mutes originating from the given account, with given optional paging parameters.
	GetAccountMutes(ctx context.Context, accountID string, paging *paging.Page) ([]*gtsmodel.UserMute, error)
}
