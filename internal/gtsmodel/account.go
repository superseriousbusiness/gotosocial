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

// package gtsmodel contains types used *internally* by GoToSocial and added/removed/selected from the database.
// These types should never be serialized and/or sent out via public APIs, as they contain sensitive information.
// The annotation used on these structs is for handling them via the go-pg ORM. See here: https://pg.uptrace.dev/models/
package gtsmodel

import (
	"net/url"
	"time"
)

// Account represents either a local or a remote fediverse account, gotosocial or otherwise (mastodon, pleroma, etc)
type Account struct {
	/*
		BASIC INFO
	*/

	// id of this account in the local database; the end-user will never need to know this, it's strictly internal
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// Username of the account, should just be a string of [a-z0-9_]. Can be added to domain to create the full username in the form ``[username]@[domain]`` eg., ``user_96@example.org``
	Username string `pg:",notnull,unique:userdomain"` // username and domain should be unique *with* each other
	// Domain of the account, will be empty if this is a local account, otherwise something like ``example.org`` or ``mastodon.social``. Should be unique with username.
	Domain string `pg:",unique:userdomain"` // username and domain

	/*
		ACCOUNT METADATA
	*/

	// Avatar image for this account
	Avatar
	// Header image for this account
	Header
	// DisplayName for this account. Can be empty, then just the Username will be used for display purposes.
	DisplayName string
	// a key/value map of fields that this account has added to their profile
	Fields map[string]string
	// A note that this account has on their profile (ie., the account's bio/description of themselves)
	Note string
	// Is this a memorial account, ie., has the user passed away?
	Memorial bool
	// This account has moved this account id in the database
	MovedToAccountID int
	// When was this account created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this account last updated?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When should this account function until
	SubscriptionExpiresAt time.Time `pg:"type:timestamp"`

	/*
		PRIVACY SETTINGS
	*/

	// Does this account need an approval for new followers?
	Locked bool
	// Should this account be shown in the instance's profile directory?
	Discoverable bool

	/*
		ACTIVITYPUB THINGS
	*/

	// What is the activitypub URI for this account discovered by webfinger?
	URI string `pg:",unique"`
	// At which URL can we see the user account in a web browser?
	URL string `pg:",unique"`
	// RemoteURL where this account is located. Will be empty if this is a local account.
	RemoteURL string `pg:",unique"`
	// Last time this account was located using the webfinger API.
	LastWebfingeredAt time.Time `pg:"type:timestamp"`
	// Address of this account's activitypub inbox, for sending activity to
	InboxURL string `pg:",unique"`
	// Address of this account's activitypub outbox
	OutboxURL string `pg:",unique"`
	// Don't support shared inbox right now so this is just a stub for a future implementation
	SharedInboxURL string `pg:",unique"`
	// URL for getting the followers list of this account
	FollowersURL string `pg:",unique"`
	// URL for getting the featured collection list of this account
	FeaturedCollectionURL string `pg:",unique"`
	// What type of activitypub actor is this account?
	ActorType string
	// This account is associated with x account id
	AlsoKnownAs string

	/*
		CRYPTO FIELDS
	*/

	Secret string
	// Privatekey for validating activitypub requests, will obviously only be defined for local accounts
	PrivateKey string
	// Publickey for encoding activitypub requests, will be defined for both local and remote accounts
	PublicKey string

	/*
		ADMIN FIELDS
	*/

	// When was this account set to have all its media shown as sensitive?
	SensitizedAt time.Time `pg:"type:timestamp"`
	// When was this account silenced (eg., statuses only visible to followers, not public)?
	SilencedAt time.Time `pg:"type:timestamp"`
	// When was this account suspended (eg., don't allow it to log in/post, don't accept media/posts from this account)
	SuspendedAt time.Time `pg:"type:timestamp"`
	// How much do we trust this account 🤔
	TrustLevel  int
	// Should we hide this account's collections?
	HideCollections bool
	// id of the user that suspended this account through an admin action
	SuspensionOrigin int
}

// Avatar represents the avatar for the account for display purposes
type Avatar struct {
	// File name of the avatar on local storage
	AvatarFileName string
	// Gif? png? jpeg?
	AvatarContentType string
	AvatarFileSize    int
	AvatarUpdatedAt   *time.Time `pg:"type:timestamp"`
	// Where can we retrieve the avatar?
	AvatarRemoteURL            *url.URL `pg:"type:text"`
	AvatarStorageSchemaVersion int
}

// Header represents the header of the account for display purposes
type Header struct {
	// File name of the header on local storage
	HeaderFileName string
	// Gif? png? jpeg?
	HeaderContentType string
	HeaderFileSize    int
	HeaderUpdatedAt   *time.Time `pg:"type:timestamp"`
	// Where can we retrieve the header?
	HeaderRemoteURL            *url.URL `pg:"type:text"`
	HeaderStorageSchemaVersion int
}
