# go-mangler

[Documentation](https://pkg.go.dev/codeberg.org/gruf/go-mangler).

To put it simply is a bit of an odd library. It aims to provide incredibly fast, unique string outputs for all default supported input data types during a given runtime instance. See `mangler.String()`for supported types.

It is useful, for example, for use as part of larger abstractions involving hashmaps. That was my particular usecase anyways...

This package does make liberal use of the "unsafe" package.

Benchmarks are below. Please note the more important thing to notice here is the relative difference in benchmark scores, the actual `ns/op`,`B/op`,`allocs/op` accounts for running through ~80 possible test cases, including some not-ideal situations.

The choice of libraries in the benchmark are just a selection of libraries that could be used in a similar manner to this one, i.e. serializing in some manner.

```
go test -run=none -benchmem -gcflags=all='-l=4' -bench=.*                            
goos: linux
goarch: amd64
pkg: codeberg.org/gruf/go-mangler
cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
BenchmarkMangle-8                        1278526               966.0 ns/op             0 B/op          0 allocs/op
BenchmarkMangleKnown-8                   3443587               345.9 ns/op             0 B/op          0 allocs/op
BenchmarkJSON-8                           228962              4717 ns/op            1849 B/op         99 allocs/op
BenchmarkLoosy-8                          307194              3447 ns/op             776 B/op         65 allocs/op
BenchmarkFmt-8                            150254              7405 ns/op            1377 B/op        143 allocs/op
BenchmarkFxmackerCbor-8                   364411              3037 ns/op            1224 B/op        105 allocs/op
BenchmarkMitchellhHashStructure-8         102272             11268 ns/op            8996 B/op       1000 allocs/op
BenchmarkCnfStructhash-8                    6789            168703 ns/op          288301 B/op       5779 allocs/op
PASS
ok      codeberg.org/gruf/go-mangler    11.715s
```
