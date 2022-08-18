/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package cache

import (
	"time"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// AccountCache is a cache wrapper to provide URL and URI lookups for gtsmodel.Account
type AccountCache struct {
	cache cache.LookupCache[string, string, *gtsmodel.Account]
}

// NewAccountCache returns a new instantiated AccountCache object
func NewAccountCache() *AccountCache {
	c := &AccountCache{}
	c.cache = cache.NewLookup(cache.LookupCfg[string, string, *gtsmodel.Account]{
		RegisterLookups: func(lm *cache.LookupMap[string, string]) {
			lm.RegisterLookup("uri")
			lm.RegisterLookup("url")
			lm.RegisterLookup("usernamedomain")
		},

		AddLookups: func(lm *cache.LookupMap[string, string], acc *gtsmodel.Account) {
			if uri := acc.URI; uri != "" {
				lm.Set("uri", uri, acc.ID)
			}
			if url := acc.URL; url != "" {
				lm.Set("url", url, acc.ID)
			}
			lm.Set("usernamedomain", usernameDomainKey(acc.Username, acc.Domain), acc.ID)
		},

		DeleteLookups: func(lm *cache.LookupMap[string, string], acc *gtsmodel.Account) {
			if uri := acc.URI; uri != "" {
				lm.Delete("uri", uri)
			}
			if url := acc.URL; url != "" {
				lm.Delete("url", url)
			}
			lm.Delete("usernamedomain", usernameDomainKey(acc.Username, acc.Domain))
		},
	})
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

// GetByID attempts to fetch a account from the cache by its ID, you will receive a copy for thread-safety
func (c *AccountCache) GetByID(id string) (*gtsmodel.Account, bool) {
	return c.cache.Get(id)
}

// GetByURL attempts to fetch a account from the cache by its URL, you will receive a copy for thread-safety
func (c *AccountCache) GetByURL(url string) (*gtsmodel.Account, bool) {
	return c.cache.GetBy("url", url)
}

// GetByURI attempts to fetch a account from the cache by its URI, you will receive a copy for thread-safety
func (c *AccountCache) GetByURI(uri string) (*gtsmodel.Account, bool) {
	return c.cache.GetBy("uri", uri)
}

func (c *AccountCache) GetByUsernameDomain(username string, domain string) (*gtsmodel.Account, bool) {
	return c.cache.GetBy("usernamedomain", usernameDomainKey(username, domain))
}

// Put places a account in the cache, ensuring that the object place is a copy for thread-safety
func (c *AccountCache) Put(account *gtsmodel.Account) {
	if account == nil || account.ID == "" {
		panic("invalid account")
	}
	c.cache.Set(account.ID, copyAccount(account))
}

// copyAccount performs a surface-level copy of account, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyAccount(account *gtsmodel.Account) *gtsmodel.Account {
	return &gtsmodel.Account{
		ID:                      account.ID,
		Username:                account.Username,
		Domain:                  account.Domain,
		AvatarMediaAttachmentID: account.AvatarMediaAttachmentID,
		AvatarMediaAttachment:   nil,
		AvatarRemoteURL:         account.AvatarRemoteURL,
		HeaderMediaAttachmentID: account.HeaderMediaAttachmentID,
		HeaderMediaAttachment:   nil,
		HeaderRemoteURL:         account.HeaderRemoteURL,
		DisplayName:             account.DisplayName,
		Fields:                  account.Fields,
		Note:                    account.Note,
		NoteRaw:                 account.NoteRaw,
		Memorial:                copyBoolPtr(account.Memorial),
		MovedToAccountID:        account.MovedToAccountID,
		Bot:                     copyBoolPtr(account.Bot),
		CreatedAt:               account.CreatedAt,
		UpdatedAt:               account.UpdatedAt,
		Reason:                  account.Reason,
		Locked:                  copyBoolPtr(account.Locked),
		Discoverable:            copyBoolPtr(account.Discoverable),
		Privacy:                 account.Privacy,
		Sensitive:               copyBoolPtr(account.Sensitive),
		Language:                account.Language,
		StatusFormat:            account.StatusFormat,
		URI:                     account.URI,
		URL:                     account.URL,
		LastWebfingeredAt:       account.LastWebfingeredAt,
		InboxURI:                account.InboxURI,
		OutboxURI:               account.OutboxURI,
		FollowingURI:            account.FollowingURI,
		FollowersURI:            account.FollowersURI,
		FeaturedCollectionURI:   account.FeaturedCollectionURI,
		ActorType:               account.ActorType,
		AlsoKnownAs:             account.AlsoKnownAs,
		PrivateKey:              account.PrivateKey,
		PublicKey:               account.PublicKey,
		PublicKeyURI:            account.PublicKeyURI,
		SensitizedAt:            account.SensitizedAt,
		SilencedAt:              account.SilencedAt,
		SuspendedAt:             account.SuspendedAt,
		HideCollections:         copyBoolPtr(account.HideCollections),
		SuspensionOrigin:        account.SuspensionOrigin,
	}
}

func usernameDomainKey(username string, domain string) string {
	u := "@" + username
	if domain != "" {
		return u + "@" + domain
	}
	return u
}
