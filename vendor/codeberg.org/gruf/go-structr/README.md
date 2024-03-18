# go-structr

A performant struct caching library with automated indexing by arbitrary combinations of fields, including support for negative results (errors!). An example use case is in database lookups.

Under the hood, go-structr maintains a hashmap per index, where each hashmap is a hashmap keyed with either 32bit or 64bit (default) hash checksum of the inputted raw index keys. The hash checksum size can be controlled by the following Go build-tags: `structr_32bit_hash`

Some example code of how you can use `go-structr` in your application:
```golang
type Cached struct {
    Username    string
    Domain      string
    URL         string
    CountryCode int
}

var c structr.Cache[*Cached]

c.Init(structr.Config[*Cached]{

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
    CopyValue: func(c *Cached) *Cached {
        c2 := new(Cached)
        *c2 = *c
        return c2
    },

    // User defined invalidation hook.
    Invalidate: func(c *Cached) {
        log.Println("invalidated:", c)
    },
})

var url string

// Load value from cache, with callback function to hydrate
// cache if value cannot be found under index name with key.
// Negative (error) results are also cached, with user definable
// errors to ignore from caching (e.g. context deadline errs).
value, err := c.LoadOne("URL", func() (*Cached, error) {
    return dbType.SelectByURL(url)
}, url)
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

// Invalidate all cached results stored under
// provided index name with give field value(s).
c.Invalidate("CountryCode", 42)
```

This is a core underpinning of [GoToSocial](https://github.com/superseriousbusiness/gotosocial)'s performance.