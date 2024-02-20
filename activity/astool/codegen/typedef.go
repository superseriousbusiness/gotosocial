package codegen

import (
	"github.com/dave/jennifer/jen"
	"sort"
	"unicode"
)

// Typedef defines a non-struct-based type, its functions, and its methods for
// Go code generation.
type Typedef struct {
	comment      string
	name         string
	concreteType jen.Code
	methods      map[string]*Method
	constructors map[string]*Function
}

// NewTypedef creates a new commented Typedef.
func NewTypedef(comment string,
	name string,
	concreteType jen.Code,
	methods []*Method,
	constructors []*Function) *Typedef {
	t := &Typedef{
		comment:      comment,
		name:         name,
		concreteType: concreteType,
		methods:      make(map[string]*Method, len(methods)),
		constructors: make(map[string]*Function, len(constructors)),
	}
	for _, m := range methods {
		t.methods[m.Name()] = m
	}
	for _, c := range constructors {
		t.constructors[c.Name()] = c
	}
	return t
}

// Definition generates the Go code required to define and implement this type,
// its methods, and its functions.
func (t *Typedef) Definition() jen.Code {
	def := jen.Empty()
	if len(t.comment) > 0 {
		def = jen.Commentf(insertNewlines(t.comment)).Line()
	}
	def = def.Type().Id(
		t.name,
	).Add(
		t.concreteType,
	)
	// Sort the functions and methods
	fs := make([]string, 0, len(t.constructors))
	for _, c := range t.constructors {
		fs = append(fs, c.Name())
	}
	ms := make([]string, 0, len(t.methods))
	for _, m := range t.methods {
		ms = append(ms, m.Name())
	}
	sort.Strings(fs)
	sort.Strings(ms)
	// Add the functions and methods in order
	for _, c := range fs {
		def = def.Line().Line().Add(t.constructors[c].Definition())
	}
	for _, m := range ms {
		def = def.Line().Line().Add(t.methods[m].Definition())
	}
	return def
}

// Method obtains the Go code to be generated for the method with a specific
// name. Panics if no such method exists.
func (t *Typedef) Method(name string) *Method {
	return t.methods[name]
}

// Constructors obtains the Go code to be generated for the function with a
// specific name. Panics if no such function exists.
func (t *Typedef) Constructors(name string) *Function {
	return t.constructors[name]
}

// ToInterface creates an interface version of this typedef.
func (t *Typedef) ToInterface(pkg, name, comment string) *Interface {
	fns := make([]FunctionSignature, 0, len(t.methods))
	for _, m := range t.methods {
		if unicode.IsUpper([]rune(m.Name())[0]) {
			fns = append(fns, m.ToFunctionSignature())
		}
	}
	return NewInterface(pkg, name, fns, comment)
}
