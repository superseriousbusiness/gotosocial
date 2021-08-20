/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Account contains functions related to account getting/setting/creation.
type Account interface {
	// GetAccountByID returns one account with the given ID, or an error if something goes wrong.
	GetAccountByID(id string) (*gtsmodel.Account, Error)

	// GetAccountByURI returns one account with the given URI, or an error if something goes wrong.
	GetAccountByURI(uri string) (*gtsmodel.Account, Error)

	// GetAccountByURL returns one account with the given URL, or an error if something goes wrong.
	GetAccountByURL(uri string) (*gtsmodel.Account, Error)

	// GetLocalAccountByUsername returns an account on this instance by its username.
	GetLocalAccountByUsername(username string) (*gtsmodel.Account, Error)

	// GetAccountFaves fetches faves/likes created by the target accountID.
	GetAccountFaves(accountID string) ([]*gtsmodel.StatusFave, Error)

	// GetAccountStatusesCount is a shortcut for the common action of counting statuses produced by accountID.
	CountAccountStatuses(accountID string) (int, Error)

	// GetAccountStatuses is a shortcut for getting the most recent statuses. accountID is optional, if not provided
	// then all statuses will be returned. If limit is set to 0, the size of the returned slice will not be limited. This can
	// be very memory intensive so you probably shouldn't do this!
	// In case of no entries, a 'no entries' error will be returned
	GetAccountStatuses(accountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]*gtsmodel.Status, Error)

	GetAccountBlocks(accountID string, maxID string, sinceID string, limit int) ([]*gtsmodel.Account, string, string, Error)

	// GetAccountLastPosted simply gets the timestamp of the most recent post by the account.
	//
	// The returned time will be zero if account has never posted anything.
	GetAccountLastPosted(accountID string) (time.Time, Error)

	// SetAccountHeaderOrAvatar sets the header or avatar for the given accountID to the given media attachment.
	SetAccountHeaderOrAvatar(mediaAttachment *gtsmodel.MediaAttachment, accountID string) Error

	// GetInstanceAccount returns the instance account for the given domain.
	// If domain is empty, this instance account will be returned.
	GetInstanceAccount(domain string) (*gtsmodel.Account, Error)
}
