package format

import "codeberg.org/gruf/go-xunsafe"

// iterSliceType returns a FormatFunc capable of iterating
// and formatting the given slice type currently in TypeIter{}.
// note this will fetch a sub-FormatFunc for the slice element
// type, and also handle special cases of []byte, []rune slices.
func (fmt *Formatter) iterSliceType(t xunsafe.TypeIter) FormatFunc {

	// Get nested element type.
	elem := t.Type.Elem()
	esz := elem.Size()

	// Get nested elem TypeIter{} with flags.
	flags := xunsafe.ReflectSliceElemFlags(elem)
	et := t.Child(elem, flags)

	// Get elem format func.
	fn := fmt.loadOrGet(et)
	if fn == nil {
		panic("unreachable")
	}

	if !needs_typestr(t) {
		return func(s *State) {
			ptr := s.P

			// Get data as unsafe slice header.
			hdr := (*xunsafe.Unsafeheader_Slice)(ptr)
			if hdr == nil || hdr.Data == nil {

				// Append nil.
				appendNil(s)
				return
			}

			// Prepend array brace.
			s.B = append(s.B, '[')

			if hdr.Len > 0 {
				for i := 0; i < hdr.Len; i++ {
					// Format at array index.
					offset := esz * uintptr(i)
					s.P = add(hdr.Data, offset)
					fn(s)

					// Append separator.
					s.B = append(s.B, ',')
				}

				// Drop final comma.
				s.B = s.B[:len(s.B)-1]
			}

			// Append array brace.
			s.B = append(s.B, ']')
		}
	}

	// Slice type string with ptrs / refs.
	typestrPtrs := typestr_with_ptrs(t)
	typestrRefs := typestr_with_refs(t)

	return func(s *State) {
		ptr := s.P

		// Get data as unsafe slice header.
		hdr := (*xunsafe.Unsafeheader_Slice)(ptr)
		if hdr == nil || hdr.Data == nil {

			// Append nil value with type.
			appendNilType(s, typestrPtrs)
			return
		}

		// Open / close braces.
		var open, close uint8
		open, close = '[', ']'

		// Include type info.
		if s.A.WithType() {
			s.B = append(s.B, typestrRefs...)
			open, close = '{', '}'
		}

		// Prepend array brace.
		s.B = append(s.B, open)

		if hdr.Len > 0 {
			for i := 0; i < hdr.Len; i++ {
				// Format at array index.
				offset := esz * uintptr(i)
				s.P = add(hdr.Data, offset)
				fn(s)

				// Append separator.
				s.B = append(s.B, ',')
			}

			// Drop final comma.
			s.B = s.B[:len(s.B)-1]
		}

		// Append array brace.
		s.B = append(s.B, close)
	}
}

func wrapByteSlice(t xunsafe.TypeIter, fn FormatFunc) FormatFunc {
	if !needs_typestr(t) {
		return func(s *State) {
			if s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII() {
				appendString(s, *(*string)(s.P))
			} else {
				fn(s)
			}
		}
	}
	typestr := typestr_with_ptrs(t)
	return func(s *State) {
		if s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII() {
			if s.A.WithType() {
				s.B = append(s.B, "("+typestr+")("...)
				appendString(s, *(*string)(s.P))
				s.B = append(s.B, ")"...)
			} else {
				appendString(s, *(*string)(s.P))
			}
		} else {
			fn(s)
		}
	}
}

func wrapRuneSlice(t xunsafe.TypeIter, fn FormatFunc) FormatFunc {
	if !needs_typestr(t) {
		return func(s *State) {
			if s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII() {
				appendString(s, string(*(*[]rune)(s.P)))
			} else {
				fn(s)
			}
		}
	}
	typestr := typestr_with_ptrs(t)
	return func(s *State) {
		if s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII() {
			if s.A.WithType() {
				s.B = append(s.B, "("+typestr+")("...)
				appendString(s, string(*(*[]rune)(s.P)))
				s.B = append(s.B, ")"...)
			} else {
				appendString(s, string(*(*[]rune)(s.P)))
			}
		} else {
			fn(s)
		}
	}
}
