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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Account contains functions related to account getting/setting/creation.
type Account interface {
	// GetAccountByID returns one account with the given ID, or an error if something goes wrong.
	GetAccountByID(ctx context.Context, id string) (*gtsmodel.Account, Error)

	// GetAccountByURI returns one account with the given URI, or an error if something goes wrong.
	GetAccountByURI(ctx context.Context, uri string) (*gtsmodel.Account, Error)

	// GetAccountByURL returns one account with the given URL, or an error if something goes wrong.
	GetAccountByURL(ctx context.Context, uri string) (*gtsmodel.Account, Error)

	// GetAccountByUsernameDomain returns one account with the given username and domain, or an error if something goes wrong.
	GetAccountByUsernameDomain(ctx context.Context, username string, domain string) (*gtsmodel.Account, Error)

	// GetAccountByPubkeyID returns one account with the given public key URI (ID), or an error if something goes wrong.
	GetAccountByPubkeyID(ctx context.Context, id string) (*gtsmodel.Account, Error)

	// GetAccountByInboxURI returns one account with the given inbox_uri, or an error if something goes wrong.
	GetAccountByInboxURI(ctx context.Context, uri string) (*gtsmodel.Account, Error)

	// GetAccountByOutboxURI returns one account with the given outbox_uri, or an error if something goes wrong.
	GetAccountByOutboxURI(ctx context.Context, uri string) (*gtsmodel.Account, Error)

	// GetAccountByFollowingURI returns one account with the given following_uri, or an error if something goes wrong.
	GetAccountByFollowingURI(ctx context.Context, uri string) (*gtsmodel.Account, Error)

	// GetAccountByFollowersURI returns one account with the given followers_uri, or an error if something goes wrong.
	GetAccountByFollowersURI(ctx context.Context, uri string) (*gtsmodel.Account, Error)

	// PopulateAccount ensures that all sub-models of an account are populated (e.g. avatar, header etc).
	PopulateAccount(ctx context.Context, account *gtsmodel.Account) error

	// PutAccount puts one account in the database.
	PutAccount(ctx context.Context, account *gtsmodel.Account) Error

	// UpdateAccount updates one account by ID.
	UpdateAccount(ctx context.Context, account *gtsmodel.Account, columns ...string) Error

	// DeleteAccount deletes one account from the database by its ID.
	// DO NOT USE THIS WHEN SUSPENDING ACCOUNTS! In that case you should mark the
	// account as suspended instead, rather than deleting from the db entirely.
	DeleteAccount(ctx context.Context, id string) Error

	// GetAccountCustomCSSByUsername returns the custom css of an account on this instance with the given username.
	GetAccountCustomCSSByUsername(ctx context.Context, username string) (string, Error)

	// GetAccountFaves fetches faves/likes created by the target accountID.
	GetAccountFaves(ctx context.Context, accountID string) ([]*gtsmodel.StatusFave, Error)

	// GetAccountStatusesCount is a shortcut for the common action of counting statuses produced by accountID.
	CountAccountStatuses(ctx context.Context, accountID string) (int, Error)

	// CountAccountPinned returns the total number of pinned statuses owned by account with the given id.
	CountAccountPinned(ctx context.Context, accountID string) (int, Error)

	// GetAccountStatuses is a shortcut for getting the most recent statuses. accountID is optional, if not provided
	// then all statuses will be returned. If limit is set to 0, the size of the returned slice will not be limited. This can
	// be very memory intensive so you probably shouldn't do this!
	//
	// In the case of no statuses, this function will return db.ErrNoEntries.
	GetAccountStatuses(ctx context.Context, accountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, mediaOnly bool, publicOnly bool) ([]*gtsmodel.Status, Error)

	// GetAccountPinnedStatuses returns ONLY statuses owned by the give accountID for which a corresponding StatusPin
	// exists in the database. Statuses which are not pinned will not be returned by this function.
	//
	// Statuses will be returned in the order in which they were pinned, from latest pinned to oldest pinned (descending).
	//
	// In the case of no statuses, this function will return db.ErrNoEntries.
	GetAccountPinnedStatuses(ctx context.Context, accountID string) ([]*gtsmodel.Status, Error)

	// GetAccountWebStatuses is similar to GetAccountStatuses, but it's specifically for returning statuses that
	// should be visible via the web view of an account. So, only public, federated statuses that aren't boosts
	// or replies.
	//
	// In the case of no statuses, this function will return db.ErrNoEntries.
	GetAccountWebStatuses(ctx context.Context, accountID string, limit int, maxID string) ([]*gtsmodel.Status, Error)

	GetAccountBlocks(ctx context.Context, accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, Error)

	// GetAccountLastPosted simply gets the timestamp of the most recent post by the account.
	//
	// If webOnly is true, then the time of the last non-reply, non-boost, public status of the account will be returned.
	//
	// The returned time will be zero if account has never posted anything.
	GetAccountLastPosted(ctx context.Context, accountID string, webOnly bool) (time.Time, Error)

	// SetAccountHeaderOrAvatar sets the header or avatar for the given accountID to the given media attachment.
	SetAccountHeaderOrAvatar(ctx context.Context, mediaAttachment *gtsmodel.MediaAttachment, accountID string) Error

	// GetInstanceAccount returns the instance account for the given domain.
	// If domain is empty, this instance account will be returned.
	GetInstanceAccount(ctx context.Context, domain string) (*gtsmodel.Account, Error)
}
