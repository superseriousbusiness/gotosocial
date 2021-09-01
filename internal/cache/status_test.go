package cache_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusCacheTestSuite struct {
	suite.Suite
	data  map[string]*gtsmodel.Status
	cache *cache.StatusCache
}

func (suite *StatusCacheTestSuite) SetupSuite() {
	suite.data = testrig.NewTestStatuses()
}

func (suite *StatusCacheTestSuite) SetupTest() {
	suite.cache = cache.NewStatusCache()
}

func (suite *StatusCacheTestSuite) TearDownTest() {
	suite.data = nil
	suite.cache = nil
}

func (suite *StatusCacheTestSuite) TestStatusCache() {
	for _, status := range suite.data {
		// Place in the cache
		suite.cache.Put(status)
	}

	for _, status := range suite.data {
		var ok bool
		var check *gtsmodel.Status

		// Check we can retrieve
		check, ok = suite.cache.GetByID(status.ID)
		if !ok && !statusIs(status, check) {
			suite.Fail("Failed to fetch expected account with ID: %s", status.ID)
		}
		check, ok = suite.cache.GetByURI(status.URI)
		if status.URI != "" && !ok && !statusIs(status, check) {
			suite.Fail("Failed to fetch expected account with URI: %s", status.URI)
		}
		check, ok = suite.cache.GetByURL(status.URL)
		if status.URL != "" && !ok && !statusIs(status, check) {
			suite.Fail("Failed to fetch expected account with URL: %s", status.URL)
		}
	}
}

func TestStatusCache(t *testing.T) {
	suite.Run(t, &StatusCacheTestSuite{})
}

func statusIs(status1, status2 *gtsmodel.Status) bool {
	return status1.ID == status2.ID && status1.URI == status2.URI && status1.URL == status2.URL
}
