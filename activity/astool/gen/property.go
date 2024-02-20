package gen

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

const (
	// Method names for generated code
	getMethod                 = "Get"
	setMethod                 = "Set"
	hasAnyMethod              = "HasAny"
	clearMethod               = "Clear"
	iteratorClearMethod       = "clear"
	isMethod                  = "Is"
	atMethodName              = "At"
	isIRIMethod               = "IsIRI"
	getIRIMethod              = "GetIRI"
	setIRIMethod              = "SetIRI"
	appendMethod              = "Append"
	prependMethod             = "Prepend"
	insertMethod              = "Insert"
	removeMethod              = "Remove"
	lenMethod                 = "Len"
	swapMethod                = "Swap"
	lessMethod                = "Less"
	kindIndexMethod           = "KindIndex"
	serializeMethod           = "Serialize"
	deserializeMethod         = "Deserialize"
	nameMethod                = "Name"
	serializeIteratorMethod   = "serialize"
	deserializeIteratorMethod = "deserialize"
	hasLanguageMethod         = "HasLanguage"
	getLanguageMethod         = "GetLanguage"
	setLanguageMethod         = "SetLanguage"
	nextMethod                = "Next"
	prevMethod                = "Prev"
	beginMethod               = "Begin"
	endMethod                 = "End"
	emptyMethod               = "Empty"
	// Context string management
	contextMethod = "JSONLDContext"
	// Member names for generated code
	unknownMemberName = "unknown"
	// Reference to the rdf:langString member! Kludge: both of these must be
	// kept in sync with the generated code.
	langMapMember       = "rdfLangStringMember"
	isLanguageMapMethod = "IsRDFLangString"
	// Kind Index constants
	iriKindIndex           = -2
	noneOrUnknownKindIndex = -1
	// iterator specific
	myIndexMemberName = "myIdx"
	parentMemberName  = "parent"
)

// join appends a bunch of Go Code together, each on their own line.
func join(s []jen.Code) *jen.Statement {
	r := jen.Empty()
	for i, stmt := range s {
		if i > 0 {
			r.Line()
		}
		r.Add(stmt)
	}
	return r
}

// Identifier determines how a name will appear in documentation and Go code.
type Identifier struct {
	// LowerName is the typical name used in documentation.
	LowerName string
	// CamelName is the typical name used in identifiers in code.
	CamelName string
}

// Kind is data that describes a concrete Go type, how to serialize and
// deserialize such types, compare the types, and other meta-information to use
// during Go code generation.
//
// Only represents values and other types.
type Kind struct {
	Name  Identifier
	Vocab string
	// ConcreteKind is expected to be properly qualified.
	ConcreteKind *jen.Statement
	Nilable      bool
	IsURI        bool

	// TODO: Untangle the package management mess so that the below do not
	// need to be duplicated.

	// These <FuncName>Fn types are for qualified names of the functions.
	// Expected to always be non-nil: a function is needed to deserialize.
	DeserializeFn *jen.Statement
	// If any of these are nil at generation time, assume to call the method
	// on the object directly (instead of a qualified function).
	SerializeFn *jen.Statement
	LessFn      *jen.Statement

	// The following are only used for values, not types, as actual implementations
	SerializeDef   *codegen.Function
	DeserializeDef *codegen.Function
	LessDef        *codegen.Function
}

// NewKindForValue creates a Kind for a value type.
func NewKindForValue(docName, idName, vocab string,
	defType *jen.Statement,
	isNilable, isURI bool,
	serializeFn, deserializeFn, lessFn *codegen.Function) *Kind {
	return &Kind{
		Name: Identifier{
			LowerName: docName,
			CamelName: idName,
		},
		Vocab:          vocab,
		ConcreteKind:   defType,
		Nilable:        isNilable,
		IsURI:          isURI,
		SerializeFn:    serializeFn.QualifiedName(),
		DeserializeFn:  deserializeFn.QualifiedName(),
		LessFn:         lessFn.QualifiedName(),
		SerializeDef:   serializeFn,
		DeserializeDef: deserializeFn,
		LessDef:        lessFn,
	}
}

