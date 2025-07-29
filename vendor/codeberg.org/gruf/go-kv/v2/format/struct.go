package format

// field stores the minimum necessary
// data for iterating and formatting
// each field in a given struct.
type field struct {
	format FormatFunc
	name   string
	offset uintptr
}

// iterStructType returns a FormatFunc capable of iterating
// and formatting the given struct type currently in typenode{}.
// note this will fetch sub-FormatFuncs for each struct field.
func (fmt *Formatter) iterStructType(t typenode) FormatFunc {
	// Number of struct fields.
	n := t.rtype.NumField()

	// Gather format functions.
	fields := make([]field, n)
	for i := 0; i < n; i++ {

		// Get struct field at index.
		sfield := t.rtype.Field(i)
		rtype := sfield.Type

		// Get nested field typenode with appropriate flags.
		flags := reflect_struct_field_flags(t.flags, rtype)
		ft := t.next(sfield.Type, flags)

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

func emptyStructType(t typenode) FormatFunc {
	if !t.needs_typestr() {
		return func(s *State) {
			// Append empty object.
			s.B = append(s.B, "{}"...)
		}
	}

	// Struct type string with refs.
	typestr := t.typestr_with_refs()

	// Append empty object
	// with type information.
	return func(s *State) {
		if s.A.WithType() {
			s.B = append(s.B, typestr...)
		}
		s.B = append(s.B, "{}"...)
	}
}

func iterSingleFieldStructType(t typenode, field field) FormatFunc {
	if field.format == nil {
		panic("nil func")
	}

	if !t.needs_typestr() {
		return func(s *State) {
			// Wrap 'fn' with braces + field name.
			s.B = append(s.B, "{"+field.name+"="...)
			field.format(s)
			s.B = append(s.B, "}"...)
		}
	}

	// Struct type string with refs.
	typestr := t.typestr_with_refs()

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

func iterMultiFieldStructType(t typenode, fields []field) FormatFunc {
	for _, field := range fields {
		if field.format == nil {
			panic("nil func")
		}
	}

	if !t.needs_typestr() {
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
	typestr := t.typestr_with_refs()

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
