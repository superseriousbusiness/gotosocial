package state

import (
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// State provides a means of dependency injection and sharing of resources
// across different subpackages of the GoToSocial codebase. DO NOT assume
// that any particular field will be initialized if you are accessing this
// during initialization. A pointer to a State{} is often passed during
// subpackage initialization, while the returned subpackage type will later
// then be set and stored within the State{} itself.
type State struct {
	// Caches provides access to this state's collection of caches.
	Caches cache.Caches

	// DB provides access to the database.
	DB db.DB

	// prevent pass-by-value.
	_ nocopy
}

// nocopy when embedded will signal linter to
// error on pass-by-value of parent struct.
type nocopy struct{}

func (*nocopy) Lock() {}

func (*nocopy) Unlock() {}
