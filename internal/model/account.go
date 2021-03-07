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

package model

import (
	"net/url"
	"time"
)

// Account represents a user account
type Account struct {
	Avatar
	Header
	ID                    int `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	Username              string
	Domain                string
	Secret                string
	PrivateKey            string
	PublicKey             string
	RemoteURL             *url.URL  `pg:"type:text"`
	SalmonURL             *url.URL  `pg:"type:text"`
	HubURL                *url.URL  `pg:"type:text"`
	CreatedAt             time.Time `pg:"type:timestamp,notnull"`
	UpdatedAt             time.Time `pg:"type:timestamp,notnull"`
	Note                  string
	DisplayName           string
	URI                   *url.URL  `pg:"type:text"`
	URL                   *url.URL  `pg:"type:text"`
	SubscriptionExpiresAt time.Time `pg:"type:timestamp"`
	Locked                bool
	LastWebfingeredAt     time.Time `pg:"type:timestamp"`
	InboxURL              *url.URL  `pg:"type:text"`
	OutboxURL             *url.URL  `pg:"type:text"`
	SharedInboxURL        *url.URL  `pg:"type:text"`
	FollowersURL          *url.URL  `pg:"type:text"`
	Protocol              int
	Memorial              bool
	MovedToAccountID      int
	FeaturedCollectionURL *url.URL `pg:"type:text"`
	Fields                map[string]string
	ActorType             string
	Discoverable          bool
	AlsoKnownAs           string
	SilencedAt            time.Time `pg:"type:timestamp"`
	SuspendedAt           time.Time `pg:"type:timestamp"`
	TrustLevel            int
	HideCollections       bool
	DevicesURL            *url.URL  `pg:"type:text"`
	SensitizedAt          time.Time `pg:"type:timestamp"`
	SuspensionOrigin      int
}

type Avatar struct {
	AvatarFileName             string
	AvatarContentType          string
	AvatarFileSize             int
	AvatarUpdatedAt            *time.Time `pg:"type:timestamp"`
	AvatarRemoteURL            *url.URL   `pg:"type:text"`
	AvatarStorageSchemaVersion int
}

type Header struct {
	HeaderFileName             string
	HeaderContentType          string
	HeaderFileSize             int
	HeaderUpdatedAt            *time.Time `pg:"type:timestamp"`
	HeaderRemoteURL            *url.URL   `pg:"type:text"`
	HeaderStorageSchemaVersion int
}

func StubAccount() *Account {

	remoteURL, _ := url.Parse("https://example.org/@someuser")

	return &Account{
		Username:  "some_user",
		Domain:    "example.org",
		RemoteURL: remoteURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
