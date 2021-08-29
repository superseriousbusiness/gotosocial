package cache_test

import (
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func TestStatusCache(t *testing.T) {
	cache := cache.NewStatusCache()

	// Attempt to place a status
	status := gtsmodel.Status{
		ID:  "id",
		URI: "uri",
		URL: "url",
	}
	cache.Put(&status)

	var ok bool
	var check *gtsmodel.Status

	// Check we can retrieve
	check, ok = cache.GetByID(status.ID)
	if !ok || !statusIs(&status, check) {
		t.Fatal("Could not find expected status")
	}
	check, ok = cache.GetByURI(status.URI)
	if !ok || !statusIs(&status, check) {
		t.Fatal("Could not find expected status")
	}
	check, ok = cache.GetByURL(status.URL)
	if !ok || !statusIs(&status, check) {
		t.Fatal("Could not find expected status")
	}
}

func statusIs(status1, status2 *gtsmodel.Status) bool {
	return status1.ID == status2.ID && status1.URI == status2.URI && status1.URL == status2.URL
}
