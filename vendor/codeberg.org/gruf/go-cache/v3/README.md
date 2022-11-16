# go-cache

Provides access to a simple yet flexible, performant TTL cache via the `Cache{}` interface and `cache.New()`. Under the hood this is returning a `ttl.Cache{}`.

## ttl

A TTL cache implementation with much of the inner workings exposed, designed to be used as a base for your own customizations, or used as-is. Access via the base package `cache.New()` is recommended in the latter case, to prevent accidental use of unsafe methods.

## lookup

`lookup.Cache` is an example of a more complex cache implementation using `ttl.Cache{}` as its underpinning. It provides caching of items under multiple keys.

## result

`result.Cache` is an example of a more complex cache implementation using `ttl.Cache{}` as its underpinning.

It provides caching specifically of loadable struct types, with automatic keying by multiple different field members and caching of negative (error) values. All useful when wrapping, for example, a database.