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

// UserCache is a cache wrapper to provide lookups for gtsmodel.User
type UserCache struct {
	cache cache.LookupCache[string, string, *gtsmodel.User]
}

// NewUserCache returns a new instantiated UserCache object
func NewUserCache() *UserCache {
	c := &UserCache{}
	c.cache = cache.NewLookup(cache.LookupCfg[string, string, *gtsmodel.User]{
		RegisterLookups: func(lm *cache.LookupMap[string, string]) {
			lm.RegisterLookup("accountid")
			lm.RegisterLookup("email")
			lm.RegisterLookup("unconfirmedemail")
			lm.RegisterLookup("confirmationtoken")
		},

		AddLookups: func(lm *cache.LookupMap[string, string], user *gtsmodel.User) {
			lm.Set("accountid", user.AccountID, user.ID)
			if email := user.Email; email != "" {
				lm.Set("email", email, user.ID)
			}
			if unconfirmedEmail := user.UnconfirmedEmail; unconfirmedEmail != "" {
				lm.Set("unconfirmedemail", unconfirmedEmail, user.ID)
			}
			if confirmationToken := user.ConfirmationToken; confirmationToken != "" {
				lm.Set("confirmationtoken", confirmationToken, user.ID)
			}
		},

		DeleteLookups: func(lm *cache.LookupMap[string, string], user *gtsmodel.User) {
			lm.Delete("accountid", user.AccountID)
			if email := user.Email; email != "" {
				lm.Delete("email", email)
			}
			if unconfirmedEmail := user.UnconfirmedEmail; unconfirmedEmail != "" {
				lm.Delete("unconfirmedemail", unconfirmedEmail)
			}
			if confirmationToken := user.ConfirmationToken; confirmationToken != "" {
				lm.Delete("confirmationtoken", confirmationToken)
			}
		},
	})
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

// GetByID attempts to fetch a user from the cache by its ID, you will receive a copy for thread-safety
func (c *UserCache) GetByID(id string) (*gtsmodel.User, bool) {
	return c.cache.Get(id)
}

// GetByAccountID attempts to fetch a user from the cache by its account ID, you will receive a copy for thread-safety
func (c *UserCache) GetByAccountID(accountID string) (*gtsmodel.User, bool) {
	return c.cache.GetBy("accountid", accountID)
}

// GetByEmail attempts to fetch a user from the cache by its email address, you will receive a copy for thread-safety
func (c *UserCache) GetByEmail(email string) (*gtsmodel.User, bool) {
	return c.cache.GetBy("email", email)
}

// GetByUnconfirmedEmail attempts to fetch a user from the cache by its confirmation token, you will receive a copy for thread-safety
func (c *UserCache) GetByConfirmationToken(token string) (*gtsmodel.User, bool) {
	return c.cache.GetBy("confirmationtoken", token)
}

// Put places a user in the cache, ensuring that the object place is a copy for thread-safety
func (c *UserCache) Put(user *gtsmodel.User) {
	if user == nil || user.ID == "" {
		panic("invalid user")
	}
	c.cache.Set(user.ID, copyUser(user))
}

// Invalidate invalidates one user from the cache using the ID of the user as key.
func (c *UserCache) Invalidate(userID string) {
	c.cache.Invalidate(userID)
}

func copyUser(user *gtsmodel.User) *gtsmodel.User {
	return &gtsmodel.User{
		ID:                     user.ID,
		CreatedAt:              user.CreatedAt,
		UpdatedAt:              user.UpdatedAt,
		Email:                  user.Email,
		AccountID:              user.AccountID,
		Account:                nil,
		EncryptedPassword:      user.EncryptedPassword,
		SignUpIP:               user.SignUpIP,
		CurrentSignInAt:        user.CurrentSignInAt,
		CurrentSignInIP:        user.CurrentSignInIP,
		LastSignInAt:           user.LastSignInAt,
		LastSignInIP:           user.LastSignInIP,
		SignInCount:            user.SignInCount,
		InviteID:               user.InviteID,
		ChosenLanguages:        user.ChosenLanguages,
		FilteredLanguages:      user.FilteredLanguages,
		Locale:                 user.Locale,
		CreatedByApplicationID: user.CreatedByApplicationID,
		CreatedByApplication:   nil,
		LastEmailedAt:          user.LastEmailedAt,
		ConfirmationToken:      user.ConfirmationToken,
		ConfirmationSentAt:     user.ConfirmationSentAt,
		ConfirmedAt:            user.ConfirmedAt,
		UnconfirmedEmail:       user.UnconfirmedEmail,
		Moderator:              copyBoolPtr(user.Moderator),
		Admin:                  copyBoolPtr(user.Admin),
		Disabled:               copyBoolPtr(user.Disabled),
		Approved:               copyBoolPtr(user.Approved),
		ResetPasswordToken:     user.ResetPasswordToken,
		ResetPasswordSentAt:    user.ResetPasswordSentAt,
	}
}
