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
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// Username of the account, should just be a string of [a-z0-9_]. Can be added to domain to create the full username in the form ``[username]@[domain]`` eg., ``user_96@example.org``
	Username string `bun:",notnull,unique:userdomain"` // username and domain should be unique *with* each other
	// Domain of the account, will be null if this is a local account, otherwise something like ``example.org`` or ``mastodon.social``. Should be unique with username.
	Domain string `bun:",unique:userdomain,nullzero"` // username and domain should be unique *with* each other

	/*
		ACCOUNT METADATA
	*/

	// ID of the avatar as a media attachment
	AvatarMediaAttachmentID string           `bun:"type:CHAR(26)"`
	AvatarMediaAttachment   *MediaAttachment `bun:"rel:belongs-to"`
	// For a non-local account, where can the header be fetched?
	AvatarRemoteURL string
	// ID of the header as a media attachment
	HeaderMediaAttachmentID string           `bun:"type:CHAR(26)"`
	HeaderMediaAttachment   *MediaAttachment `bun:"rel:belongs-to"`
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
	MovedToAccountID string `bun:"type:CHAR(26)"`
	// When was this account created?
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this account last updated?
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// Does this account identify itself as a bot?
	Bot bool
	// What reason was given for signing up when this account was created?
	Reason string

	/*
		USER AND PRIVACY PREFERENCES
	*/

	// Does this account need an approval for new followers?
	Locked bool `bun:",default:true"`
	// Should this account be shown in the instance's profile directory?
	Discoverable bool `bun:",default:false"`
	// Default post privacy for this account
	Privacy Visibility `bun:",default:'public'"`
	// Set posts from this account to sensitive by default?
	Sensitive bool `bun:",default:false"`
	// What language does this account post in?
	Language string `bun:",default:'en'"`

	/*
		ACTIVITYPUB THINGS
	*/

	// What is the activitypub URI for this account discovered by webfinger?
	URI string `bun:",unique"`
	// At which URL can we see the user account in a web browser?
	URL string `bun:",unique"`
	// Last time this account was located using the webfinger API.
	LastWebfingeredAt time.Time `bun:",nullzero"`
	// Address of this account's activitypub inbox, for sending activity to
	InboxURI string `bun:",unique"`
	// Address of this account's activitypub outbox
	OutboxURI string `bun:",unique"`
	// URI for getting the following list of this account
	FollowingURI string `bun:",unique"`
	// URI for getting the followers list of this account
	FollowersURI string `bun:",unique"`
	// URL for getting the featured collection list of this account
	FeaturedCollectionURI string `bun:",unique"`
	// What type of activitypub actor is this account?
	ActorType string
	// This account is associated with x account id
	AlsoKnownAs string

	/*
		CRYPTO FIELDS
	*/

	// Privatekey for validating activitypub requests, will only be defined for local accounts
	PrivateKey *rsa.PrivateKey
	// Publickey for encoding activitypub requests, will be defined for both local and remote accounts
	PublicKey *rsa.PublicKey
	// Web-reachable location of this account's public key
	PublicKeyURI string

	/*
		ADMIN FIELDS
	*/

	// When was this account set to have all its media shown as sensitive?
	SensitizedAt time.Time `bun:",nullzero"`
	// When was this account silenced (eg., statuses only visible to followers, not public)?
	SilencedAt time.Time `bun:",nullzero"`
	// When was this account suspended (eg., don't allow it to log in/post, don't accept media/posts from this account)
	SuspendedAt time.Time `bun:",nullzero"`
	// Should we hide this account's collections?
	HideCollections bool
	// id of the database entry that caused this account to become suspended -- can be an account ID or a domain block ID
	SuspensionOrigin string `bun:"type:CHAR(26)"`
}

// Field represents a key value field on an account, for things like pronouns, website, etc.
// VerifiedAt is optional, to be used only if Value is a URL to a webpage that contains the
// username of the user.
type Field struct {
	Name       string
	Value      string
	VerifiedAt time.Time `bun:",nullzero"`
}
