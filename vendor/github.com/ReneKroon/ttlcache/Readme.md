## TTLCache - an in-memory cache with expiration

TTLCache is a simple key/value cache in golang with the following functions:

1. Thread-safe
2. Individual expiring time or global expiring time, you can choose
3. Auto-Extending expiration on `Get` -or- DNS style TTL, see `SkipTtlExtensionOnHit(bool)`
4. Fast and memory efficient
5. Can trigger callback on key expiration
6. Cleanup resources by calling `Close()` at end of lifecycle.

Note (issue #25): by default, due to historic reasons, the TTL will be reset on each cache hit and you need to explicitly configure the cache to use a TTL that will not get extended.

[![Build Status](https://travis-ci.org/ReneKroon/ttlcache.svg?branch=master)](https://travis-ci.org/ReneKroon/ttlcache)

#### Usage
```go
import (
  "time"
  "fmt"

  "github.com/ReneKroon/ttlcache"
)

func main () {
  newItemCallback := func(key string, value interface{}) {
		fmt.Printf("New key(%s) added\n", key)
  }
  checkExpirationCallback := func(key string, value interface{}) bool {
		if key == "key1" {
		    // if the key equals "key1", the value
		    // will not be allowed to expire
		    return false
		}
		// all other values are allowed to expire
		return true
	}
  expirationCallback := func(key string, value interface{}) {
		fmt.Printf("This key(%s) has expired\n", key)
	}

  cache := ttlcache.NewCache()
  defer cache.Close()
  cache.SetTTL(time.Duration(10 * time.Second))
  cache.SetExpirationCallback(expirationCallback)

  cache.Set("key", "value")
  cache.SetWithTTL("keyWithTTL", "value", 10 * time.Second)

  value, exists := cache.Get("key")
  count := cache.Count()
  result := cache.Remove("key")
}
```

#### TTLCache - Some design considerations

1. The complexity of the current cache is already quite high. Therefore i will not add 'convenience' features like an interface to supply a function to get missing keys. 
2. The locking should be done only in the functions of the Cache struct. Else data races can occur or recursive locks are needed, which are both unwanted.
3. I prefer correct functionality over fast tests. It's ok for new tests to take seconds to proof something.

#### Original Project

TTLCache was forked from [wunderlist/ttlcache](https://github.com/wunderlist/ttlcache) to add extra functions not avaiable in the original scope.
The main differences are:

1. A item can store any kind of object, previously, only strings could be saved
2. Optionally, you can add callbacks too: check if a value should expire, be notified if a value expires, and be notified when new values are added to the cache
3. The expiration can be either global or per item
4. Can exist items without expiration time
5. Expirations and callbacks are realtime. Don't have a pooling time to check anymore, now it's done with a heap.
