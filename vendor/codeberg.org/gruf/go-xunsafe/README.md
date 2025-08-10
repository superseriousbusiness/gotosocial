# go-xunsafe

xunafe as in EXTRA UNSAFE. this exposes reflect and internal ABI package data structures in a manner that is kept up-to-date and version gated by the compile-time Go version. This package performs no validity checks on the data you provide to it, it is expected that you know what you are doing. Not to mention that some of the internal ABI data must only be exposed VERY carefully given that it is in non garbage-collected memory.
