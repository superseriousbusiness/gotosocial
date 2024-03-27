# go-structr

A library with a series of performant data types with automated struct value indexing. Indexing is supported via arbitrary combinations of fields, and in the case of the cache type, negative results (errors!) are also supported.

Under the hood, go-structr maintains a hashmap per index, where each hashmap is a hashmap keyed by serialized input key type. This is handled by the incredibly performant serialization library [go-mangler](https://codeberg.org/gruf/go-mangler), which at this point in time supports just about **any** arbitrary type, so feel free to index by *anything*!

## Cache example

```golang
type Cached struct {
    Username    string
    Domain      string
    URL         string
    CountryCode int
}

var c structr.Cache[*Cached]

c.Init(structr.CacheConfig[*Cached]{

    // Fields this cached struct type
    // will be indexed and stored under.
    Indices: []structr.IndexConfig{
        {Fields: "Username,Domain", AllowZero: true},
        {Fields: "URL"},
        {Fields: "CountryCode", Multiple: true},
    },

    // Maximum LRU cache size before
    // new entries cause evictions.
    MaxSize: 1000,

    // User provided value copy function to
    // reduce need for reflection + ensure
    // concurrency safety for returned values.
    Copy: func(c *Cached) *Cached {
        c2 := new(Cached)
        *c2 = *c
        return c2
    },

    // User defined invalidation hook.
    Invalidate: func(c *Cached) {
        log.Println("invalidated:", c)
    },
})

// Access and store indexes ahead-of-time for perf.
usernameDomainIndex := c.Index("Username,Domain")
urlIndex := c.Index("URL")
countryCodeIndex := c.Index("CountryCode")

var url string

// Generate URL index key.
urlKey := urlIndex.Key(url)

// Load value from cache, with callback function to hydrate
// cache if value cannot be found under index name with key.
// Negative (error) results are also cached, with user definable
// errors to ignore from caching (e.g. context deadline errs).
value, err := c.LoadOne(urlIndex, func() (*Cached, error) {
    return dbType.SelectByURL(url)
}, urlKey)
if err != nil {
    return nil, err
}

// Store value in cache, only if provided callback
// function returns without error. Passes value through
// invalidation hook regardless of error return value.
//
// On success value will be automatically added to and
// accessible under all initially configured indices.
if err := c.Store(value, func() error {
    return dbType.Insert(value)
}); err != nil {
    return nil, err
}

// Generate country code index key.
countryCodeKey := countryCodeIndex.Key(42)

// Invalidate all cached results stored under
// provided index name with give field value(s).
c.Invalidate(countryCodeIndex, countryCodeKey)
```

## Queue example

```golang

type Queued struct{
    Username    string
    Domain      string
    URL         string
    CountryCode int
}

var q structr.Queue[*Queued]

q.Init(structr.QueueConfig[*Cached]{

    // Fields this queued struct type
    // will be indexed and stored under.
    Indices: []structr.IndexConfig{
        {Fields: "Username,Domain", AllowZero: true},
        {Fields: "URL"},
        {Fields: "CountryCode", Multiple: true},
    },

    // User defined pop hook.
    Pop: func(c *Cached) {
        log.Println("popped:", c)
    },
})

// Access and store indexes ahead-of-time for perf.
usernameDomainIndex := q.Index("Username,Domain")
urlIndex := q.Index("URL")
countryCodeIndex := q.Index("CountryCode")

// ...
q.PushBack(Queued{
    Username:   "billybob",
    Domain:     "google.com",
    URL:        "https://some.website.here",
    CountryCode: 42,
})

// ...
queued, ok := q.PopFront()

// Generate country code index key.
countryCodeKey := countryCodeIndex.Key(42)

// ...
queuedByCountry := q.Pop(countryCodeIndex, countryCodeKey)
```

## Notes

This is a core underpinning of [GoToSocial](https://github.com/superseriousbusiness/gotosocial)'s performance.