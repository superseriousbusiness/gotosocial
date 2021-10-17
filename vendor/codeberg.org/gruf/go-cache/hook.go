package cache

// Hook defines a function hook that can be supplied as a callback
type Hook func(key string, value interface{})

func emptyHook(string, interface{}) {}
