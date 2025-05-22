// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package timeline

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"codeberg.org/gruf/go-structr"
	"github.com/stretchr/testify/assert"
)

var testStatusMeta = []*StatusMeta{
	{
		ID:               "06B19VYTHEG01F3YW13RQE0QM8",
		AccountID:        "06B1A61MZEBBVDSNPRJAA8F2C4",
		BoostOfID:        "06B1A5KQWGQ1ABM3FA7TDX1PK8",
		BoostOfAccountID: "06B1A6707818050PCK8SJAEC6G",
	},
	{
		ID:               "06B19VYTJFT0KDWT5C1CPY0XNC",
		AccountID:        "06B1A61MZN3ZQPZVNGEFBNYBJW",
		BoostOfID:        "06B1A5KQWSGFN4NNRV34KV5S9R",
		BoostOfAccountID: "06B1A6707HY8RAXG7JPCWR7XD4",
	},
	{
		ID:               "06B19VYTJ6WZQPRVNJHPEZH04W",
		AccountID:        "06B1A61MZY7E0YB6G01VJX8ERR",
		BoostOfID:        "06B1A5KQX5NPGSYGH8NC7HR1GR",
		BoostOfAccountID: "06B1A6707XCSAF0MVCGGYF9160",
	},
	{
		ID:               "06B19VYTJPKGG8JYCR1ENAV7KC",
		AccountID:        "06B1A61N07K1GC35PJ3CZ4M020",
		BoostOfID:        "06B1A5KQXG6ZCWE1R7C7KR7RYW",
		BoostOfAccountID: "06B1A67084W6SB6P6HJB7K5DSG",
	},
	{
		ID:               "06B19VYTHRR8S35QXC5A6VE2YW",
		AccountID:        "06B1A61N0P1TGQDVKANNG4AKP4",
		BoostOfID:        "06B1A5KQY3K839Z6S5HHAJKSWW",
		BoostOfAccountID: "06B1A6708SPJC3X3ZG3SGG8BN8",
	},
}

func TestStatusTimelinePreloader(t *testing.T) {
	ctx, cncl := context.WithCancel(t.Context())
	defer cncl()

	var tt StatusTimeline
	tt.Init(1000)

	// Start goroutine to add some
	// concurrent usage to preloader.
	var started atomic.Int32
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			tt.preloader.Check()
			started.Add(1)
		}
	}()

	// Wait until goroutine running.
	for started.Load() == 0 {
		time.Sleep(time.Millisecond)
	}

	// Variable to check whether
	// our hook funcs are called.
	var called bool
	reset := func() { called = false }

	// "no error" preloader hook.
	preloadNoErr := func() error {
		called = true
		return nil
	}

	// "error return" preloader hook.
	preloadErr := func() error {
		called = true
		return errors.New("oh no")
	}

	// Check that on fail does not mark as preloaded.
	err := tt.preloader.CheckPreload(preloadErr)
	assert.Error(t, err)
	assert.False(t, tt.preloader.Check())
	assert.True(t, called)
	reset()

	// Check that on success marks itself as preloaded.
	err = tt.preloader.CheckPreload(preloadNoErr)
	assert.NoError(t, err)
	assert.True(t, tt.preloader.Check())
	assert.True(t, called)
	reset()

	// Check that preload func not called again
	// if it's already in the 'preloaded' state.
	err = tt.preloader.CheckPreload(preloadErr)
	assert.NoError(t, err)
	assert.True(t, tt.preloader.Check())
	assert.False(t, called)
	reset()

	// Ensure that a clear operation
	// successfully unsets preloader.
	tt.preloader.Clear()
	assert.False(t, tt.preloader.Check())
	assert.False(t, called)
	reset()

	// Ensure that it can be marked as preloaded again.
	err = tt.preloader.CheckPreload(preloadNoErr)
	assert.NoError(t, err)
	assert.True(t, tt.preloader.Check())
	assert.True(t, called)
	reset()
}

func TestStatusTimelineLoadLimit(t *testing.T) {
	var tt StatusTimeline
	tt.Init(1000)

	// Prepare new context for the duration of this test.
	ctx, cncl := context.WithCancel(t.Context())
	defer cncl()

	// Clone the input test status data.
	data := slices.Clone(testStatusMeta)

	// Insert test data into timeline.
	_ = tt.cache.Insert(data...)

	// Manually mark timeline as 'preloaded'.
	tt.preloader.CheckPreload(func() error { return nil })

	// Craft a new page for selection,
	// setting placeholder min / max values
	// but in particular setting a limit
	// HIGHER than currently cached values.
	page := new(paging.Page)
	page.Min = paging.MinID(id.Lowest)
	page.Max = paging.MaxID(id.Highest)
	page.Limit = len(data) + 10

	// Load crafted page from the cache. This
	// SHOULD load all cached entries, then
	// generate an extra 10 statuses up to limit.
	apiStatuses, _, _, err := tt.Load(ctx,
		page,
		loadGeneratedStatusPage,
		loadStatusIDsFrom(data),
		nil, // no filtering
		func(status *gtsmodel.Status) (*apimodel.Status, error) { return new(apimodel.Status), nil },
	)
	assert.NoError(t, err)
	assert.Len(t, apiStatuses, page.Limit)
}

