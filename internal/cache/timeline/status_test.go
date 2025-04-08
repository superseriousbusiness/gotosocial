package timeline

import (
	"slices"
	"testing"

	"codeberg.org/gruf/go-structr"
	"github.com/stretchr/testify/assert"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
