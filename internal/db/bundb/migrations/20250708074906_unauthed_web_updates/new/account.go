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

package gtsmodel

import (
	"crypto/rsa"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250708074906_unauthed_web_updates/common"
)

type Account struct {
	ID                      string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt               time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt               time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	FetchedAt               time.Time       `bun:"type:timestamptz,nullzero"`
	Username                string          `bun:",nullzero,notnull,unique:accounts_username_domain_uniq"`
	Domain                  string          `bun:",nullzero,unique:accounts_username_domain_uniq"`
	AvatarMediaAttachmentID string          `bun:"type:CHAR(26),nullzero"`
	AvatarRemoteURL         string          `bun:",nullzero"`
	HeaderMediaAttachmentID string          `bun:"type:CHAR(26),nullzero"`
	HeaderRemoteURL         string          `bun:",nullzero"`
	DisplayName             string          `bun:",nullzero"`
	EmojiIDs                []string        `bun:"emojis,array"`
	Fields                  []*Field        `bun:",nullzero"`
	FieldsRaw               []*Field        `bun:",nullzero"`
	Note                    string          `bun:",nullzero"`
	NoteRaw                 string          `bun:",nullzero"`
	AlsoKnownAsURIs         []string        `bun:"also_known_as_uris,array"`
	AlsoKnownAs             []*Account      `bun:"-"`
	MovedToURI              string          `bun:",nullzero"`
	MovedTo                 *Account        `bun:"-"`
	MoveID                  string          `bun:"type:CHAR(26),nullzero"`
	Locked                  *bool           `bun:",nullzero,notnull,default:true"`
	Discoverable            *bool           `bun:",nullzero,notnull,default:false"`
	URI                     string          `bun:",nullzero,notnull,unique"`
	URL                     string          `bun:",nullzero"`
	InboxURI                string          `bun:",nullzero"`
	SharedInboxURI          *string         `bun:""`
	OutboxURI               string          `bun:",nullzero"`
	FollowingURI            string          `bun:",nullzero"`
	FollowersURI            string          `bun:",nullzero"`
	FeaturedCollectionURI   string          `bun:",nullzero"`
	ActorType               int16           `bun:",nullzero,notnull"`
	PrivateKey              *rsa.PrivateKey `bun:""`
	PublicKey               *rsa.PublicKey  `bun:",notnull"`
	PublicKeyURI            string          `bun:",nullzero,notnull,unique"`
	PublicKeyExpiresAt      time.Time       `bun:"type:timestamptz,nullzero"`
	MemorializedAt          time.Time       `bun:"type:timestamptz,nullzero"`
	SensitizedAt            time.Time       `bun:"type:timestamptz,nullzero"`
	SilencedAt              time.Time       `bun:"type:timestamptz,nullzero"`
	SuspendedAt             time.Time       `bun:"type:timestamptz,nullzero"`
	SuspensionOrigin        string          `bun:"type:CHAR(26),nullzero"`

	// Added in this migration:
	HidesToPublicFromUnauthedWeb *bool `bun:",nullzero,notnull,default:false"`
	HidesCcPublicFromUnauthedWeb *bool `bun:",nullzero,notnull,default:false"`
}

type Field struct {
	Name       string
	Value      string
	VerifiedAt time.Time `bun:",nullzero"`
}

type AccountSettings struct {
	AccountID                      string            `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt                      time.Time         `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt                      time.Time         `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	Privacy                        common.Visibility `bun:",nullzero,default:3"`
	Sensitive                      *bool             `bun:",nullzero,notnull,default:false"`
	Language                       string            `bun:",nullzero,notnull,default:'en'"`
	StatusContentType              string            `bun:",nullzero"`
	Theme                          string            `bun:",nullzero"`
	CustomCSS                      string            `bun:",nullzero"`
	EnableRSS                      *bool             `bun:",nullzero,notnull,default:false"`
	HideCollections                *bool             `bun:",nullzero,notnull,default:false"`
	WebLayout                      int16             `bun:",nullzero,notnull,default:1"`
	InteractionPolicyDirect        *struct{}         `bun:""`
	InteractionPolicyMutualsOnly   *struct{}         `bun:""`
	InteractionPolicyFollowersOnly *struct{}         `bun:""`
	InteractionPolicyUnlocked      *struct{}         `bun:""`
	InteractionPolicyPublic        *struct{}         `bun:""`

	// Removed in this migration:
	// WebVisibility common.Visibility `bun:",nullzero,notnull,default:3"`
}
