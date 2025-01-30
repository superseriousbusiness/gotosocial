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

// Package gtsmodel contains types used *internally* by GoToSocial and added/removed/selected from the database.
// These types should never be serialized and/or sent out via public APIs, as they contain sensitive information.
// The annotation used on these structs is for handling them via the bun-db ORM.
// See here for more info on bun model annotations: https://bun.uptrace.dev/guide/models.html
package gtsmodel

import (
	"crypto/rsa"
	"slices"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Account represents either a local or a remote fediverse
// account, gotosocial or otherwise (mastodon, pleroma, etc).
type Account struct {
	ID                      string           `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt               time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created.
	UpdatedAt               time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item was last updated.
	FetchedAt               time.Time        `bun:"type:timestamptz,nullzero"`                                   // when was item (remote) last fetched.
	Username                string           `bun:",nullzero,notnull,unique:usernamedomain"`                     // Username of the account, should just be a string of [a-zA-Z0-9_]. Can be added to domain to create the full username in the form ``[username]@[domain]`` eg., ``user_96@example.org``. Username and domain should be unique *with* each other
	Domain                  string           `bun:",nullzero,unique:usernamedomain"`                             // Domain of the account, will be null if this is a local account, otherwise something like ``example.org``. Should be unique with username.
	AvatarMediaAttachmentID string           `bun:"type:CHAR(26),nullzero"`                                      // Database ID of the media attachment, if present
	AvatarMediaAttachment   *MediaAttachment `bun:"rel:belongs-to"`                                              // MediaAttachment corresponding to avatarMediaAttachmentID
	AvatarRemoteURL         string           `bun:",nullzero"`                                                   // For a non-local account, where can the header be fetched?
	HeaderMediaAttachmentID string           `bun:"type:CHAR(26),nullzero"`                                      // Database ID of the media attachment, if present
	HeaderMediaAttachment   *MediaAttachment `bun:"rel:belongs-to"`                                              // MediaAttachment corresponding to headerMediaAttachmentID
	HeaderRemoteURL         string           `bun:",nullzero"`                                                   // For a non-local account, where can the header be fetched?
	DisplayName             string           `bun:""`                                                            // DisplayName for this account. Can be empty, then just the Username will be used for display purposes.
	EmojiIDs                []string         `bun:"emojis,array"`                                                // Database IDs of any emojis used in this account's bio, display name, etc
	Emojis                  []*Emoji         `bun:"attached_emojis,m2m:account_to_emojis"`                       // Emojis corresponding to emojiIDs. https://bun.uptrace.dev/guide/relations.html#many-to-many-relation
	Fields                  []*Field         `bun:""`                                                            // A slice of of fields that this account has added to their profile.
	FieldsRaw               []*Field         `bun:""`                                                            // The raw (unparsed) content of fields that this account has added to their profile, without conversion to HTML, only available when requester = target
	Note                    string           `bun:""`                                                            // A note that this account has on their profile (ie., the account's bio/description of themselves)
	NoteRaw                 string           `bun:""`                                                            // The raw contents of .Note without conversion to HTML, only available when requester = target
	Memorial                *bool            `bun:",default:false"`                                              // Is this a memorial account, ie., has the user passed away?
	AlsoKnownAsURIs         []string         `bun:"also_known_as_uris,array"`                                    // This account is associated with these account URIs.
	AlsoKnownAs             []*Account       `bun:"-"`                                                           // This account is associated with these accounts (field not stored in the db).
	MovedToURI              string           `bun:",nullzero"`                                                   // This account has (or claims to have) moved to this account URI. Even if this field is set the move may not yet have been processed. Check `move` for this.
	MovedTo                 *Account         `bun:"-"`                                                           // This account has moved to this account (field not stored in the db).
	MoveID                  string           `bun:"type:CHAR(26),nullzero"`                                      // ID of a Move in the database for this account. Only set if we received or created a Move activity for which this account URI was the origin.
	Move                    *Move            `bun:"-"`                                                           // Move corresponding to MoveID, if set.
	Bot                     *bool            `bun:",default:false"`                                              // Does this account identify itself as a bot?
	Locked                  *bool            `bun:",default:true"`                                               // Does this account need an approval for new followers?
	Discoverable            *bool            `bun:",default:false"`                                              // Should this account be shown in the instance's profile directory?
	URI                     string           `bun:",nullzero,notnull,unique"`                                    // ActivityPub URI for this account.
	URL                     string           `bun:",nullzero,unique"`                                            // Web URL for this account's profile
	InboxURI                string           `bun:",nullzero,unique"`                                            // Address of this account's ActivityPub inbox, for sending activity to
	SharedInboxURI          *string          `bun:""`                                                            // Address of this account's ActivityPub sharedInbox. Gotcha warning: this is a string pointer because it has three possible states: 1. We don't know yet if the account has a shared inbox -- null. 2. We know it doesn't have a shared inbox -- empty string. 3. We know it does have a shared inbox -- url string.
	OutboxURI               string           `bun:",nullzero,unique"`                                            // Address of this account's activitypub outbox
	FollowingURI            string           `bun:",nullzero,unique"`                                            // URI for getting the following list of this account
	FollowersURI            string           `bun:",nullzero,unique"`                                            // URI for getting the followers list of this account
	FeaturedCollectionURI   string           `bun:",nullzero,unique"`                                            // URL for getting the featured collection list of this account
	ActorType               string           `bun:",nullzero,notnull"`                                           // What type of activitypub actor is this account?
	PrivateKey              *rsa.PrivateKey  `bun:""`                                                            // Privatekey for signing activitypub requests, will only be defined for local accounts
	PublicKey               *rsa.PublicKey   `bun:",notnull"`                                                    // Publickey for authorizing signed activitypub requests, will be defined for both local and remote accounts
	PublicKeyURI            string           `bun:",nullzero,notnull,unique"`                                    // Web-reachable location of this account's public key
	PublicKeyExpiresAt      time.Time        `bun:"type:timestamptz,nullzero"`                                   // PublicKey will expire/has expired at given time, and should be fetched again as appropriate. Only ever set for remote accounts.
	SensitizedAt            time.Time        `bun:"type:timestamptz,nullzero"`                                   // When was this account set to have all its media shown as sensitive?
	SilencedAt              time.Time        `bun:"type:timestamptz,nullzero"`                                   // When was this account silenced (eg., statuses only visible to followers, not public)?
	SuspendedAt             time.Time        `bun:"type:timestamptz,nullzero"`                                   // When was this account suspended (eg., don't allow it to log in/post, don't accept media/posts from this account)
	SuspensionOrigin        string           `bun:"type:CHAR(26),nullzero"`                                      // id of the database entry that caused this account to become suspended -- can be an account ID or a domain block ID
	Settings                *AccountSettings `bun:"-"`                                                           // gtsmodel.AccountSettings for this account.
	Stats                   *AccountStats    `bun:"-"`                                                           // gtsmodel.AccountStats for this account.
}

// UsernameDomain returns account @username@domain (missing domain if local).
func (a *Account) UsernameDomain() string {
	if a.IsLocal() {
		return "@" + a.Username
	}
	return "@" + a.Username + "@" + a.Domain
}

// IsLocal returns whether account is a local user account.
func (a *Account) IsLocal() bool {
	return a.Domain == "" ||
		a.Domain == config.GetHost() ||
		a.Domain == config.GetAccountDomain()
}

// IsRemote returns whether account is a remote user account.
func (a *Account) IsRemote() bool {
	return !a.IsLocal()
}

// IsNew returns whether an account is "new" in the sense
// that it has not been previously stored in the database.
func (a *Account) IsNew() bool {
	return a.CreatedAt.IsZero()
}

// IsInstance returns whether account is an instance internal actor account.
func (a *Account) IsInstance() bool {
	if a.IsLocal() {
		// Check if our instance account.
		return a.Username == config.GetHost()
	}

	// Check if remote instance account.
	return a.Username == a.Domain ||
		a.FollowersURI == "" ||
		a.FollowingURI == "" ||
		(a.Username == "internal.fetch" && strings.Contains(a.Note, "internal service actor")) ||
		a.Username == "instance.actor" // <- misskey
}

// EmojisPopulated returns whether emojis are
// populated according to current EmojiIDs.
func (a *Account) EmojisPopulated() bool {
	if len(a.EmojiIDs) != len(a.Emojis) {
		// this is the quickest indicator.
		return false
	}

	// Emojis must be in same order.
	for i, id := range a.EmojiIDs {
		if a.Emojis[i] == nil {
			log.Warnf(nil, "nil emoji in slice for account %s", a.URI)
			continue
		}
		if a.Emojis[i].ID != id {
			return false
		}
	}

	return true
}

// AlsoKnownAsPopulated returns whether alsoKnownAs accounts
// are populated according to current AlsoKnownAsURIs.
func (a *Account) AlsoKnownAsPopulated() bool {
	if len(a.AlsoKnownAsURIs) != len(a.AlsoKnownAs) {
		// this is the quickest indicator.
		return false
	}

	// Accounts must be in same order.
	for i, uri := range a.AlsoKnownAsURIs {
		if a.AlsoKnownAs[i] == nil {
			log.Warnf(nil, "nil account in alsoKnownAs slice for account %s", a.URI)
			continue
		}
		if a.AlsoKnownAs[i].URI != uri {
			return false
		}
	}

	return true
}

// PubKeyExpired returns true if the account's public key
// has been marked as expired, and the expiry time has passed.
func (a *Account) PubKeyExpired() bool {
	if a == nil {
		return false
	}

	return !a.PublicKeyExpiresAt.IsZero() &&
		a.PublicKeyExpiresAt.Before(time.Now())
}

// IsAliasedTo returns true if account
// is aliased to the given account URI.
func (a *Account) IsAliasedTo(uri string) bool {
	return slices.Contains(a.AlsoKnownAsURIs, uri)
}

// IsSuspended returns true if account
// has been suspended from this instance.
func (a *Account) IsSuspended() bool {
	return !a.SuspendedAt.IsZero()
}

// IsMoving returns true if
// account is Moving or has Moved.
func (a *Account) IsMoving() bool {
	return a.MovedToURI != "" || a.MoveID != ""
}

// AccountToEmoji is an intermediate struct to facilitate the many2many relationship between an account and one or more emojis.
type AccountToEmoji struct {
	AccountID string   `bun:"type:CHAR(26),unique:accountemoji,nullzero,notnull"`
	Account   *Account `bun:"rel:belongs-to"`
	EmojiID   string   `bun:"type:CHAR(26),unique:accountemoji,nullzero,notnull"`
	Emoji     *Emoji   `bun:"rel:belongs-to"`
}

// Field represents a key value field on an account, for things like pronouns, website, etc.
// VerifiedAt is optional, to be used only if Value is a URL to a webpage that contains the
// username of the user.
type Field struct {
	Name       string    // Name of this field.
	Value      string    // Value of this field.
	VerifiedAt time.Time `bun:",nullzero"` // This field was verified at (optional).
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
	Requested           bool   // Do you have a pending follow request targeting this user?
	RequestedBy         bool   // Does the user have a pending follow request targeting you?
	DomainBlocking      bool   // Are you blocking this user's domain?
	Endorsed            bool   // Are you featuring this user on your profile?
	Note                string // Your note on this account.
}

// Theme represents a user-selected
// CSS theme for an account.
type Theme struct {
	// User-facing title of this theme.
	Title string

	// User-facing description of this theme.
	Description string

	// FileName of this theme in the themes
	// directory (eg., `light-blurple.css`).
	FileName string
}
