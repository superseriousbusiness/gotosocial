package state

import "github.com/superseriousbusiness/gotosocial/internal/log"

// nocopy when embedded will signal linter to
// error on pass-by-value of parent struct.
type nocopy struct{}

func (*nocopy) Lock() {}

func (*nocopy) Unlock() {}

// tryUntil will attempt to call 'do' for 'count' attempts, before panicking with 'msg'.
func tryUntil(msg string, count int, do func() bool) {
	for i := 0; i < count && !do(); i++ {
	}
	log.Panic("failed %s after %d tries", msg, count)
}
