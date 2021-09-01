package cache_test

import (
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
			suite.Fail("Failed to fetch expected account with ID: %s", account.ID)
		}
		check, ok = suite.cache.GetByURI(account.URI)
		if account.URI != "" && !ok && !accountIs(account, check) {
			suite.Fail("Failed to fetch expected account with URI: %s", account.URI)
		}
		check, ok = suite.cache.GetByURL(account.URL)
		if account.URL != "" && !ok && !accountIs(account, check) {
			suite.Fail("Failed to fetch expected account with URL: %s", account.URL)
		}
	}
}

func TestAccountCache(t *testing.T) {
	suite.Run(t, &AccountCacheTestSuite{})
}

func accountIs(account1, account2 *gtsmodel.Account) bool {
	return account1.ID == account2.ID && account1.URI == account2.URI && account1.URL == account2.URL
}
