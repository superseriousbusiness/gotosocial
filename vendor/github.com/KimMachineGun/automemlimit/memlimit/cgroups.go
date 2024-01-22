package memlimit

import (
	"errors"
)

var (
	// ErrNoCgroup is returned when the process is not in cgroup.
	ErrNoCgroup = errors.New("process is not in cgroup")
	// ErrCgroupsNotSupported is returned when the system does not support cgroups.
	ErrCgroupsNotSupported = errors.New("cgroups is not supported on this system")
)
