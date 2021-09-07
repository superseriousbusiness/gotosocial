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

package trans

import (
	"crypto/rsa"
	"time"
)

// Account represents the minimum viable representation of an account for export/import.
type Account struct {
	Type                  TransType       `json:"type" bun:"-"`
	ID                    string          `json:"id"`
	CreatedAt             *time.Time      `json:"createdAt"`
	Username              string          `json:"username"`
	Domain                string          `json:"domain,omitempty" bun:",nullzero"`
	Locked                bool            `json:"locked"`
	Language              string          `json:"language,omitempty"`
	URI                   string          `json:"uri"`
	URL                   string          `json:"url"`
	InboxURI              string          `json:"inboxURI"`
	OutboxURI             string          `json:"outboxURI"`
	FollowingURI          string          `json:"followingUri"`
	FollowersURI          string          `json:"followersUri"`
	FeaturedCollectionURI string          `json:"featuredCollectionUri"`
	ActorType             string          `json:"actorType"`
	PrivateKey            *rsa.PrivateKey `json:"-" mapstructure:"-"`
	PrivateKeyString      string          `json:"privateKey,omitempty" bun:"-" mapstructure:"privateKey"`
	PublicKey             *rsa.PublicKey  `json:"-" mapstructure:"-"`
	PublicKeyString       string          `json:"publicKey,omitempty" bun:"-" mapstructure:"publicKey"`
	PublicKeyURI          string          `json:"publicKeyUri"`
	SuspendedAt           *time.Time      `json:"suspendedAt,omitempty"`
	SuspensionOrigin      string          `json:"suspensionOrigin,omitempty" bun:",nullzero"`
}
