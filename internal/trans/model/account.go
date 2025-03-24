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

package trans

import (
	"crypto/rsa"
	"time"
)

// Account represents the minimum viable representation of an account for export/import.
type Account struct {
	Type                  Type            `json:"type" bun:"-"`
	ID                    string          `json:"id" bun:",nullzero"`
	CreatedAt             *time.Time      `json:"createdAt" bun:",nullzero"`
	Username              string          `json:"username" bun:",nullzero"`
	Domain                string          `json:"domain,omitempty" bun:",nullzero"`
	HeaderRemoteURL       string          `json:"headerRemoteURL,omitempty" bun:",nullzero"`
	AvatarRemoteURL       string          `json:"avatarRemoteURL,omitempty" bun:",nullzero"`
	DisplayName           string          `json:"displayName,omitempty" bun:",nullzero"`
	Note                  string          `json:"note,omitempty" bun:",nullzero"`
	NoteRaw               string          `json:"noteRaw,omitempty" bun:",nullzero"`
	Locked                *bool           `json:"locked"`
	Discoverable          *bool           `json:"discoverable"`
	URI                   string          `json:"uri" bun:",nullzero"`
	URL                   string          `json:"url" bun:",nullzero"`
	InboxURI              string          `json:"inboxURI" bun:",nullzero"`
	OutboxURI             string          `json:"outboxURI" bun:",nullzero"`
	FollowingURI          string          `json:"followingUri" bun:",nullzero"`
	FollowersURI          string          `json:"followersUri" bun:",nullzero"`
	FeaturedCollectionURI string          `json:"featuredCollectionUri" bun:",nullzero"`
	ActorType             int16           `json:"actorType" bun:",nullzero"`
	PrivateKey            *rsa.PrivateKey `json:"-" mapstructure:"-"`
	PrivateKeyString      string          `json:"privateKey,omitempty" mapstructure:"privateKey" bun:"-"`
	PublicKey             *rsa.PublicKey  `json:"-" mapstructure:"-"`
	PublicKeyString       string          `json:"publicKey,omitempty" mapstructure:"publicKey" bun:"-"`
	PublicKeyURI          string          `json:"publicKeyUri" bun:",nullzero"`
	SensitizedAt          *time.Time      `json:"sensitizedAt,omitempty" bun:",nullzero"`
	SilencedAt            *time.Time      `json:"silencedAt,omitempty" bun:",nullzero"`
	SuspendedAt           *time.Time      `json:"suspendedAt,omitempty" bun:",nullzero"`
	SuspensionOrigin      string          `json:"suspensionOrigin,omitempty" bun:",nullzero"`
}

type AccountSettings struct {
	AccountID string
}
