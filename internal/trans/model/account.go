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
	Memorial              *bool           `json:"memorial"`
	Bot                   *bool           `json:"bot"`
	Reason                string          `json:"reason,omitempty" bun:",nullzero"`
	Locked                *bool           `json:"locked"`
	Discoverable          *bool           `json:"discoverable"`
	Privacy               string          `json:"privacy,omitempty" bun:",nullzero"`
	Sensitive             *bool           `json:"sensitive"`
	Language              string          `json:"language,omitempty" bun:",nullzero"`
	StatusFormat          string          `json:"statusFormat,omitempty" bun:",nullzero"`
	URI                   string          `json:"uri" bun:",nullzero"`
	URL                   string          `json:"url" bun:",nullzero"`
	InboxURI              string          `json:"inboxURI" bun:",nullzero"`
	OutboxURI             string          `json:"outboxURI" bun:",nullzero"`
	FollowingURI          string          `json:"followingUri" bun:",nullzero"`
	FollowersURI          string          `json:"followersUri" bun:",nullzero"`
	FeaturedCollectionURI string          `json:"featuredCollectionUri" bun:",nullzero"`
	ActorType             string          `json:"actorType" bun:",nullzero"`
	PrivateKey            *rsa.PrivateKey `json:"-" mapstructure:"-"`
	PrivateKeyString      string          `json:"privateKey,omitempty" mapstructure:"privateKey" bun:"-"`
	PublicKey             *rsa.PublicKey  `json:"-" mapstructure:"-"`
	PublicKeyString       string          `json:"publicKey,omitempty" mapstructure:"publicKey" bun:"-"`
	PublicKeyURI          string          `json:"publicKeyUri" bun:",nullzero"`
	SensitizedAt          *time.Time      `json:"sensitizedAt,omitempty" bun:",nullzero"`
	SilencedAt            *time.Time      `json:"silencedAt,omitempty" bun:",nullzero"`
	SuspendedAt           *time.Time      `json:"suspendedAt,omitempty" bun:",nullzero"`
	HideCollections       *bool           `json:"hideCollections"`
	SuspensionOrigin      string          `json:"suspensionOrigin,omitempty" bun:",nullzero"`
}
