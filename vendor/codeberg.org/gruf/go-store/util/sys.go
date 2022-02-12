package util

import "syscall"

// RetryOnEINTR is a low-level filesystem function for retrying syscalls on O_EINTR received
func RetryOnEINTR(do func() error) error {
	for {
		err := do()
		if err == syscall.EINTR {
			continue
		}
		return err
	}
}
