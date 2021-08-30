package cache_test

import (
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func TestAccountCache(t *testing.T) {
	cache := cache.NewAccountCache()

	// Attempt to place an account
	account := gtsmodel.Account{
		ID:  "id",
		URI: "uri",
		URL: "url",
	}
	cache.Put(&account)

	var ok bool
	var check *gtsmodel.Account

	// Check we can retrieve
	check, ok = cache.GetByID(account.ID)
	if !ok || !accountIs(&account, check) {
		t.Fatal("Could not find expected status")
	}
	check, ok = cache.GetByURI(account.URI)
	if !ok || !accountIs(&account, check) {
		t.Fatal("Could not find expected status")
	}
	check, ok = cache.GetByURL(account.URL)
	if !ok || !accountIs(&account, check) {
		t.Fatal("Could not find expected status")
	}
}

func accountIs(account1, account2 *gtsmodel.Account) bool {
	return account1.ID == account2.ID && account1.URI == account2.URI && account1.URL == account2.URL
}
