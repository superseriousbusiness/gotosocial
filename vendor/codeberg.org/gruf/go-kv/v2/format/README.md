# format

a low-level string formatting library that takes arbitrary input types as interfaces, and arguments as a struct. this does not contain any printf-like argument parsing, only log-friendly serialization of arbitrary input arguments. (noting that our output is noticably more log-friendly for struct / map types than stdlib "fmt").

benchmarks:
```shell
goos: linux
goarch: amd64
pkg: codeberg.org/gruf/go-kv/v2/format
cpu: AMD Ryzen 7 7840U w/ Radeon  780M Graphics

# go-kv/v2/format (i.e. latest)
BenchmarkFormatV2Append
BenchmarkFormatV2Append-16                590422              1977 ns/op             488 B/op         23 allocs/op
BenchmarkFormatV2AppendVerbose
BenchmarkFormatV2AppendVerbose-16         375628              2981 ns/op            1704 B/op         45 allocs/op

# go-kv/format (i.e. v1)
BenchmarkFormatAppend
BenchmarkFormatAppend-16                  208357              5883 ns/op            2624 B/op        169 allocs/op
BenchmarkFormatAppendVerbose
BenchmarkFormatAppendVerbose-16            35916             33563 ns/op            3734 B/op        208 allocs/op

# fmt (i.e. stdlib)
BenchmarkFmtAppend
BenchmarkFmtAppend-16                     147722              8418 ns/op            4747 B/op        191 allocs/op
BenchmarkFmtAppendVerbose
BenchmarkFmtAppendVerbose-16              167112              7238 ns/op            4401 B/op        178 allocs/op

PASS
ok      codeberg.org/gruf/go-kv/v2/format
```