func TestStatusTimelineUnprepare(t *testing.T) {
	var tt StatusTimeline
	tt.Init(1000)

	// Clone the input test status data.
	data := slices.Clone(testStatusMeta)

	// Bodge some 'prepared'
	// models on test data.
	for _, meta := range data {
		meta.prepared = &apimodel.Status{}
	}

	// Insert test data into timeline.
	_ = tt.cache.Insert(data...)

	for _, meta := range data {
		// Unprepare this status with ID.
		tt.UnprepareByStatusIDs(meta.ID)

		// Check the item is unprepared.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value.prepared)
	}

	// Clear and reinsert.
	tt.cache.Clear()
	tt.cache.Insert(data...)

	for _, meta := range data {
		// Unprepare this status with boost ID.
		tt.UnprepareByStatusIDs(meta.BoostOfID)

		// Check the item is unprepared.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value.prepared)
	}

	// Clear and reinsert.
	tt.cache.Clear()
	tt.cache.Insert(data...)

	for _, meta := range data {
		// Unprepare this status with account ID.
		tt.UnprepareByAccountIDs(meta.AccountID)

		// Check the item is unprepared.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value.prepared)
	}

	// Clear and reinsert.
	tt.cache.Clear()
	tt.cache.Insert(data...)

	for _, meta := range data {
		// Unprepare this status with boost account ID.
		tt.UnprepareByAccountIDs(meta.BoostOfAccountID)

		// Check the item is unprepared.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value.prepared)
	}
}

func TestStatusTimelineRemove(t *testing.T) {
	var tt StatusTimeline
	tt.Init(1000)

	// Clone the input test status data.
	data := slices.Clone(testStatusMeta)

	// Insert test data into timeline.
	_ = tt.cache.Insert(data...)

	for _, meta := range data {
		// Remove this status with ID.
		tt.RemoveByStatusIDs(meta.ID)

		// Check the item is now gone.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value)
	}

	// Clear and reinsert.
	tt.cache.Clear()
	tt.cache.Insert(data...)

	for _, meta := range data {
		// Remove this status with boost ID.
		tt.RemoveByStatusIDs(meta.BoostOfID)

		// Check the item is now gone.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value)
	}

	// Clear and reinsert.
	tt.cache.Clear()
	tt.cache.Insert(data...)

	for _, meta := range data {
		// Remove this status with account ID.
		tt.RemoveByAccountIDs(meta.AccountID)

		// Check the item is now gone.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value)
	}

	// Clear and reinsert.
	tt.cache.Clear()
	tt.cache.Insert(data...)

	for _, meta := range data {
		// Remove this status with boost account ID.
		tt.RemoveByAccountIDs(meta.BoostOfAccountID)

		// Check the item is now gone.
		value := getStatusByID(&tt, meta.ID)
		assert.Nil(t, value)
	}
}

func TestStatusTimelineInserts(t *testing.T) {
	var tt StatusTimeline
	tt.Init(1000)

	// Clone the input test status data.
	data := slices.Clone(testStatusMeta)

	// Insert test data into timeline.
	l := tt.cache.Insert(data...)
	assert.Equal(t, len(data), l)

	// Ensure 'min' value status
	// in the timeline is expected.
	minID := minStatusID(data)
	assert.Equal(t, minID, minStatus(&tt).ID)

	// Ensure 'max' value status
	// in the timeline is expected.
	maxID := maxStatusID(data)
	assert.Equal(t, maxID, maxStatus(&tt).ID)

	// Manually mark timeline as 'preloaded'.
	tt.preloader.CheckPreload(func() error { return nil })

	// Specifically craft a boost of latest (i.e. max) status in timeline.
	boost := &gtsmodel.Status{ID: "06B1A00PQWDZZH9WK9P5VND35C", BoostOfID: maxID}

	// Insert boost into the timeline
	// checking for 'repeatBoost' notifier.
	repeatBoost := tt.InsertOne(boost, nil)
	assert.True(t, repeatBoost)

	// This should be the new 'max'
	// and have 'repeatBoost' set.
	newMax := maxStatus(&tt)
	assert.Equal(t, boost.ID, newMax.ID)
	assert.True(t, newMax.repeatBoost)

	// Specifically craft 2 boosts of some unseen status in the timeline.
	boost1 := &gtsmodel.Status{ID: "06B1A121YEX02S0AY48X93JMDW", BoostOfID: "unseen"}
	boost2 := &gtsmodel.Status{ID: "06B1A12TG2NTJC9P270EQXS08M", BoostOfID: "unseen"}

	// Insert boosts into the timeline, ensuring
	// first is not 'repeat', but second one is.
	repeatBoost1 := tt.InsertOne(boost1, nil)
	repeatBoost2 := tt.InsertOne(boost2, nil)
	assert.False(t, repeatBoost1)
	assert.True(t, repeatBoost2)
}

