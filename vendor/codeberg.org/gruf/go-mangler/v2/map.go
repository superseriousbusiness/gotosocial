package mangler

import (
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

// iterMapType returns a Mangler capable of iterating
// and mangling the given map type currently in TypeIter{}.
// note this will fetch sub-Manglers for key / value types.
func iterMapType(t xunsafe.TypeIter) Mangler {

	// Key / value types.
	key := t.Type.Key()
	elem := t.Type.Elem()

	// Get nested k / v TypeIters with appropriate flags.
	flagsKey := xunsafe.ReflectMapKeyFlags(key)
	flagsVal := xunsafe.ReflectMapElemFlags(elem)
	kt := t.Child(key, flagsKey)
	vt := t.Child(elem, flagsVal)

	// Get key mangler.
	kfn := loadOrGet(kt)
	if kfn == nil {
		return nil
	}

	// Get value mangler.
	vfn := loadOrGet(vt)
	if vfn == nil {
		return nil
	}

	// Final map type.
	rtype := t.Type
	flags := t.Flag

	return func(buf []byte, ptr unsafe.Pointer) []byte {
		if ptr == nil || *(*unsafe.Pointer)(ptr) == nil {
			// Append nil indicator.
			buf = append(buf, '0')
			return buf
		}

		// Build reflect value, and then a map iterator.
		v := xunsafe.BuildReflectValue(rtype, ptr, flags)
		i := xunsafe.GetMapIter(v)

		// Before len.
		l := len(buf)

		// Append not-nil flag.
		buf = append(buf, '1')

		for i.Next() {
			// Pass to map key func.
			ptr = xunsafe.Map_Key(i)
			buf = kfn(buf, ptr)

			// Add key seperator.
			buf = append(buf, ':')

			// Pass to map elem func.
			ptr = xunsafe.Map_Elem(i)
			buf = vfn(buf, ptr)

			// Add comma seperator.
			buf = append(buf, ',')
		}

		if len(buf) != l {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}
