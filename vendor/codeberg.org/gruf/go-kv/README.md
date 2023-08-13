# go-kv

This library provides a key-value field structure `kv.Field{}` that plays well with the `"fmt"` package. It gives an easy means of appending key-value fields to log entries, in a manner that also happens to look nice! (it's not far removed from using a `map[string]interface{}`).

The formatting for these key-value fields is handled by the `"fmt"` package by default. If you set the `kvformat` build tag then it will use a custom formatting library found under `format/`. You can see the benchmarks for both below.

benchmarks:
```
grufwub @ ~/Projects/main/go-kv
--> go test -run=none -benchmem -bench=.*
goos: linux
goarch: amd64
pkg: codeberg.org/gruf/go-kv
cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
BenchmarkFieldAppendMulti-8       125241              9389 ns/op             849 B/op         98 allocs/op
BenchmarkFieldStringMulti-8       113227             10444 ns/op            3029 B/op        120 allocs/op
BenchmarkFieldFprintfMulti-8      147915              7448 ns/op            1121 B/op        115 allocs/op
BenchmarkFieldsAppend-8           189126              6255 ns/op             849 B/op         98 allocs/op
BenchmarkFieldsString-8           166219              6517 ns/op            3798 B/op        100 allocs/op
PASS
ok      codeberg.org/gruf/go-kv 6.169s

grufwub @ ~/Projects/main/go-kv
--> go test -run=none -benchmem -bench=.* -tags=kvformat                                                                                                                                                                                       
goos: linux
goarch: amd64
pkg: codeberg.org/gruf/go-kv
cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
BenchmarkFieldAppendMulti-8       190161              5709 ns/op             592 B/op         56 allocs/op
BenchmarkFieldStringMulti-8       161763              7930 ns/op            3240 B/op         95 allocs/op
BenchmarkFieldFprintfMulti-8      181557              6207 ns/op            1120 B/op        115 allocs/op
BenchmarkFieldsAppend-8           247052              4580 ns/op             592 B/op         56 allocs/op
BenchmarkFieldsString-8           231235              5103 ns/op            1768 B/op         58 allocs/op
PASS
ok      codeberg.org/gruf/go-kv 6.134s
```
