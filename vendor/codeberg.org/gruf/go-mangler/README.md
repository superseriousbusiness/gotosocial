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
BenchmarkMangle-8                         533011              2003 ns/op            1168 B/op        120 allocs/op
BenchmarkMangleKnown
BenchmarkMangleKnown-8                    817060              1458 ns/op            1168 B/op        120 allocs/op
BenchmarkJSON
BenchmarkJSON-8                           188637              5899 ns/op            4211 B/op        142 allocs/op
BenchmarkFmt
BenchmarkFmt-8                            162735              7053 ns/op            2257 B/op        161 allocs/op
BenchmarkFxmackerCbor
BenchmarkFxmackerCbor-8                   362403              3336 ns/op            1496 B/op        122 allocs/op
BenchmarkMitchellhHashStructure
BenchmarkMitchellhHashStructure-8         113982             10079 ns/op            8443 B/op        961 allocs/op
BenchmarkCnfStructhash
BenchmarkCnfStructhash-8                    7162            167613 ns/op          288619 B/op       5841 allocs/op
PASS
ok      codeberg.org/gruf/go-mangler    11.352s
```