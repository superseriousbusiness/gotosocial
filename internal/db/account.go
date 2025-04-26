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
	"net/netip"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// Account contains functions related to account getting/setting/creation.
type Account interface {
	// GetAccountByID returns one account with the given ID.
	GetAccountByID(ctx context.Context, id string) (*gtsmodel.Account, error)

	// GetAccountsByIDs returns accounts corresponding to given IDs.
	GetAccountsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Account, error)

	// GetAccountByURI returns one account with the given ActivityStreams URI.
	GetAccountByURI(ctx context.Context, uri string) (*gtsmodel.Account, error)

	// GetOneAccountByURL returns *one* account with the given ActivityStreams URL.
	// If more than one account has the given url, ErrMultipleEntries will be returned.
	GetOneAccountByURL(ctx context.Context, url string) (*gtsmodel.Account, error)

	// GetAccountsByURL returns accounts with the given ActivityStreams URL.
	GetAccountsByURL(ctx context.Context, url string) ([]*gtsmodel.Account, error)

	// GetAccountByUsernameDomain returns one account with the given username and domain.
	GetAccountByUsernameDomain(ctx context.Context, username string, domain string) (*gtsmodel.Account, error)

	// GetAccountByPubkeyID returns one account with the given public key URI (ID).
	GetAccountByPubkeyID(ctx context.Context, id string) (*gtsmodel.Account, error)

	// GetOneAccountByInboxURI returns one account with the given inbox_uri.
	// If more than one account has the given URL, ErrMultipleEntries will be returned.
	GetOneAccountByInboxURI(ctx context.Context, uri string) (*gtsmodel.Account, error)

	// GetOneAccountByOutboxURI returns one account with the given outbox_uri.
	// If more than one account has the given uri, ErrMultipleEntries will be returned.
	GetOneAccountByOutboxURI(ctx context.Context, uri string) (*gtsmodel.Account, error)

	// GetAccountsByMovedToURI returns any accounts with given moved_to_uri set.
	GetAccountsByMovedToURI(ctx context.Context, uri string) ([]*gtsmodel.Account, error)

	// GetAccounts returns accounts
	// with the given parameters.
	GetAccounts(
		ctx context.Context,
		origin string,
		status string,
		mods bool,
		invitedBy string,
		username string,
		displayName string,
		domain string,
		email string,
		ip netip.Addr,
		page *paging.Page,
	) (
		[]*gtsmodel.Account,
		error,
	)

	// PopulateAccount ensures that all sub-models of an account are populated (e.g. avatar, header etc).
	PopulateAccount(ctx context.Context, account *gtsmodel.Account) error

	// PutAccount puts one account in the database.
	PutAccount(ctx context.Context, account *gtsmodel.Account) error

	// UpdateAccount updates one account by ID.
	UpdateAccount(ctx context.Context, account *gtsmodel.Account, columns ...string) error

	// DeleteAccount deletes one account from the database by its ID.
	// DO NOT USE THIS WHEN SUSPENDING ACCOUNTS! In that case you should mark the
	// account as suspended instead, rather than deleting from the db entirely.
	DeleteAccount(ctx context.Context, id string) error

	// GetAccountCustomCSSByUsername returns the custom css of an account on this instance with the given username.
	GetAccountCustomCSSByUsername(ctx context.Context, username string) (string, error)

	// GetAccountFaves fetches faves/likes created by the target accountID.
	GetAccountFaves(ctx context.Context, accountID string) ([]*gtsmodel.StatusFave, error)

	// GetAccountsUsingEmoji fetches all account models using emoji with given ID stored in their 'emojis' column.
	GetAccountsUsingEmoji(ctx context.Context, emojiID string) ([]*gtsmodel.Account, error)

	// GetAccountStatuses is a shortcut for getting the most recent statuses. accountID is optional, if not provided
	// then all statuses will be returned. If limit is set to 0, the size of the returned slice will not be limited. This can
	// be very memory intensive so you probably shouldn't do this!
	//
	// In the case of no statuses, this function will return db.ErrNoEntries.
	GetAccountStatuses(ctx context.Context, accountID string, limit int, excludeReplies bool, excludeReblogs bool, maxID string, minID string, mediaOnly bool, publicOnly bool) ([]*gtsmodel.Status, error)

	// GetAccountPinnedStatuses returns ONLY statuses owned by the give accountID for which a corresponding StatusPin
	// exists in the database. Statuses which are not pinned will not be returned by this function.
	//
	// Statuses will be returned in the order in which they were pinned, from latest pinned to oldest pinned (descending).
	//
	// In the case of no statuses, this function will return db.ErrNoEntries.
	GetAccountPinnedStatuses(ctx context.Context, accountID string) ([]*gtsmodel.Status, error)

	// GetAccountWebStatuses is similar to GetAccountStatuses, but it's specifically for
	// returning statuses that should be visible via the web view of a *LOCAL* account.
	//
	// In the case of no statuses, this function will return db.ErrNoEntries.
	GetAccountWebStatuses(ctx context.Context, account *gtsmodel.Account, mediaOnly bool, limit int, maxID string) ([]*gtsmodel.Status, error)

	// GetInstanceAccount returns the instance account for the given domain.
	// If domain is empty, this instance account will be returned.
	GetInstanceAccount(ctx context.Context, domain string) (*gtsmodel.Account, error)

	// Get local account settings with the given ID.
	GetAccountSettings(ctx context.Context, id string) (*gtsmodel.AccountSettings, error)

	// Store local account settings.
	PutAccountSettings(ctx context.Context, settings *gtsmodel.AccountSettings) error

	// Update local account settings.
	UpdateAccountSettings(ctx context.Context, settings *gtsmodel.AccountSettings, columns ...string) error

	// PopulateAccountStats either creates account stats for the given
	// account by performing COUNT(*) database queries, or retrieves
	// existing stats from the database, and attaches stats to account.
	//
	// If account is local and stats were last regenerated > 48 hours ago,
	// stats will always be regenerated using COUNT(*) queries, to prevent drift.
	PopulateAccountStats(ctx context.Context, account *gtsmodel.Account) error

	// StubAccountStats creates zeroed account stats for the given account,
	// skipping COUNT(*) queries, upserts them in the DB, and attaches them
	// to the account model.
	//
	// Useful following fresh dereference of a remote account, or fresh
	// creation of a local account, when you know all COUNT(*) queries
	// would return 0 anyway.
	StubAccountStats(ctx context.Context, account *gtsmodel.Account) error

	// RegenerateAccountStats creates, upserts, and returns stats
	// for the given account, and attaches them to the account model.
	//
	// Unlike GetAccountStats, it will always get the database stats fresh.
	// This can be used to "refresh" stats.
	//
	// Because this involves database calls that can be expensive (on Postgres
	// specifically), callers should prefer GetAccountStats in 99% of cases.
	RegenerateAccountStats(ctx context.Context, account *gtsmodel.Account) error

	// Update account stats.
	UpdateAccountStats(ctx context.Context, stats *gtsmodel.AccountStats, columns ...string) error

	// DeleteAccountStats deletes the accountStats entry for the given accountID.
	DeleteAccountStats(ctx context.Context, accountID string) error
}