// NewKindForType creates a Kind for an ActivitySteams type.
func NewKindForType(docName, idName, vocab string) *Kind {
	return &Kind{
		// Name must use toIdentifier for vocabValuePackage and
		// valuePackage to be the same.
		Name: Identifier{
			LowerName: docName,
			CamelName: idName,
		},
		Vocab:   vocab,
		Nilable: true,
		IsURI:   false,
		// Instead of populating:
		//   - ConcreteKind
		//   - DeserializeFn
		//   - SerializeFn (Not populated for types)
		//   - LessFn      (Not populated for types)
		//
		// The TypeGenerator is responsible for calling SetKindFns on
		// the properties, to property wire a Property's Kind back to
		// the Type's implementation.
	}
}

// lessFnCode creates the correct code calling this Kind's less function
// depending on whether the Kind is a value or a type.
func (k Kind) lessFnCode(this, other *jen.Statement) *jen.Statement {
	// LessFn is nil case -- call comparison Less method directly on the LHS
	lessCall := this.Clone().Dot(compareLessMethod).Call(other.Clone())
	if k.isValue() {
		// LessFn is indeed a function -- call this function
		lessCall = k.LessFn.Clone().Call(
			this.Clone(),
			other.Clone(),
		)
	}
	return lessCall
}

// lessFnCode creates the correct code calling this Kind's deserialize function
// depending on whether the Kind is a value or a type.
func (k Kind) deserializeFnCode(m, ctx *jen.Statement) *jen.Statement {
	if k.isValue() {
		return k.DeserializeFn.Clone().Call(m)
	} else {
		// If LessFn is nil, this means it is a type. Which requires an
		// additional Call and the context.
		return k.DeserializeFn.Clone().Call().Call(m, ctx)
	}
}

// isValue returns true if this Kind is a value, or false if it is a type.
func (k Kind) isValue() bool {
	// LessFn is not nil, this means it is a value.
	// If LessFn is nil, this means it is a type. Types will have their
	// LessThan method called directly on the type.
	return k.LessFn != nil
}

// PropertyGenerator is a common base struct used in both Functional and
// NonFunctional ActivityStreams properties. It provides common naming patterns,
// logic, and common Go code to be generated.
//
// It also properly handles the concept of generating Go code for property
// iterators, which are needed for NonFunctional properties.
type PropertyGenerator struct {
	vocabName             string
	vocabURI              *url.URL
	vocabAlias            string
	managerMethods        []*codegen.Method
	packageManager        *PackageManager
	name                  Identifier
	comment               string
	kinds                 []Kind
	hasNaturalLanguageMap bool
	asIterator            bool
}

// HasNaturalLanguageMap returns whether this property has a natural language
// map.
func (p *PropertyGenerator) HasNaturalLanguageMap() bool {
	return p.hasNaturalLanguageMap
}

// VocabName returns this property's vocabulary name.
func (p *PropertyGenerator) VocabName() string {
	return p.vocabName
}

// GetKinds gets this property's kinds.
func (p *PropertyGenerator) GetKinds() []Kind {
	return p.kinds
}

// GetPrivatePackage gets this property's private Package.
func (p *PropertyGenerator) GetPrivatePackage() Package {
	return p.packageManager.PrivatePackage()
}

// GetPublicPackage gets this property's public Package.
func (p *PropertyGenerator) GetPublicPackage() Package {
	return p.packageManager.PublicPackage()
}

