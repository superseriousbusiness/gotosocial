package httpclient

import "time"

// sleepch returns a blocking sleep channel and cancel function.
func sleepch(d time.Duration) (<-chan time.Time, func() bool) {
	t := time.NewTimer(d)
	return t.C, t.Stop
}
