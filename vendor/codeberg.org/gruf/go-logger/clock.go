package logger

import (
	"sync"
	"time"

	"codeberg.org/gruf/go-nowish"
)

var (
	clock     = nowish.Clock{}
	clockOnce = sync.Once{}
)

// startClock starts the global nowish clock
func startClock() {
	clockOnce.Do(func() {
		clock.Start(time.Millisecond * 10)
		clock.SetFormat("2006-01-02 15:04:05")
	})
}
