package timeline

import (
	"sort"
	"sync"
)

type sharedCache struct {
	data      map[string]*post
	maxLength int
	*sync.Mutex
}

func newSharedCache(maxLength int) *sharedCache {
	return &sharedCache{
		data:      make(map[string]*post),
		maxLength: maxLength,
	}
}

func (s *sharedCache) shrink() {
	// check if the length is longer than max size
	toRemove := len(s.data) - s.maxLength

	if toRemove > 0 {
		// we have stuff to remove so lock the map while we work
		s.Lock()
		defer s.Unlock()

		// we need to time-sort the map to remove the oldest entries
		// the below code gives us a slice of keys, arranged from newest to oldest
		postSlice := make([]*post, 0, len(s.data))
		for _, v := range s.data {
			postSlice = append(postSlice, v)
		}
		sort.Slice(postSlice, func(i int, j int) bool {
			return postSlice[i].createdAt.After(postSlice[j].createdAt)
		})

		// now for each entry we have to remove, delete the entry from the map by its status ID
		for i := 0; i < toRemove; i = i + 1 {
			statusID := postSlice[i].statusID
			delete(s.data, statusID)
		}
	}
}

func (s *sharedCache) put(post *post) {
	s.Lock()
	defer s.Unlock()
	s.data[post.statusID] = post
}

func (s *sharedCache) get(statusID string) *post {
	return s.data[statusID]
}
