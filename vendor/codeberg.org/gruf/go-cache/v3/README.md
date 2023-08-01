# go-cache

Provides access to simple, yet flexible, and performant caches (with TTL if required) via the `cache.Cache{}` and `cache.TTLCache{}` interfaces.

## simple

A `cache.Cache{}` implementation with much more of the inner workings exposed. Designed to be used as a base for your own customizations, or used as-is.

## ttl

A `cache.TTLCache{}` implementation with much more of the inner workings exposed. Designed to be used as a base for your own customizations, or used as-is.

## result

`result.Cache` is an example of a more complex cache implementation using `ttl.Cache{}` as its underpinning.

It provides caching specifically of loadable struct types, with automatic keying by multiple different field members and caching of negative (error) values. All useful when wrapping, for example, a database.