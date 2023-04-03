//go:build linux
// +build linux

package memlimit

import (
	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/containerd/cgroups/v3/cgroup2"
)

const (
	cgroupMountPoint = "/sys/fs/cgroup"
)

// FromCgroup returns the memory limit based on the cgroups version on this system.
func FromCgroup() (uint64, error) {
	switch cgroups.Mode() {
	case cgroups.Legacy:
		return FromCgroupV1()
	case cgroups.Hybrid, cgroups.Unified:
		return FromCgroupV2()
	}
	return 0, ErrNoCgroup
}

// FromCgroupV1 returns the memory limit from the cgroup v1.
func FromCgroupV1() (uint64, error) {
	cg, err := cgroup1.Load(cgroup1.RootPath, cgroup1.WithHiearchy(
		cgroup1.SingleSubsystem(cgroup1.Default, cgroup1.Memory),
	))
	if err != nil {
		return 0, err
	}

	metrics, err := cg.Stat(cgroup1.IgnoreNotExist)
	if err != nil {
		return 0, err
	} else if metrics.Memory == nil {
		return 0, ErrNoLimit
	}

	return metrics.Memory.HierarchicalMemoryLimit, nil
}

// FromCgroupV2 returns the memory limit from the cgroup v2.
func FromCgroupV2() (uint64, error) {
	path, err := cgroup2.NestedGroupPath("")
	if err != nil {
		return 0, err
	}

	m, err := cgroup2.Load(path, cgroup2.WithMountpoint(cgroupMountPoint))
	if err != nil {
		return 0, err
	}

	stats, err := m.Stat()
	if err != nil {
		return 0, err
	} else if stats.Memory == nil {
		return 0, ErrNoLimit
	}

	return stats.Memory.UsageLimit, nil
}
