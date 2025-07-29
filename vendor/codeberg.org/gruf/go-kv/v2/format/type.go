package format

import (
	"reflect"
	"strings"
)

// typenode ...
type typenode struct {
	typeinfo
	parent *typenode
}

// typeinfo ...
type typeinfo struct {
	rtype reflect.Type
	flags reflect_flag
}

// new_typenode returns a new typenode{} with reflect.Type and flags.
func new_typenode(t reflect.Type, flags reflect_flag) typenode {
	return typenode{typeinfo: typeinfo{
		rtype: t,
		flags: flags,
	}}
}

// key returns data (i.e. type value info)
// to store a FormatFunc under in a cache.
func (n typenode) key() typeinfo {
	return n.typeinfo
}

// indirect returns whether reflect_flagIndir is set for given type flags.
func (n typenode) indirect() bool {
	return n.flags&reflect_flagIndir != 0
}

// iface_indir returns the result of abi.Type{}.IfaceIndir() for underlying type.
func (n typenode) iface_indir() bool {
	return abi_Type_IfaceIndir(n.rtype)
}

// next ...
func (n typenode) next(t reflect.Type, flags reflect_flag) typenode {
	child := new_typenode(t, flags)
	child.parent = &n
	return child
}

// visit ...
func (n typenode) visit() bool {
	t := n.rtype

	// Check if type is already encountered further up tree.
	for node := n.parent; node != nil; node = node.parent {
		if node.rtype == t {
			return false
		}
	}

	return true
}

// needs_typestr returns whether the type contained in the
// receiving typenode{} needs type string information prefixed
// when the TypeMask argument flag bit is set. Certain types
// don't need this as the parent type already indicates this.
func (n typenode) needs_typestr() bool {
	if n.parent == nil {
		return true
	}
	switch p := n.parent.rtype; p.Kind() {
	case reflect.Pointer:
		return n.parent.needs_typestr()
	case reflect.Slice,
		reflect.Array,
		reflect.Map:
		return false
	default:
		return true
	}
}

// typestr_with_ptrs returns the type string for
// current typenode{} with asterisks for pointers.
func (n typenode) typestr_with_ptrs() string {
	t := n.rtype

	// Check for parent.
	if n.parent == nil {
		return t.String()
	}

	// Get parent type.
	p := n.parent.rtype

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
// current typenode{} with ampersands for pointers.
func (n typenode) typestr_with_refs() string {
	t := n.rtype

	// Check for parent.
	if n.parent == nil {
		return t.String()
	}

	// Get parent type.
	p := n.parent.rtype

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
