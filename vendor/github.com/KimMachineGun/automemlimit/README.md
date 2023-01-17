# automemlimit

[![Go Reference](https://pkg.go.dev/badge/github.com/KimMachineGun/automemlimit.svg)](https://pkg.go.dev/github.com/KimMachineGun/automemlimit)
[![Go Report Card](https://goreportcard.com/badge/github.com/KimMachineGun/automemlimit)](https://goreportcard.com/report/github.com/KimMachineGun/automemlimit)
[![Test](https://github.com/KimMachineGun/automemlimit/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/KimMachineGun/automemlimit/actions/workflows/test.yml)

Automatically set `GOMEMLIMIT` to match Linux [cgroups(7)](https://man7.org/linux/man-pages/man7/cgroups.7.html) memory limit.

See more details about `GOMEMLIMIT` [here](https://tip.golang.org/doc/gc-guide#Memory_limit).

## Installation

```shell
go get github.com/KimMachineGun/automemlimit@latest
```

## Usage

```go
package main

// By default, it sets `GOMEMLIMIT` to 90% of cgroup's memory limit.
// You can find more details of its behavior from the doc comment of memlimit.SetGoMemLimitWithEnv.
import _ "github.com/KimMachineGun/automemlimit"
```

or

```go
package main

import "github.com/KimMachineGun/automemlimit/memlimit"

func init() {
	memlimit.SetGoMemLimitWithEnv()
	memlimit.SetGoMemLimit(0.9)
	memlimit.SetGoMemLimitWithProvider(memlimit.Limit(1024*1024), 0.9)
	memlimit.SetGoMemLimitWithProvider(memlimit.FromCgroup, 0.9)
	memlimit.SetGoMemLimitWithProvider(memlimit.FromCgroupV1, 0.9)
	memlimit.SetGoMemLimitWithProvider(memlimit.FromCgroupV2, 0.9)
}
```
