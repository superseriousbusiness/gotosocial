package cache

// Hook defines a function hook that can be supplied as a callback.
type Hook[Key comparable, Value any] func(key Key, value Value)

func emptyHook[K comparable, V any](K, V) {}
