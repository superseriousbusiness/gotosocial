package mangler

import (
	"codeberg.org/gruf/go-xunsafe"
)

// visit checks if current type has already
// appeared in the TypeIter{}'s parent heirarchy.
func visit(iter xunsafe.TypeIter) bool {
	t := iter.Type

	// Check if type is already encountered further up tree.
	for node := iter.Parent; node != nil; node = node.Parent {
		if node.Type == t {
			return false
		}
	}

	return true
}
