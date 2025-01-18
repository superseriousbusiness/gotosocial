package util

import (
	"runtime"

	"golang.org/x/sys/cpu"
)

func CompilerSupported() bool {
	switch runtime.GOOS {
	case "linux", "android",
		"windows", "darwin",
		"freebsd", "netbsd", "dragonfly",
		"solaris", "illumos":
		break
	default:
		return false
	}
	switch runtime.GOARCH {
	case "amd64":
		return cpu.X86.HasSSE41
	case "arm64":
		return true
	default:
		return false
	}
}
