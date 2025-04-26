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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// Account represents either a local or a remote ActivityPub actor.
// https://www.w3.org/TR/activitypub/#actor-objects
type Account struct {
	// Database ID of the account.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Datetime when the account was created.
	// Corresponds to ActivityStreams `published` prop.
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Datetime when was the account was last updated,
	// ie., when the actor last sent out an Update
	// activity, or if never, when it was `published`.
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Datetime when the account was last fetched /
	// dereferenced by this GoToSocial instance.
	FetchedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Username of the account.
	//
	// Corresponds to AS `preferredUsername` prop, which gives
	// no uniqueness guarantee. However, we do enforce uniqueness
	// for it as, in practice, it always is and we rely on this.
	Username string `bun:",nullzero,notnull,unique:accounts_username_domain_uniq"`

	// Domain of the account, discovered via webfinger.
	//
	// Null if this is a local account, otherwise
	// something like `example.org`.
	Domain string `bun:",nullzero,unique:accounts_username_domain_uniq"`

	// Database ID of the account's avatar MediaAttachment, if set.
	AvatarMediaAttachmentID string `bun:"type:CHAR(26),nullzero"`

	// MediaAttachment corresponding to AvatarMediaAttachmentID.
	AvatarMediaAttachment *MediaAttachment `bun:"-"`

	// URL of the avatar media.
	//
	// Null for local accounts.
	AvatarRemoteURL string `bun:",nullzero"`

	// Database ID of the account's header MediaAttachment, if set.
	HeaderMediaAttachmentID string `bun:"type:CHAR(26),nullzero"`

	// MediaAttachment corresponding to HeaderMediaAttachmentID.
	HeaderMediaAttachment *MediaAttachment `bun:"-"`

	// URL of the header media.
	//
	// Null for local accounts.
	HeaderRemoteURL string `bun:",nullzero"`

	// Display name for this account, if set.
	//
	// Corresponds to the ActivityStreams `name` property.
	//
	// If null, fall back to username for display purposes.
	DisplayName string `bun:",nullzero"`

	// Database IDs of any emojis used in
	// this account's bio, display name, etc
	EmojiIDs []string `bun:"emojis,array"`

	// Emojis corresponding to EmojiIDs.
	Emojis []*Emoji `bun:"-"`

	// A slice of of key/value fields that
	// this account has added to their profile.
	//
	// Corresponds to schema.org PropertyValue types in `attachments`.
	Fields []*Field `bun:",nullzero"`

	// The raw (unparsed) content of fields that this
	// account has added to their profile, before
	// conversion to HTML.
	//
	// Only set for local accounts.
	FieldsRaw []*Field `bun:",nullzero"`

	// A note that this account has on their profile
	// (ie., the account's bio/description of themselves).
	//
	// Corresponds to the ActivityStreams `summary` property.
	Note string `bun:",nullzero"`

	// The raw (unparsed) version of Note, before conversion to HTML.
	//
	// Only set for local accounts.
	NoteRaw string `bun:",nullzero"`

	// ActivityPub URI/IDs by which this account is also known.
	//
	// Corresponds to the ActivityStreams `alsoKnownAs` property.
	AlsoKnownAsURIs []string `bun:"also_known_as_uris,array"`

	// Accounts matching AlsoKnownAsURIs.
	AlsoKnownAs []*Account `bun:"-"`

	// URI/ID to which the account has (or claims to have) moved.
	//
	// Corresponds to the ActivityStreams `movedTo` property.
	//
	// Even if this field is set the move may not yet have been
	// processed. Check `move` for this.
	MovedToURI string `bun:",nullzero"`

	// Account matching MovedToURI.
	MovedTo *Account `bun:"-"`

	// ID of a Move in the database for this account.
	// Only set if we received or created a Move activity
	// for which this account URI was the origin.
	MoveID string `bun:"type:CHAR(26),nullzero"`

	// Move corresponding to MoveID, if set.
	Move *Move `bun:"-"`

	// True if account requires manual approval of Follows.
	//
	// Corresponds to AS `manuallyApprovesFollowers` prop.
	Locked *bool `bun:",nullzero,notnull,default:true"`

	// True if account has opted in to being shown in
	// directories and exposed to search engines.
	//
	// Corresponds to the toot `discoverable` property.
	Discoverable *bool `bun:",nullzero,notnull,default:false"`

	// ActivityPub URI/ID for this account.
	//
	// Must be set, must be unique.
	URI string `bun:",nullzero,notnull,unique"`

	// URL at which a web representation of this
	// account should be available, if set.
	//
	// Corresponds to ActivityStreams `url` prop.
	URL string `bun:",nullzero"`

	// URI of the actor's inbox.
	//
	// Corresponds to ActivityPub `inbox` property.
	//
	// According to AP this MUST be set, but some
	// implementations don't set it for service actors.
	InboxURI string `bun:",nullzero"`

	// URI/ID of this account's sharedInbox, if set.
	//
	// Corresponds to ActivityPub `endpoints.sharedInbox`.
	//
	// Gotcha warning: this is a string pointer because
	// it has three possible states:
	//
	//   1. null: We don't know (yet) if actor has a shared inbox.
	//   2. empty: We know it doesn't have a shared inbox.
	//   3. not empty: We know it does have a shared inbox.
	SharedInboxURI *string `bun:""`

	// URI/ID of the actor's outbox.
	//
	// Corresponds to ActivityPub `outbox` property.
	//
	// According to AP this MUST be set, but some
	// implementations don't set it for service actors.
	OutboxURI string `bun:",nullzero"`

	// URI/ID of the actor's following collection.
	//
	// Corresponds to ActivityPub `following` property.
	//
	// According to AP this SHOULD be set.
	FollowingURI string `bun:",nullzero"`

	// URI/ID of the actor's followers collection.
	//
	// Corresponds to ActivityPub `followers` property.
	//
	// According to AP this SHOULD be set.
	FollowersURI string `bun:",nullzero"`

	// URI/ID of the actor's featured collection.
	//
	// Corresponds to the Toot `featured` property.
	FeaturedCollectionURI string `bun:",nullzero"`

	// ActivityStreams type of the actor.
	//
	// Application, Group, Organization, Person, or Service.
	ActorType AccountActorType `bun:",nullzero,notnull"`

	// Private key for signing http requests.
	//
	// Only defined for local accounts
	PrivateKey *rsa.PrivateKey `bun:""`

	// Public key for authorizing signed http requests.
	//
	// Defined for both local and remote accounts
	PublicKey *rsa.PublicKey `bun:",notnull"`

	// Dereferenceable location of this actor's public key.
	//
	// Corresponds to https://w3id.org/security/v1 `publicKey.id`.
	PublicKeyURI string `bun:",nullzero,notnull,unique"`

	// Datetime at which public key will expire/has expired,
	// and should be fetched again as appropriate.
	//
	// Only ever set for remote accounts.
	PublicKeyExpiresAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was marked as a "memorial",
	// ie., user owning the account has passed away.
	MemorializedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was set to
	// have all its media shown as sensitive.
	SensitizedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was silenced.
	SilencedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Datetime at which account was suspended.
	SuspendedAt time.Time `bun:"type:timestamptz,nullzero"`

	// ID of the database entry that caused this account to
	// be suspended. Can be an account ID or a domain block ID.
	SuspensionOrigin string `bun:"type:CHAR(26),nullzero"`

	// gtsmodel.AccountSettings for this account.
	//
	// Local, non-instance-actor accounts only.
	Settings *AccountSettings `bun:"-"`

	// gtsmodel.AccountStats for this account.
	//
	// Local accounts only.
	Stats *AccountStats `bun:"-"`
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

// AccountActorType is the ActivityStreams type of an actor.
type AccountActorType enumType

const (
	AccountActorTypeUnknown      AccountActorType = 0
	AccountActorTypeApplication  AccountActorType = 1 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-application
	AccountActorTypeGroup        AccountActorType = 2 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-group
	AccountActorTypeOrganization AccountActorType = 3 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-organization
	AccountActorTypePerson       AccountActorType = 4 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-person
	AccountActorTypeService      AccountActorType = 5 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-service
)

// String returns a stringified form of AccountActorType.
func (t AccountActorType) String() string {
	switch t {
	case AccountActorTypeApplication:
		return "Application"
	case AccountActorTypeGroup:
		return "Group"
	case AccountActorTypeOrganization:
		return "Organization"
	case AccountActorTypePerson:
		return "Person"
	case AccountActorTypeService:
		return "Service"
	default:
		panic("invalid actor type")
	}
}

// ParseAccountActorType returns an
// actor type from the given value.
func ParseAccountActorType(in string) AccountActorType {
	switch strings.ToLower(in) {
	case "application":
		return AccountActorTypeApplication
	case "group":
		return AccountActorTypeGroup
	case "organization":
		return AccountActorTypeOrganization
	case "person":
		return AccountActorTypePerson
	case "service":
		return AccountActorTypeService
	default:
		return AccountActorTypeUnknown
	}
}

func (t AccountActorType) IsBot() bool {
	return t == AccountActorTypeApplication || t == AccountActorTypeService
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
