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

package cache_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type AccountCacheTestSuite struct {
	suite.Suite
	data  map[string]*gtsmodel.Account
	cache *cache.AccountCache
}

func (suite *AccountCacheTestSuite) SetupSuite() {
	suite.data = testrig.NewTestAccounts()
}

func (suite *AccountCacheTestSuite) SetupTest() {
	suite.cache = cache.NewAccountCache()
}

func (suite *AccountCacheTestSuite) TearDownTest() {
	suite.data = nil
	suite.cache = nil
}

func (suite *AccountCacheTestSuite) TestAccountCache() {
	for _, account := range suite.data {
		// Place in the cache
		suite.cache.Put(account)
	}

	for _, account := range suite.data {
		var ok bool
		var check *gtsmodel.Account

		// Check we can retrieve
		check, ok = suite.cache.GetByID(account.ID)
		if !ok && !accountIs(account, check) {
			suite.Fail(fmt.Sprintf("Failed to fetch expected account with ID: %s", account.ID))
		}
		check, ok = suite.cache.GetByURI(account.URI)
		if account.URI != "" && !ok && !accountIs(account, check) {
			suite.Fail(fmt.Sprintf("Failed to fetch expected account with URI: %s", account.URI))
		}
		check, ok = suite.cache.GetByURL(account.URL)
		if account.URL != "" && !ok && !accountIs(account, check) {
			suite.Fail(fmt.Sprintf("Failed to fetch expected account with URL: %s", account.URL))
		}
		check, ok = suite.cache.GetByPubkeyID(account.PublicKeyURI)
		if account.PublicKeyURI != "" && !ok && !accountIs(account, check) {
			suite.Fail(fmt.Sprintf("Failed to fetch expected account with public key URI: %s", account.PublicKeyURI))
		}
		check, ok = suite.cache.GetByUsernameDomain(account.Username, account.Domain)
		if !ok && !accountIs(account, check) {
			suite.Fail(fmt.Sprintf("Failed to fetch expected account with username/domain: %s/%s", account.Username, account.Domain))
		}
	}
}

func TestAccountCache(t *testing.T) {
	suite.Run(t, &AccountCacheTestSuite{})
}

func accountIs(account1, account2 *gtsmodel.Account) bool {
	if account1 == nil || account2 == nil {
		return account1 == account2
	}
	return account1.ID == account2.ID &&
		account1.URI == account2.URI &&
		account1.URL == account2.URL &&
		account1.PublicKeyURI == account2.PublicKeyURI
}