func TestStatusTimelineTrim(t *testing.T) {
	var tt StatusTimeline
	tt.Init(1000)

	// Clone the input test status data.
	data := slices.Clone(testStatusMeta)

	// Insert test data into timeline.
	_ = tt.cache.Insert(data...)

	// From here it'll be easier to have DESC sorted
	// test data for reslicing and checking against.
	slices.SortFunc(data, func(a, b *StatusMeta) int {
		const k = +1
		switch {
		case a.ID < b.ID:
			return +k
		case b.ID < a.ID:
			return -k
		default:
			return 0
		}
	})

	// Set manual cutoff for trim.
	tt.cut = len(data) - 1

	// Perform trim.
	tt.Trim()

	// The post trim length should be tt.cut
	assert.Equal(t, tt.cut, tt.cache.Len())

	// It specifically should have removed
	// the oldest (i.e. min) status element.
	minID := data[len(data)-1].ID
	assert.NotEqual(t, minID, minStatus(&tt).ID)
	assert.False(t, containsStatusID(&tt, minID))

	// Drop trimmed status.
	data = data[:len(data)-1]

	// Set smaller cutoff for trim.
	tt.cut = len(data) - 2

	// Perform trim.
	tt.Trim()

	// The post trim length should be tt.cut
	assert.Equal(t, tt.cut, tt.cache.Len())

	// It specifically should have removed
	// the oldest 2 (i.e. min) status elements.
	minID1 := data[len(data)-1].ID
	minID2 := data[len(data)-2].ID
	assert.NotEqual(t, minID1, minStatus(&tt).ID)
	assert.NotEqual(t, minID2, minStatus(&tt).ID)
	assert.False(t, containsStatusID(&tt, minID1))
	assert.False(t, containsStatusID(&tt, minID2))

	// Trim at desired length
	// should cause no change.
	before := tt.cache.Len()
	tt.Trim()
	assert.Equal(t, before, tt.cache.Len())
}

// loadStatusIDsFrom imitates loading of statuses of given IDs from the database, instead selecting
// statuses with appropriate IDs from the given slice of status meta, converting them to statuses.
func loadStatusIDsFrom(data []*StatusMeta) func(ids []string) ([]*gtsmodel.Status, error) {
	return func(ids []string) ([]*gtsmodel.Status, error) {
		var statuses []*gtsmodel.Status
		for _, id := range ids {
			i := slices.IndexFunc(data, func(s *StatusMeta) bool {
				return s.ID == id
			})
			if i < 0 || i >= len(data) {
				panic(fmt.Sprintf("could not find %s in %v", id, log.VarDump(data)))
			}
			statuses = append(statuses, &gtsmodel.Status{
				ID:               data[i].ID,
				AccountID:        data[i].AccountID,
				BoostOfID:        data[i].BoostOfID,
				BoostOfAccountID: data[i].BoostOfAccountID,
			})
		}
		return statuses, nil
	}
}

// loadGeneratedStatusPage imitates loading of a given page of statuses,
// simply generating new statuses until the given page's limit is reached.
func loadGeneratedStatusPage(page *paging.Page) ([]*gtsmodel.Status, error) {
	var statuses []*gtsmodel.Status
	for range page.Limit {
		statuses = append(statuses, &gtsmodel.Status{
			ID:               id.NewULID(),
			AccountID:        id.NewULID(),
			BoostOfID:        id.NewULID(),
			BoostOfAccountID: id.NewULID(),
		})
	}
	return statuses, nil
}

// containsStatusID returns whether timeline contains a status with ID.
func containsStatusID(t *StatusTimeline, id string) bool {
	return getStatusByID(t, id) != nil
}

// getStatusByID attempts to fetch status with given ID from timeline.
func getStatusByID(t *StatusTimeline, id string) *StatusMeta {
	for _, value := range t.cache.Range(structr.Desc) {
		if value.ID == id {
			return value
		}
	}
	return nil
}

// maxStatus returns the newest (i.e. highest value ID) status in timeline.
func maxStatus(t *StatusTimeline) *StatusMeta {
	var meta *StatusMeta
	for _, value := range t.cache.Range(structr.Desc) {
		meta = value
		break
	}
	return meta
}

// minStatus returns the oldest (i.e. lowest value ID) status in timeline.
func minStatus(t *StatusTimeline) *StatusMeta {
	var meta *StatusMeta
	for _, value := range t.cache.Range(structr.Asc) {
		meta = value
		break
	}
	return meta
}

// minStatusID returns the oldest (i.e. lowest value ID) status in metas.
func minStatusID(metas []*StatusMeta) string {
	var min string
	min = metas[0].ID
	for i := 1; i < len(metas); i++ {
		if metas[i].ID < min {
			min = metas[i].ID
		}
	}
	return min
}

// maxStatusID returns the newest (i.e. highest value ID) status in metas.
func maxStatusID(metas []*StatusMeta) string {
	var max string
	max = metas[0].ID
	for i := 1; i < len(metas); i++ {
		if metas[i].ID > max {
			max = metas[i].ID
		}
	}
	return max
}
