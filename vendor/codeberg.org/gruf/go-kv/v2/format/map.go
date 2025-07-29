package format

import "unsafe"

// iterMapType returns a FormatFunc capable of iterating
// and formatting the given map type currently in typenode{}.
// note this will fetch sub-FormatFuncs for key / value types.
func (fmt *Formatter) iterMapType(t typenode) FormatFunc {

	// Key / value types.
	key := t.rtype.Key()
	elem := t.rtype.Elem()

	// Get nested k / v typenodes with appropriate flags.
	flagsKey := reflect_map_key_flags(key)
	flagsVal := reflect_map_elem_flags(elem)
	kt := t.next(t.rtype.Key(), flagsKey)
	vt := t.next(t.rtype.Elem(), flagsVal)

	// Get key format func.
	kfn := fmt.loadOrGet(kt)
	if kfn == nil {
		panic("unreachable")
	}

	// Get value format func.
	vfn := fmt.loadOrGet(vt)
	if vfn == nil {
		panic("unreachable")
	}

	// Final map type.
	rtype := t.rtype
	flags := t.flags

	// Map type string with ptrs / refs.
	typestrPtrs := t.typestr_with_ptrs()
	typestrRefs := t.typestr_with_refs()

	if !t.needs_typestr() {
		return func(s *State) {
			if s.P == nil || *(*unsafe.Pointer)(s.P) == nil {
				// Append nil.
				appendNil(s)
				return
			}

			// Build reflect value, and then a map iter.
			v := build_reflect_value(rtype, s.P, flags)
			i := map_iter(v)

			// Prepend object brace.
			s.B = append(s.B, '{')

			// Before len.
			l := len(s.B)

			for i.Next() {
				// Pass to key fn.
				s.P = map_key(i)
				kfn(s)

				// Add key seperator.
				s.B = append(s.B, '=')

				// Pass to elem fn.
				s.P = map_elem(i)
				vfn(s)

				// Add comma pair seperator.
				s.B = append(s.B, ',', ' ')
			}

			if len(s.B) != l {
				// Drop final ", ".
				s.B = s.B[:len(s.B)-2]
			}

			// Append object brace.
			s.B = append(s.B, '}')
		}
	}

	return func(s *State) {
		if s.P == nil || *(*unsafe.Pointer)(s.P) == nil {
			// Append nil value with type.
			appendNilType(s, typestrPtrs)
			return
		}

		// Build reflect value, and then a map iter.
		v := build_reflect_value(rtype, s.P, flags)
		i := map_iter(v)

		// Include type info.
		if s.A.WithType() {
			s.B = append(s.B, typestrRefs...)
		}

		// Prepend object brace.
		s.B = append(s.B, '{')

		// Before len.
		l := len(s.B)

		for i.Next() {
			// Pass to key fn.
			s.P = map_key(i)
			kfn(s)

			// Add key seperator.
			s.B = append(s.B, '=')

			// Pass to elem fn.
			s.P = map_elem(i)
			vfn(s)

			// Add comma pair seperator.
			s.B = append(s.B, ',', ' ')
		}

		if len(s.B) != l {
			// Drop final ", ".
			s.B = s.B[:len(s.B)-2]
		}

		// Append object brace.
		s.B = append(s.B, '}')
	}
}
