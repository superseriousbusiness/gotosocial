package format

import (
	"reflect"
	"strings"

	"codeberg.org/gruf/go-xunsafe"
)

// visit ...
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

// needs_typestr returns whether the type contained in the
// receiving TypeIter{} needs type string information prefixed
// when the TypeMask argument flag bit is set. Certain types
// don't need this as the parent type already indicates this.
func needs_typestr(iter xunsafe.TypeIter) bool {
	if iter.Parent == nil {
		return true
	}
	switch p := iter.Parent.Type; p.Kind() {
	case reflect.Pointer:
		return needs_typestr(*iter.Parent)
	case reflect.Slice,
		reflect.Array,
		reflect.Map:
		return false
	default:
		return true
	}
}

// typestr_with_ptrs returns the type string for
// current xunsafe.TypeIter{} with asterisks for pointers.
func typestr_with_ptrs(iter xunsafe.TypeIter) string {
	t := iter.Type

	// Check for parent.
	if iter.Parent == nil {
		return t.String()
	}

	// Get parent type.
	p := iter.Parent.Type

	// If parent is not ptr, then
	// this was not a deref'd ptr.
	if p.Kind() != reflect.Pointer {
		return t.String()
	}

	// Return un-deref'd
	// ptr (parent) type.
	return p.String()
}

// typestr_with_refs returns the type string for
// current xunsafe.TypeIter{} with ampersands for pointers.
func typestr_with_refs(iter xunsafe.TypeIter) string {
	t := iter.Type

	// Check for parent.
	if iter.Parent == nil {
		return t.String()
	}

	// Get parent type.
	p := iter.Parent.Type

	var d int

	// Count number of dereferences.
	for p.Kind() == reflect.Pointer {
		p = p.Elem()
		d++
	}

	if d <= 0 {
		// Prefer just returning our
		// own string if possible, to
		// reduce number of strings
		// we need to allocate.
		return t.String()
	}

	// Value type str.
	str := t.String()

	// Return with type ptrs
	// symbolized by 'refs'.
	var buf strings.Builder
	buf.Grow(len(str) + d)
	for i := 0; i < d; i++ {
		buf.WriteByte('&')
	}
	buf.WriteString(str)
	return buf.String()
}
