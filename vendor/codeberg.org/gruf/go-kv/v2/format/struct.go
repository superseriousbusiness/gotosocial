package format

import "codeberg.org/gruf/go-xunsafe"

// field stores the minimum necessary
// data for iterating and formatting
// each field in a given struct.
type field struct {
	format FormatFunc
	name   string
	offset uintptr
}

// iterStructType returns a FormatFunc capable of iterating
// and formatting the given struct type currently in TypeIter{}.
// note this will fetch sub-FormatFuncs for each struct field.
func (fmt *Formatter) iterStructType(t xunsafe.TypeIter) FormatFunc {
	// Number of struct fields.
	n := t.Type.NumField()

	// Gather format functions.
	fields := make([]field, n)
	for i := 0; i < n; i++ {

		// Get struct field at index.
		sfield := t.Type.Field(i)
		rtype := sfield.Type

		// Get nested field TypeIter with appropriate flags.
		flags := xunsafe.ReflectStructFieldFlags(t.Flag, rtype)
		ft := t.Child(sfield.Type, flags)

		// Get field format func.
		fn := fmt.loadOrGet(ft)
		if fn == nil {
			panic("unreachable")
		}

		// Set field info.
		fields[i] = field{
			format: fn,
			name:   sfield.Name,
			offset: sfield.Offset,
		}
	}

	// Handle no. fields.
	switch len(fields) {
	case 0:
		return emptyStructType(t)
	case 1:
		return iterSingleFieldStructType(t, fields[0])
	default:
		return iterMultiFieldStructType(t, fields)
	}
}

func emptyStructType(t xunsafe.TypeIter) FormatFunc {
	if !needs_typestr(t) {
		return func(s *State) {
			// Append empty object.
			s.B = append(s.B, "{}"...)
		}
	}

	// Struct type string with refs.
	typestr := typestr_with_refs(t)

	// Append empty object
	// with type information.
	return func(s *State) {
		if s.A.WithType() {
			s.B = append(s.B, typestr...)
		}
		s.B = append(s.B, "{}"...)
	}
}

func iterSingleFieldStructType(t xunsafe.TypeIter, field field) FormatFunc {
	if field.format == nil {
		panic("nil func")
	}

	if !needs_typestr(t) {
		return func(s *State) {
			// Wrap 'fn' with braces + field name.
			s.B = append(s.B, "{"+field.name+"="...)
			field.format(s)
			s.B = append(s.B, "}"...)
		}
	}

	// Struct type string with refs.
	typestr := typestr_with_refs(t)

	return func(s *State) {
		// Include type info.
		if s.A.WithType() {
			s.B = append(s.B, typestr...)
		}

		// Wrap 'fn' with braces + field name.
		s.B = append(s.B, "{"+field.name+"="...)
		field.format(s)
		s.B = append(s.B, "}"...)
	}
}

func iterMultiFieldStructType(t xunsafe.TypeIter, fields []field) FormatFunc {
	for _, field := range fields {
		if field.format == nil {
			panic("nil func")
		}
	}

	if !needs_typestr(t) {
		return func(s *State) {
			ptr := s.P

			// Prepend object brace.
			s.B = append(s.B, '{')

			for i := 0; i < len(fields); i++ {
				// Get struct field ptr via offset.
				s.P = add(ptr, fields[i].offset)

				// Append field name and value separator.
				s.B = append(s.B, fields[i].name+"="...)

				// Format i'th field.
				fields[i].format(s)
				s.B = append(s.B, ',', ' ')
			}

			// Drop final ", ".
			s.B = s.B[:len(s.B)-2]

			// Append object brace.
			s.B = append(s.B, '}')
		}
	}

	// Struct type string with refs.
	typestr := typestr_with_refs(t)

	return func(s *State) {
		ptr := s.P

		// Include type info.
		if s.A.WithType() {
			s.B = append(s.B, typestr...)
		}

		// Prepend object brace.
		s.B = append(s.B, '{')

		for i := 0; i < len(fields); i++ {
			// Get struct field ptr via offset.
			s.P = add(ptr, fields[i].offset)

			// Append field name and value separator.
			s.B = append(s.B, fields[i].name+"="...)

			// Format i'th field.
			fields[i].format(s)
			s.B = append(s.B, ',', ' ')
		}

		// Drop final ", ".
		s.B = s.B[:len(s.B)-2]

		// Append object brace.
		s.B = append(s.B, '}')
	}
}
