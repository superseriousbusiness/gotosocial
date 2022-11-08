# go-mangler

[Documentation](https://pkg.go.dev/codeberg.org/gruf/go-mangler).

To put it simply is a bit of an odd library. It aims to provide incredibly fast, unique string outputs for all default supported input data types during a given runtime instance.

It is useful, for example, for use as part of larger abstractions involving hashmaps. That was my particular usecase anyways...

This package does make liberal use of the "unsafe" package.

Benchmarks are below. Those with missing values panicked during our set of benchmarks, usually a case of not handling nil values elegantly. Please note the more important thing to notice here is the relative difference in benchmark scores, the actual `ns/op`,`B/op`,`allocs/op` accounts for running through over 80 possible test cases, including some not-ideal situations.

The choice of libraries in the benchmark are just a selection of libraries that could be used in a similar manner to this one, i.e. serializing in some manner.

```
goos: linux
goarch: amd64
pkg: codeberg.org/gruf/go-mangler
cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
BenchmarkMangle
BenchmarkMangle-8                         723278              1593 ns/op            1168 B/op        120 allocs/op
BenchmarkMangleHash
BenchmarkMangleHash-8                     405380              2788 ns/op            4496 B/op        214 allocs/op
BenchmarkJSON
BenchmarkJSON-8                           199360              6116 ns/op            4243 B/op        142 allocs/op
BenchmarkBinary
BenchmarkBinary-8                         ------              ---- ns/op            ---- B/op        --- allocs/op
BenchmarkFmt
BenchmarkFmt-8                            168500              7111 ns/op            2256 B/op        161 allocs/op
BenchmarkKelindarBinary
BenchmarkKelindarBinary-8                 ------              ---- ns/op            ---- B/op        --- allocs/op
BenchmarkFxmackerCbor
BenchmarkFxmackerCbor-8                   361416              3255 ns/op            1495 B/op        122 allocs/op
BenchmarkMitchellhHashStructure
BenchmarkMitchellhHashStructure-8         117672             10493 ns/op            8443 B/op        961 allocs/op
BenchmarkCnfStructhash
BenchmarkCnfStructhash-8                    7078            161926 ns/op          288644 B/op       5843 allocs/op
PASS
ok      codeberg.org/gruf/go-mangler    14.377s
```