package xunsafe

import "reflect"

// TypeIter provides a simple wrapper for
// a means of following reflected types.
type TypeIter struct {
	TypeInfo
	Parent *TypeIter
}

// TypeInfo wraps reflect type information
// along with flags specifying further details
// necessary due to type nesting.
type TypeInfo struct {
	Type reflect.Type
	Flag Reflect_flag
}

// ToTypeIter creates a new TypeIter{} from reflect type and flags.
func ToTypeIter(rtype reflect.Type, flags Reflect_flag) TypeIter {
	return TypeIter{TypeInfo: TypeInfo{rtype, flags}}
}

// TypeIterFrom creates new TypeIter from interface value type.
// Note this will always assume the initial value passed to you
// will be coming from an interface.
func TypeIterFrom(a any) TypeIter {
	rtype := reflect.TypeOf(a)
	flags := ReflectIfaceElemFlags(rtype)
	return ToTypeIter(rtype, flags)
}

// Indirect returns whether Reflect_flagIndir is set on receiving TypeInfo{}.Flag.
func (t TypeInfo) Indirect() bool {
	return t.Flag&Reflect_flagIndir != 0
}

// IfaceIndir calls Abi_Type_IfaceIndir() on receiving TypeInfo{}.Type.
func (t TypeInfo) IfaceIndir() bool {
	return Abi_Type_IfaceIndir(t.Type)
}

// Child returns a new TypeIter{} for given type and flags, with parent pointing to receiver.
func (i TypeIter) Child(rtype reflect.Type, flags Reflect_flag) TypeIter {
	child := ToTypeIter(rtype, flags)
	child.Parent = &i
	return child
}
