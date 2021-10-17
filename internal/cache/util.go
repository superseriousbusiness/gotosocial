package cache

import "codeberg.org/gruf/go-cache"

func lmSetIf(lm *cache.LookupMap, id string, key string, origKey string) {
	if key != "" {
		lm.Set(id, key, origKey)
	}
}

func lmDeleteIf(lm *cache.LookupMap, id string, key string) {
	if key != "" {
		lm.Delete(id, key)
	}
}
