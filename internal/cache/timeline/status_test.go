package timeline

import (
	"slices"
	"testing"

	"codeberg.org/gruf/go-structr"
	"github.com/stretchr/testify/assert"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

var testStatusMeta = []*StatusMeta{
	{ID: "06B19VYTHEG01F3YW13RQE0QM8"},
	{ID: "06B19VYTJFT0KDWT5C1CPY0XNC"},
	{ID: "06B19VYTJ6WZQPRVNJHPEZH04W"},
	{ID: "06B19VYTJPKGG8JYCR1ENAV7KC"},
	{ID: "06B19VYTHRR8S35QXC5A6VE2YW"},
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
	for _, value := range t.cache.Range(structr.Desc) {
		if value.ID == id {
			return true
		}
	}
	return false
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
