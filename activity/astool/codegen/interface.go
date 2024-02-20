package codegen

import (
	"github.com/dave/jennifer/jen"
	"sort"
)

// sortedFunctionSignature sorts FunctionSignatures by name.
type sortedFunctionSignature []FunctionSignature

// Less compares Names.
func (s sortedFunctionSignature) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// Swap values.
func (s sortedFunctionSignature) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Len is the length of this slice.
func (s sortedFunctionSignature) Len() int {
	return len(s)
}

// FunctionSignature is an interface's function definition without
// an implementation.
type FunctionSignature struct {
	Name    string
	Params  []jen.Code
	Ret     []jen.Code
	Comment string
}

// Signature returns the uncommented raw signature.
func (f FunctionSignature) Signature() jen.Code {
	sig := jen.Func().Params(f.Params...)
	if len(f.Ret) > 0 {
		sig.Params(f.Ret...)
	}
	return sig
}

// Interface manages and generates a Golang interface definition.
type Interface struct {
	qual      *jen.Statement
	name      string
	functions []FunctionSignature
	comment   string
}

// NewInterface creates an Interface.
func NewInterface(pkg, name string,
	funcs []FunctionSignature,
	comment string) *Interface {
	i := &Interface{
		qual:      jen.Qual(pkg, name),
		name:      name,
		functions: funcs,
		comment:   comment,
	}
	sort.Sort(sortedFunctionSignature(i.functions))
	return i
}

// Definition produces the Golang code.
func (i Interface) Definition() jen.Code {
	stmts := jen.Empty()
	if len(i.comment) > 0 {
		stmts = jen.Comment(insertNewlines(i.comment)).Line()
	}
	defs := make([]jen.Code, 0, len(i.functions))
	for _, fn := range i.functions {
		def := jen.Empty()
		if len(fn.Comment) > 0 {
			def.Comment(insertNewlinesIndented(fn.Comment)).Line()
		}
		def.Id(fn.Name).Params(fn.Params...)
		if len(fn.Ret) > 0 {
			def.Params(fn.Ret...)
		}
		defs = append(defs, def)
	}
	return stmts.Type().Id(i.name).Interface(defs...)
}
