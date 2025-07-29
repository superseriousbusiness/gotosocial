package format

// iterSliceType returns a FormatFunc capable of iterating
// and formatting the given slice type currently in typenode{}.
// note this will fetch a sub-FormatFunc for the slice element
// type, and also handle special cases of []byte, []rune slices.
func (fmt *Formatter) iterSliceType(t typenode) FormatFunc {

	// Get nested element type.
	elem := t.rtype.Elem()
	esz := elem.Size()

	// Get nested elem typenode with flags.
	flags := reflect_slice_elem_flags(elem)
	et := t.next(elem, flags)

	// Get elem format func.
	fn := fmt.loadOrGet(et)
	if fn == nil {
		panic("unreachable")
	}

	if !t.needs_typestr() {
		return func(s *State) {
			ptr := s.P

			// Get data as unsafe slice header.
			hdr := (*unsafeheader_Slice)(ptr)
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
	typestrPtrs := t.typestr_with_ptrs()
	typestrRefs := t.typestr_with_refs()

	return func(s *State) {
		ptr := s.P

		// Get data as unsafe slice header.
		hdr := (*unsafeheader_Slice)(ptr)
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

func wrapByteSlice(t typenode, fn FormatFunc) FormatFunc {
	if !t.needs_typestr() {
		return func(s *State) {
			if s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII() {
				appendString(s, *(*string)(s.P))
			} else {
				fn(s)
			}
		}
	}
	typestr := t.typestr_with_ptrs()
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

func wrapRuneSlice(t typenode, fn FormatFunc) FormatFunc {
	if !t.needs_typestr() {
		return func(s *State) {
			if s.A.AsText() || s.A.AsQuotedText() || s.A.AsQuotedASCII() {
				appendString(s, string(*(*[]rune)(s.P)))
			} else {
				fn(s)
			}
		}
	}
	typestr := t.typestr_with_ptrs()
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