// SetKindFns allows TypeGenerators to later notify this Property what functions
// to use when generating the serialization code.
//
// The name parameter must match the LowerName of an Identifier.
//
// This feels very hacky.
func (p *PropertyGenerator) SetKindFns(docName, idName, vocab string, qualKind *jen.Statement, deser *codegen.Method) error {
	for i, kind := range p.kinds {
		if kind.Name.LowerName == docName && kind.Vocab == vocab {
			if kind.SerializeFn != nil || kind.DeserializeFn != nil || kind.LessFn != nil {
				return fmt.Errorf("property kind already has serialization functions set for %q: %s", docName, p.PropertyName())
			}
			kind.ConcreteKind = qualKind
			kind.DeserializeFn = deser.On(managerInitName())
			p.managerMethods = append(p.managerMethods, deser)
			p.kinds[i] = kind
			return nil
		}
	}
	// In the case of extended types applying themselves to their parents'
	// range, they will be missing from the property's kinds list. Append a
	// new kind to handle this use case.
	k := NewKindForType(docName, idName, vocab)
	k.ConcreteKind = qualKind
	k.DeserializeFn = deser.On(managerInitName())
	p.managerMethods = append(p.managerMethods, deser)
	p.kinds = append(p.kinds, *k)
	return nil
}

// getAllManagerMethods returns the list of manager methods used by this
// property.
func (p *PropertyGenerator) getAllManagerMethods() []*codegen.Method {
	return p.managerMethods
}

// StructName returns the name of the type, which may or may not be a struct,
// to generate.
func (p *PropertyGenerator) StructName() string {
	if p.asIterator {
		return p.name.CamelName
	}
	return fmt.Sprintf("%s%sProperty", p.VocabName(), p.name.CamelName)
}

// iteratorTypeName determines the identifier to use for the iterator type.
func (p *PropertyGenerator) iteratorTypeName() Identifier {
	s := fmt.Sprintf("%s%s", p.VocabName(), p.name.CamelName)
	return Identifier{
		LowerName: s,
		CamelName: fmt.Sprintf("%sPropertyIterator", s),
	}
}

// InterfaceName returns the interface name of the property type.
func (p *PropertyGenerator) InterfaceName() string {
	return p.StructName()
}

// parentTypeInterfaceName is useful for iterators that need the base property
// type's interface name.
func (p *PropertyGenerator) parentTypeInterfaceName() string {
	return strings.TrimSuffix(p.StructName(), "Iterator")
}

// PropertyName returns the name of this property, as defined in
// specifications. It is not suitable for use in generated code function
// identifiers.
func (p *PropertyGenerator) PropertyName() string {
	return p.name.LowerName
}

// Comments returns the comment for this property.
func (p *PropertyGenerator) Comments() string {
	return p.comment
}

// DeserializeFnName returns the identifier of the function that deserializes
// raw JSON into the generated Go type.
func (p *PropertyGenerator) DeserializeFnName() string {
	if p.asIterator {
		return fmt.Sprintf("%s%s", deserializeIteratorMethod, p.name.CamelName)
	}
	return fmt.Sprintf("%s%sProperty", deserializeMethod, p.name.CamelName)
}

// getFnName returns the identifier of the function that fetches concrete types
// of the property.
func (p *PropertyGenerator) getFnName(i int) string {
	if len(p.kinds) == 1 {
		return getMethod
	}
	return fmt.Sprintf("%s%s%s", getMethod, p.kinds[i].Vocab, p.kindCamelName(i))
}

// setFnName returns the identifier of the function that sets concrete types
// of the property.
func (p *PropertyGenerator) setFnName(i int) string {
	if len(p.kinds) == 1 {
		return setMethod
	}
	return fmt.Sprintf("%s%s%s", setMethod, p.kinds[i].Vocab, p.kindCamelName(i))
}

// serializeFnName returns the identifier of the function that serializes the
// generated Go type into raw JSON.
func (p *PropertyGenerator) serializeFnName() string {
	if p.asIterator {
		return serializeIteratorMethod
	}
	return serializeMethod
}

// kindCamelName returns an identifier-friendly name for the kind at the
// specified index.
//
// It will panic if 'i' is out of range.
func (p *PropertyGenerator) kindCamelName(i int) string {
	return p.kinds[i].Name.CamelName
}

// memberName returns the identifier to use for the kind at the specified index.
//
// It will panic if 'i' is out of range.
func (p *PropertyGenerator) memberName(i int) string {
	k := p.kinds[i]
	v := strings.ToLower(k.Vocab)
	return fmt.Sprintf("%s%sMember", v, k.Name.CamelName)
}

