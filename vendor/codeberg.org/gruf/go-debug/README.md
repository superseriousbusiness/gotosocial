# go-debug

This library provides a very simple method for compile-time or runtime determined debug checks, set using build tags.

The compile-time checks use Go constants, so when disabled your debug code will not be compiled.

The possible build tags are:

- "debug" || "" = debug determined at compile-time

- "debugenv" = debug determined at runtime using the $DEBUG environment variable

An example for how this works in practice can be seen by the following code:

```
func main() {
	println("debug.DEBUG() =", debug.DEBUG())
}
```

```
# Debug determined at compile-time, it is disabled
$ go run .
debug.DEBUG() = false

# Debug determined at compile-time, it is enabled
$ go run -tags=debug .
debug.DEBUG() = true

# Debug determined at runtime, $DEBUG is not set
$ go run -tags=debugenv .
debug.DEBUG() = false

# Debug determined at runtime, $DEBUG is set
$ DEBUG=y go run -tags=debugenv .
debug.DEBUG() = true
```