# go-mangler

[Documentation](https://pkg.go.dev/codeberg.org/gruf/go-mangler).

To put it simply is a bit of an odd library. It aims to provide incredibly fast, unique string outputs for all default supported input data types during a given runtime instance.

It is useful, for example, for use as part of larger abstractions involving hashmaps. That was my particular usecase anyways...

This package does make liberal use of the "unsafe" package.

Benchmarks are below. Those with missing values panicked during our set of benchmarks, usually a case of not handling nil values elegantly. Please note the more important thing to notice here is the relative difference in benchmark scores, the actual `ns/op`,`B/op`,`allocs/op` accounts for running through over 80 possible test cases, including some not-ideal situations.

The choice of libraries in the benchmark are just a selection of libraries that could be used in a similar manner to this one, i.e. serializing in some manner.

```
go test -run=none -benchmem -gcflags=all='-l=4' -bench=.*                            
goos: linux
goarch: amd64
pkg: codeberg.org/gruf/go-mangler
cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
BenchmarkMangle
BenchmarkMangle-8                         877761              1323 ns/op               0 B/op          0 allocs/op
BenchmarkMangleKnown
BenchmarkMangleKnown-8                   1462954               814.5 ns/op             0 B/op          0 allocs/op
BenchmarkJSON
BenchmarkJSON-8                           199930              5910 ns/op            2698 B/op        119 allocs/op
BenchmarkLoosy
BenchmarkLoosy-8                          307575              3718 ns/op             664 B/op         53 allocs/op
BenchmarkBinary
BenchmarkBinary-8                         413216              2640 ns/op            3824 B/op        116 allocs/op
BenchmarkFmt
BenchmarkFmt-8                            133429              8568 ns/op            3010 B/op        207 allocs/op
BenchmarkFxmackerCbor
BenchmarkFxmackerCbor-8                   258562              4268 ns/op            2118 B/op        134 allocs/op
BenchmarkMitchellhHashStructure
BenchmarkMitchellhHashStructure-8          88941             13049 ns/op           10269 B/op       1096 allocs/op
BenchmarkCnfStructhash
BenchmarkCnfStructhash-8                    5586            179537 ns/op          290373 B/op       5863 allocs/op
PASS
ok      codeberg.org/gruf/go-mangler    12.469s
```
