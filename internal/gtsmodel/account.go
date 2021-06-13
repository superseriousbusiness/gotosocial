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

// Package gtsmodel contains types used *internally* by GoToSocial and added/removed/selected from the database.
// These types should never be serialized and/or sent out via public APIs, as they contain sensitive information.
// The annotation used on these structs is for handling them via the go-pg ORM (hence why they're in this db subdir).
// See here for more info on go-pg model annotations: https://pg.uptrace.dev/models/
package gtsmodel

import (
	"crypto/rsa"
	"time"
)

// Account represents either a local or a remote fediverse account, gotosocial or otherwise (mastodon, pleroma, etc)
type Account struct {
	/*
		BASIC INFO
	*/

	// id of this account in the local database
	ID string `pg:"type:CHAR(26),pk,notnull,unique"`
	// Username of the account, should just be a string of [a-z0-9_]. Can be added to domain to create the full username in the form ``[username]@[domain]`` eg., ``user_96@example.org``
	Username string `pg:",notnull,unique:userdomain"` // username and domain should be unique *with* each other
	// Domain of the account, will be null if this is a local account, otherwise something like ``example.org`` or ``mastodon.social``. Should be unique with username.
	Domain string `pg:",unique:userdomain"` // username and domain should be unique *with* each other

	/*
		ACCOUNT METADATA
	*/

	// ID of the avatar as a media attachment
	AvatarMediaAttachmentID string `pg:"type:CHAR(26)"`
	// For a non-local account, where can the header be fetched?
	AvatarRemoteURL string
	// ID of the header as a media attachment
	HeaderMediaAttachmentID string `pg:"type:CHAR(26)"`
	// For a non-local account, where can the header be fetched?
	HeaderRemoteURL string
	// DisplayName for this account. Can be empty, then just the Username will be used for display purposes.
	DisplayName string
	// a key/value map of fields that this account has added to their profile
	Fields []Field
	// A note that this account has on their profile (ie., the account's bio/description of themselves)
	Note string
	// Is this a memorial account, ie., has the user passed away?
	Memorial bool
	// This account has moved this account id in the database
	MovedToAccountID string `pg:"type:CHAR(26)"`
	// When was this account created?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this account last updated?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Does this account identify itself as a bot?
	Bot bool
	// What reason was given for signing up when this account was created?
	Reason string

	/*
		USER AND PRIVACY PREFERENCES
	*/

	// Does this account need an approval for new followers?
	Locked bool `pg:",default:false"`
	// Should this account be shown in the instance's profile directory?
	Discoverable bool `pg:",default:false"`
	// Default post privacy for this account
	Privacy Visibility `pg:",default:'public'"`
	// Set posts from this account to sensitive by default?
	Sensitive bool `pg:",default:false"`
	// What language does this account post in?
	Language string `pg:",default:'en'"`

	/*
		ACTIVITYPUB THINGS
	*/

	// What is the activitypub URI for this account discovered by webfinger?
	URI string `pg:",unique"`
	// At which URL can we see the user account in a web browser?
	URL string `pg:",unique"`
	// Last time this account was located using the webfinger API.
	LastWebfingeredAt time.Time `pg:"type:timestamp"`
	// Address of this account's activitypub inbox, for sending activity to
	InboxURI string `pg:",unique"`
	// Address of this account's activitypub outbox
	OutboxURI string `pg:",unique"`
	// URI for getting the following list of this account
	FollowingURI string `pg:",unique"`
	// URI for getting the followers list of this account
	FollowersURI string `pg:",unique"`
	// URL for getting the featured collection list of this account
	FeaturedCollectionURI string `pg:",unique"`
	// What type of activitypub actor is this account?
	ActorType string
	// This account is associated with x account id
	AlsoKnownAs string

	/*
		CRYPTO FIELDS
	*/

	// Privatekey for validating activitypub requests, will obviously only be defined for local accounts
	PrivateKey *rsa.PrivateKey
	// Publickey for encoding activitypub requests, will be defined for both local and remote accounts
	PublicKey *rsa.PublicKey
	// Web-reachable location of this account's public key
	PublicKeyURI string

	/*
		ADMIN FIELDS
	*/

	// When was this account set to have all its media shown as sensitive?
	SensitizedAt time.Time `pg:"type:timestamp"`
	// When was this account silenced (eg., statuses only visible to followers, not public)?
	SilencedAt time.Time `pg:"type:timestamp"`
	// When was this account suspended (eg., don't allow it to log in/post, don't accept media/posts from this account)
	SuspendedAt time.Time `pg:"type:timestamp"`
	// Should we hide this account's collections?
	HideCollections bool
	// id of the user that suspended this account through an admin action
	SuspensionOrigin string
}

// Field represents a key value field on an account, for things like pronouns, website, etc.
// VerifiedAt is optional, to be used only if Value is a URL to a webpage that contains the
// username of the user.
type Field struct {
	Name       string
	Value      string
	VerifiedAt time.Time `pg:"type:timestamp"`
}
