package timeline

import (
	"sync"
	"time"
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
	toRemove := len(s.data) - s.maxLength
	if toRemove > 0 {
		s.Lock()
		defer s.Unlock()
		oldest := time.Now()
		oldestIDs := make([]string, toRemove)
		for id, post := range s.data {
			if post.createdAt.Before(oldest) {
				oldest = post.createdAt
				oldestIDs = append(oldestIDs, id)
			}
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
