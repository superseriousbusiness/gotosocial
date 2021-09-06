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
	ID                    string          `json:"id"`
	CreatedAt             *time.Time      `json:"created_at"`
	UpdatedAt             *time.Time      `json:"updated_at"`
	Username              string          `json:"username"`
	Domain                string          `json:"domain,omitempty"`
	Locked                bool            `json:"locked"`
	Language              string          `json:"language,omitempty"`
	URI                   string          `json:"uri"`
	URL                   string          `json:"url"`
	InboxURI              string          `json:"inbox_uri"`
	OutboxURI             string          `json:"outbox_uri"`
	FollowingURI          string          `json:"following_uri"`
	FollowersURI          string          `json:"followers_uri"`
	FeaturedCollectionURI string          `json:"featured_collection_uri"`
	ActorType             string          `json:"actor_type"`
	PrivateKey            *rsa.PrivateKey `json:"private_key,omitempty"`
	PublicKey             *rsa.PublicKey  `json:"public_key"`
	PublicKeyURI          string          `json:"public_key_uri"`
	SuspendedAt           *time.Time      `json:"suspended_at,omitempty"`
	SuspensionOrigin      string          `json:"suspension_origin,omitempty"`
}
