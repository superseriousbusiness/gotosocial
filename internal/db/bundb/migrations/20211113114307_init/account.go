/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
// The annotation used on these structs is for handling them via the bun-db ORM.
// See here for more info on bun model annotations: https://bun.uptrace.dev/guide/models.html
package gtsmodel

import (
	"crypto/rsa"
	"time"
)

// Account represents either a local or a remote fediverse account, gotosocial or otherwise (mastodon, pleroma, etc).
type Account struct {
	ID                      string           `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                               // id of this item in the database
	CreatedAt               time.Time        `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                        // when was item created
	UpdatedAt               time.Time        `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                        // when was item last updated
	Username                string           `validate:"required" bun:",nullzero,notnull,unique:userdomain"`                                                         // Username of the account, should just be a string of [a-zA-Z0-9_]. Can be added to domain to create the full username in the form ``[username]@[domain]`` eg., ``user_96@example.org``. Username and domain should be unique *with* each other
	Domain                  string           `validate:"omitempty,fqdn" bun:",nullzero,unique:userdomain"`                                                           // Domain of the account, will be null if this is a local account, otherwise something like ``example.org``. Should be unique with username.
	AvatarMediaAttachmentID string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                                                // Database ID of the media attachment, if present
	AvatarMediaAttachment   *MediaAttachment `validate:"-" bun:"rel:belongs-to"`                                                                                     // MediaAttachment corresponding to avatarMediaAttachmentID
	AvatarRemoteURL         string           `validate:"omitempty,url" bun:",nullzero"`                                                                              // For a non-local account, where can the header be fetched?
	HeaderMediaAttachmentID string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                                                // Database ID of the media attachment, if present
	HeaderMediaAttachment   *MediaAttachment `validate:"-" bun:"rel:belongs-to"`                                                                                     // MediaAttachment corresponding to headerMediaAttachmentID
	HeaderRemoteURL         string           `validate:"omitempty,url" bun:",nullzero"`                                                                              // For a non-local account, where can the header be fetched?
	DisplayName             string           `validate:"-" bun:""`                                                                                                   // DisplayName for this account. Can be empty, then just the Username will be used for display purposes.
	Fields                  []Field          `validate:"-"`                                                                                                          // a key/value map of fields that this account has added to their profile
	Note                    string           `validate:"-" bun:""`                                                                                                   // A note that this account has on their profile (ie., the account's bio/description of themselves)
	Memorial                bool             `validate:"-" bun:",default:false"`                                                                                     // Is this a memorial account, ie., has the user passed away?
	AlsoKnownAs             string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                                                // This account is associated with x account id (TODO: migrate to be AlsoKnownAsID)
	MovedToAccountID        string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                                                // This account has moved this account id in the database
	Bot                     bool             `validate:"-" bun:",default:false"`                                                                                     // Does this account identify itself as a bot?
	Reason                  string           `validate:"-" bun:""`                                                                                                   // What reason was given for signing up when this account was created?
	Locked                  bool             `validate:"-" bun:",default:true"`                                                                                      // Does this account need an approval for new followers?
	Discoverable            bool             `validate:"-" bun:",default:false"`                                                                                     // Should this account be shown in the instance's profile directory?
	Privacy                 Visibility       `validate:"required_without=Domain,omitempty,oneof=public unlocked followers_only mutuals_only direct" bun:",nullzero"` // Default post privacy for this account
	Sensitive               bool             `validate:"-" bun:",default:false"`                                                                                     // Set posts from this account to sensitive by default?
	Language                string           `validate:"omitempty,bcp47_language_tag" bun:",nullzero,notnull,default:'en'"`                                          // What language does this account post in?
	URI                     string           `validate:"required,url" bun:",nullzero,notnull,unique"`                                                                // ActivityPub URI for this account.
	URL                     string           `validate:"required_without=Domain,omitempty,url" bun:",nullzero,unique"`                                               // Web URL for this account's profile
	LastWebfingeredAt       time.Time        `validate:"required_with=Domain" bun:"type:timestamptz,nullzero"`                                                       // Last time this account was refreshed/located with webfinger.
	InboxURI                string           `validate:"required_without=Domain,omitempty,url" bun:",nullzero,unique"`                                               // Address of this account's ActivityPub inbox, for sending activity to
	OutboxURI               string           `validate:"required_without=Domain,omitempty,url" bun:",nullzero,unique"`                                               // Address of this account's activitypub outbox
	FollowingURI            string           `validate:"required_without=Domain,omitempty,url" bun:",nullzero,unique"`                                               // URI for getting the following list of this account
	FollowersURI            string           `validate:"required_without=Domain,omitempty,url" bun:",nullzero,unique"`                                               // URI for getting the followers list of this account
	FeaturedCollectionURI   string           `validate:"required_without=Domain,omitempty,url" bun:",nullzero,unique"`                                               // URL for getting the featured collection list of this account
	ActorType               string           `validate:"oneof=Application Group Organization Person Service" bun:",nullzero,notnull"`                                // What type of activitypub actor is this account?
	PrivateKey              *rsa.PrivateKey  `validate:"required_without=Domain"`                                                                                    // Privatekey for validating activitypub requests, will only be defined for local accounts
	PublicKey               *rsa.PublicKey   `validate:"required"`                                                                                                   // Publickey for encoding activitypub requests, will be defined for both local and remote accounts
	PublicKeyURI            string           `validate:"required,url" bun:",nullzero,notnull,unique"`                                                                // Web-reachable location of this account's public key
	SensitizedAt            time.Time        `validate:"-" bun:"type:timestamptz,nullzero"`                                                                          // When was this account set to have all its media shown as sensitive?
	SilencedAt              time.Time        `validate:"-" bun:"type:timestamptz,nullzero"`                                                                          // When was this account silenced (eg., statuses only visible to followers, not public)?
	SuspendedAt             time.Time        `validate:"-" bun:"type:timestamptz,nullzero"`                                                                          // When was this account suspended (eg., don't allow it to log in/post, don't accept media/posts from this account)
	HideCollections         bool             `validate:"-" bun:",default:false"`                                                                                     // Hide this account's collections
	SuspensionOrigin        string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                                                // id of the database entry that caused this account to become suspended -- can be an account ID or a domain block ID
}

// Field represents a key value field on an account, for things like pronouns, website, etc.
// VerifiedAt is optional, to be used only if Value is a URL to a webpage that contains the
// username of the user.
type Field struct {
	Name       string    `validate:"required"`          // Name of this field.
	Value      string    `validate:"required"`          // Value of this field.
	VerifiedAt time.Time `validate:"-" bun:",nullzero"` // This field was verified at (optional).
}

// Relationship describes a requester's relationship with another account.
type Relationship struct {
	ID                  string // The account id.
	Following           bool   // Are you following this user?
	ShowingReblogs      bool   // Are you receiving this user's boosts in your home timeline?
	Notifying           bool   // Have you enabled notifications for this user?
	FollowedBy          bool   // Are you followed by this user?
	Blocking            bool   // Are you blocking this user?
	BlockedBy           bool   // Is this user blocking you?
	Muting              bool   // Are you muting this user?
	MutingNotifications bool   // Are you muting notifications from this user?
	Requested           bool   // Do you have a pending follow request for this user?
	DomainBlocking      bool   // Are you blocking this user's domain?
	Endorsed            bool   // Are you featuring this user on your profile?
	Note                string // Your note on this account.
}