// hasMemberName returns the identifier to use for struct members that determine
// whether non-nilable types have been set. Panics if called for a Kind that is
// nilable.
func (p *PropertyGenerator) hasMemberName(i int) string {
	if len(p.kinds) == 1 && p.kinds[0].Nilable {
		panic("PropertyGenerator.hasMemberName called for nilable single value")
	}
	return fmt.Sprintf("has%sMember", p.kinds[i].Name.CamelName)
}

// clearMethodName returns the identifier to use for methods that clear all
// values from the property.
func (p *PropertyGenerator) clearMethodName() string {
	if p.asIterator {
		return iteratorClearMethod
	}
	return clearMethod
}

// commonMethods returns methods common to every property.
func (p *PropertyGenerator) commonMethods() (m []*codegen.Method) {
	if p.asIterator {
		// Next & Prev methods
		m = append(m, codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			nextMethod,
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), p.InterfaceName())},
			[]jen.Code{
				jen.If(
					jen.Id(codegen.This()).Dot(myIndexMemberName).Op("+").Lit(1).Op(">=").Id(codegen.This()).Dot(parentMemberName).Dot(lenMethod).Call(),
				).Block(
					jen.Return(jen.Nil()),
				).Else().Block(
					jen.Return(
						jen.Id(codegen.This()).Dot(parentMemberName).Dot(atMethodName).Call(jen.Id(codegen.This()).Dot(myIndexMemberName).Op("+").Lit(1)),
					),
				),
			},
			fmt.Sprintf("%s returns the next iterator, or nil if there is no next iterator.", nextMethod)))
		m = append(m, codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			prevMethod,
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), p.InterfaceName())},
			[]jen.Code{
				jen.If(
					jen.Id(codegen.This()).Dot(myIndexMemberName).Op("-").Lit(1).Op("<").Lit(0),
				).Block(
					jen.Return(jen.Nil()),
				).Else().Block(
					jen.Return(
						jen.Id(codegen.This()).Dot(parentMemberName).Dot(atMethodName).Call(jen.Id(codegen.This()).Dot(myIndexMemberName).Op("-").Lit(1)),
					),
				),
			},
			fmt.Sprintf("%s returns the previous iterator, or nil if there is no previous iterator.", prevMethod)))
	}
	return m
}

// isMethodName returns the identifier to use for methods that determine if a
// property holds a specific Kind of value.
func (p *PropertyGenerator) isMethodName(i int) string {
	return fmt.Sprintf("%s%s%s", isMethod, p.kinds[i].Vocab, p.kindCamelName(i))
}

// ConstructorFn creates a constructor function with a default vocabulary
// alias.
func (p *PropertyGenerator) ConstructorFn() *codegen.Function {
	return codegen.NewCommentedFunction(
		p.GetPrivatePackage().Path(),
		fmt.Sprintf("%s%s", constructorName, p.StructName()),
		/*params=*/ nil,
		[]jen.Code{
			jen.Op("*").Qual(p.GetPrivatePackage().Path(), p.StructName()),
		},
		[]jen.Code{
			jen.Return(
				jen.Op("&").Qual(p.GetPrivatePackage().Path(), p.StructName()).Values(
					jen.Dict{
						jen.Id(aliasMember): jen.Lit(p.vocabAlias),
					},
				),
			),
		},
		fmt.Sprintf("%s%s creates a new %s property.", constructorName, p.StructName(), p.PropertyName()))
}

// hasURIKind returns true if this property already has a Kind that is a URI.
func (p *PropertyGenerator) hasURIKind() bool {
	for _, k := range p.kinds {
		if k.IsURI {
			return true
		}
	}
	return false
}

// hasTypeKind returns true if this property has a Kind that is a type.
func (p *PropertyGenerator) hasTypeKind() bool {
	for _, k := range p.kinds {
		if !k.isValue() {
			return true
		}
	}
	return false
}
