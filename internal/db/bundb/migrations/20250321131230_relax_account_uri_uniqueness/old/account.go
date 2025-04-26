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

	"code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250321131230_relax_account_uri_uniqueness/common"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel           `bun:"table:accounts"`
	ID                      string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt               time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt               time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	FetchedAt               time.Time       `bun:"type:timestamptz,nullzero"`
	Username                string          `bun:",nullzero,notnull,unique:usernamedomain"`
	Domain                  string          `bun:",nullzero,unique:usernamedomain"`
	AvatarMediaAttachmentID string          `bun:"type:CHAR(26),nullzero"`
	AvatarRemoteURL         string          `bun:",nullzero"`
	HeaderMediaAttachmentID string          `bun:"type:CHAR(26),nullzero"`
	HeaderRemoteURL         string          `bun:",nullzero"`
	DisplayName             string          `bun:""`
	EmojiIDs                []string        `bun:"emojis,array"`
	Fields                  []*common.Field `bun:""`
	FieldsRaw               []*common.Field `bun:""`
	Note                    string          `bun:""`
	NoteRaw                 string          `bun:""`
	Memorial                *bool           `bun:",default:false"`
	AlsoKnownAsURIs         []string        `bun:"also_known_as_uris,array"`
	MovedToURI              string          `bun:",nullzero"`
	MoveID                  string          `bun:"type:CHAR(26),nullzero"`
	Bot                     *bool           `bun:",default:false"`
	Locked                  *bool           `bun:",default:true"`
	Discoverable            *bool           `bun:",default:false"`
	URI                     string          `bun:",nullzero,notnull,unique"`
	URL                     string          `bun:",nullzero,unique"`
	InboxURI                string          `bun:",nullzero,unique"`
	SharedInboxURI          *string         `bun:""`
	OutboxURI               string          `bun:",nullzero,unique"`
	FollowingURI            string          `bun:",nullzero,unique"`
	FollowersURI            string          `bun:",nullzero,unique"`
	FeaturedCollectionURI   string          `bun:",nullzero,unique"`
	ActorType               string          `bun:",nullzero,notnull"`
	PrivateKey              *rsa.PrivateKey `bun:""`
	PublicKey               *rsa.PublicKey  `bun:",notnull"`
	PublicKeyURI            string          `bun:",nullzero,notnull,unique"`
	PublicKeyExpiresAt      time.Time       `bun:"type:timestamptz,nullzero"`
	SensitizedAt            time.Time       `bun:"type:timestamptz,nullzero"`
	SilencedAt              time.Time       `bun:"type:timestamptz,nullzero"`
	SuspendedAt             time.Time       `bun:"type:timestamptz,nullzero"`
	SuspensionOrigin        string          `bun:"type:CHAR(26),nullzero"`
}
