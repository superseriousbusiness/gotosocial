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

package search_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/search"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type SearchGetTestSuite struct {
	SearchStandardTestSuite
}

func (suite *SearchGetTestSuite) testSearch(query string, resolve bool, expectedHTTPStatus int) (*apimodel.SearchResult, error) {
	requestPath := fmt.Sprintf("%s?q=%s&resolve=%t", search.BasePathV1, query, resolve)
	recorder := httptest.NewRecorder()

	ctx := suite.newContext(recorder, requestPath)

	suite.searchModule.SearchGETHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()

	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		return nil, fmt.Errorf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	searchResult := &apimodel.SearchResult{}
	if err := json.Unmarshal(b, searchResult); err != nil {
		return nil, err
	}

	return searchResult, nil
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByURI() {
	query := "https://unknown-instance.com/users/brand_new_person"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestring() {
	query := "@brand_new_person@unknown-instance.com"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringUppercase() {
	query := "@Some_User@example.org"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringNoLeadingAt() {
	query := "brand_new_person@unknown-instance.com"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringWithHostMetaFallback() {
	resolve := true

	// Grab the webfinger cache so we can check it's empty when we start
	wc := suite.state.Caches.GTS.Webfinger()
	suite.Equal(0, wc.Len(), "expected webfinger cache to be empty")

	// Do a first search request for a valid user
	_, err := suite.testSearch("someone@misconfigured-instance.com", resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check that the webfinger cache got populated
	suite.Equal(1, wc.Len(), "expected webfinger cache to only have 1 entry")

	wc.Lock()
	suite.True(wc.Cache.Has("misconfigured-instance.com"), "expected webfinger cache to have an entry for misconfigured-instance")
	// Store when the current item is going to expire. This goes through Cache.Get
	// to avoid updating the item expiry. We can then check later if the cache item
	// got updated as expected
	entry, _ := wc.Cache.Get("misconfigured-instance.com")
	wc.Unlock()
	originalExpiryTime := entry.Expiry

	// Do a request for the same domain but a valid user
	_, err = suite.testSearch("nobody@misconfigured-instance.com", resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check the cache state, there should still only be 1 entry
	suite.Equal(1, wc.Len(), "expected webfinger cache to still only have 1 entry")

	wc.Lock()
	suite.True(wc.Cache.Has("misconfigured-instance.com"), "expected webfinger cache to still have an entry for misconfigured-instance")
	// Fetch when the cached item is going to expire. This should now
	// be different from the original entry
	entryNew, _ := wc.Cache.Get("misconfigured-instance.com")
	wc.Unlock()
	renewedExpiryTime := entryNew.Expiry

	// Check the TTL has changed, indicating we renewed the cached entry
	suite.NotEqual(renewedExpiryTime, originalExpiryTime, "expected webfinger cache entry to have its TTL extended")

	// Do a request for the same domain but an invalid user
	_, err = suite.testSearch("invalid@misconfigured-instance.com", resolve, http.StatusInternalServerError)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Check the cache state again for good measure
	suite.Equal(1, wc.Len(), "expected webfinger cache to still only have 1 entry")

	wc.Lock()
	suite.True(wc.Cache.Has("misconfigured-instance.com"), "expected webfinger cache to still have an entry for misconfigured-instance")
	// Fetch when the cached item is going to expire. This should not have
	// changed from the previous request, as a failed webfinger should not
	// renew the item expiry
	entryFailed, _ := wc.Cache.Get("misconfigured-instance.com")
	wc.Unlock()
	expiryFailed := entryFailed.Expiry

	// The TTL should not have changed compared to the previous fetch
	suite.Equal(expiryFailed, renewedExpiryTime, "expected webfinger cache item to not have its TTL extended")
}

func (suite *SearchGetTestSuite) TestSearchRemoteAccountByNamestringNoResolve() {
	query := "@brand_new_person@unknown-instance.com"
	resolve := false

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByNamestring() {
	query := "@the_mighty_zork"
	resolve := false

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByNamestringWithDomain() {
	query := "@the_mighty_zork@localhost:8080"
	resolve := false

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchNonexistingLocalAccountByNamestringResolveTrue() {
	query := "@somone_made_up@localhost:8080"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByURI() {
	query := "http://localhost:8080/users/the_mighty_zork"
	resolve := false

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchLocalAccountByURL() {
	query := "http://localhost:8080/@the_mighty_zork"
	resolve := false

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Accounts, 1) {
		suite.FailNow("expected 1 account in search results but got 0")
	}

	gotAccount := searchResult.Accounts[0]
	suite.NotNil(gotAccount)
}

func (suite *SearchGetTestSuite) TestSearchNonexistingLocalAccountByURL() {
	query := "http://localhost:8080/@the_shmighty_shmork"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
}

func (suite *SearchGetTestSuite) TestSearchStatusByURL() {
	query := "https://turnip.farm/users/turniplover6969/statuses/70c53e54-3146-42d5-a630-83c8b6c7c042"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if !suite.Len(searchResult.Statuses, 1) {
		suite.FailNow("expected 1 status in search results but got 0")
	}

	gotStatus := searchResult.Statuses[0]
	suite.NotNil(gotStatus)
}

func (suite *SearchGetTestSuite) TestSearchBlockedDomainURL() {
	query := "https://replyguys.com/@someone"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func (suite *SearchGetTestSuite) TestSearchBlockedDomainNamestring() {
	query := "@someone@replyguys.com"
	resolve := true

	searchResult, err := suite.testSearch(query, resolve, http.StatusOK)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(searchResult.Accounts, 0)
	suite.Len(searchResult.Statuses, 0)
	suite.Len(searchResult.Hashtags, 0)
}

func TestSearchGetTestSuite(t *testing.T) {
	suite.Run(t, &SearchGetTestSuite{})
}
