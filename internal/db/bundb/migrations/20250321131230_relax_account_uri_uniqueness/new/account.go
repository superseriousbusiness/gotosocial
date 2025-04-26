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
	"strings"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250321131230_relax_account_uri_uniqueness/common"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel           `bun:"table:new_accounts"`
	ID                      string           `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt               time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt               time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	FetchedAt               time.Time        `bun:"type:timestamptz,nullzero"`
	Username                string           `bun:",nullzero,notnull,unique:accounts_username_domain_uniq"`
	Domain                  string           `bun:",nullzero,unique:accounts_username_domain_uniq"`
	AvatarMediaAttachmentID string           `bun:"type:CHAR(26),nullzero"`
	AvatarRemoteURL         string           `bun:",nullzero"`
	HeaderMediaAttachmentID string           `bun:"type:CHAR(26),nullzero"`
	HeaderRemoteURL         string           `bun:",nullzero"`
	DisplayName             string           `bun:",nullzero"`
	EmojiIDs                []string         `bun:"emojis,array"`
	Fields                  []*common.Field  `bun:",nullzero"`
	FieldsRaw               []*common.Field  `bun:",nullzero"`
	Note                    string           `bun:",nullzero"`
	NoteRaw                 string           `bun:",nullzero"`
	MemorializedAt          time.Time        `bun:"type:timestamptz,nullzero"`
	AlsoKnownAsURIs         []string         `bun:"also_known_as_uris,array"`
	MovedToURI              string           `bun:",nullzero"`
	MoveID                  string           `bun:"type:CHAR(26),nullzero"`
	Locked                  *bool            `bun:",nullzero,notnull,default:true"`
	Discoverable            *bool            `bun:",nullzero,notnull,default:false"`
	URI                     string           `bun:",nullzero,notnull,unique"`
	URL                     string           `bun:",nullzero"`
	InboxURI                string           `bun:",nullzero"`
	SharedInboxURI          *string          `bun:""`
	OutboxURI               string           `bun:",nullzero"`
	FollowingURI            string           `bun:",nullzero"`
	FollowersURI            string           `bun:",nullzero"`
	FeaturedCollectionURI   string           `bun:",nullzero"`
	ActorType               AccountActorType `bun:",nullzero,notnull"`
	PrivateKey              *rsa.PrivateKey  `bun:""`
	PublicKey               *rsa.PublicKey   `bun:",notnull"`
	PublicKeyURI            string           `bun:",nullzero,notnull,unique"`
	PublicKeyExpiresAt      time.Time        `bun:"type:timestamptz,nullzero"`
	SensitizedAt            time.Time        `bun:"type:timestamptz,nullzero"`
	SilencedAt              time.Time        `bun:"type:timestamptz,nullzero"`
	SuspendedAt             time.Time        `bun:"type:timestamptz,nullzero"`
	SuspensionOrigin        string           `bun:"type:CHAR(26),nullzero"`
}

type AccountActorType int16

const (
	AccountActorTypeUnknown      AccountActorType = 0
	AccountActorTypeApplication  AccountActorType = 1 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-application
	AccountActorTypeGroup        AccountActorType = 2 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-group
	AccountActorTypeOrganization AccountActorType = 3 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-organization
	AccountActorTypePerson       AccountActorType = 4 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-person
	AccountActorTypeService      AccountActorType = 5 // https://www.w3.org/TR/activitystreams-vocabulary/#dfn-service
)

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
