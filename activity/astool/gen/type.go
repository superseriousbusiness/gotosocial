package gen

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

const (
	typeInterfaceName          = "Type"
	typePropertyConstructor    = "typePropertyConstructor"
	jsonLDContextInterfaceName = "jsonldContexter"
	extendedByMethod           = "IsExtendedBy"
	extendingMethod            = "IsExtending"
	extendsMethod              = "Extends"
	isAMethod                  = "IsOrExtends"
	disjointWithMethod         = "IsDisjointWith"
	typeNameMethod             = "GetTypeName"
	vocabURIMethod             = "VocabularyURI"
	serializeMethodName        = "Serialize"
	deserializeFnName          = "Deserialize"
	compareLessMethod          = "LessThan"
	getUnknownMethod           = "GetUnknownProperties"
	unknownMember              = "unknown"
	aliasMember                = "alias"
	getMethodFormat            = "Get%s"
	constructorName            = "New"
)

const (
	// The following are all kluges to refer to specific properties: the
	// 'type' and 'id' property members, types, and functions!
	//
	// TODO: Figure out how to obtain these names at code-generation
	// runtime.
	typeMember    = JSONLDVocabName + jsonLDTypeCamelName
	getIdFunction = getMethod + JSONLDVocabName + jsonLDIdCamelName
	setIdFunction = setMethod + JSONLDVocabName + jsonLDIdCamelName
	idType        = JSONLDVocabName + jsonLDIdCamelName + "Property"
)

// typePropertyConstructorName returns the package variable name for the
// constructor of a Type property.
func typePropertyConstructorName() string {
	return typePropertyConstructor
}

// TypeInterface returns the Type Interface that is needed for ActivityStream
// types to compile for methods dealing with extending, in the inheritance
// sense.
func TypeInterface(pkg Package) *codegen.Interface {
	comment := fmt.Sprintf("%s represents an ActivityStreams type.", typeInterfaceName)
	funcs := []codegen.FunctionSignature{
		{
			Name:    typeNameMethod,
			Params:  nil,
			Ret:     []jen.Code{jen.String()},
			Comment: fmt.Sprintf("%s returns the ActivityStreams type name.", typeNameMethod),
		},
		{
			Name:    vocabURIMethod,
			Params:  nil,
			Ret:     []jen.Code{jen.String()},
			Comment: fmt.Sprintf("%s returns the vocabulary's URI as a string.", vocabURIMethod),
		},
		{
			Name:    getIdFunction,
			Params:  nil,
			Ret:     []jen.Code{jen.Qual(pkg.Path(), idType)},
			Comment: fmt.Sprintf("%s returns the \"id\" property if it exists, and nil otherwise.", getIdFunction),
		},
		{
			Name:    setIdFunction,
			Params:  []jen.Code{jen.Qual(pkg.Path(), idType)},
			Ret:     nil,
			Comment: fmt.Sprintf("%s sets the \"id\" property.", setIdFunction),
		},
		{
			Name:    contextMethod,
			Params:  nil,
			Ret:     []jen.Code{jen.Map(jen.String()).String()},
			Comment: fmt.Sprintf("%s returns the JSONLD URIs required in the context string for this property and the specific values that are set. The value in the map is the alias used to import the property's value or values.", contextMethod),
		},
		{
			Name:    serializeMethodName,
			Params:  nil,
			Ret:     []jen.Code{jen.Map(jen.String()).Interface(), jen.Error()},
			Comment: fmt.Sprintf("%s converts this into an interface representation suitable for marshalling into a text or binary format.", serializeMethodName),
		},
	}
	return codegen.NewInterface(pkg.Path(), typeInterfaceName, funcs, comment)
}

// ContextInterface returns a jsonldContexter interface that is needed for
// ActivityStream types to recursively determine what context strings need to
// exist in a JSON-LD @context value for linked-data peers to parse.
//
// It is a private interface to make the implementation easier, not needed by
// anything outside the package this implementation is in.
func ContextInterface(pkg Package) *codegen.Interface {
	comment := fmt.Sprintf("%s is a private interface to determine the JSON-LD contexts and aliases needed for functional and non-functional properties. It is a helper interface for this implementation.", jsonLDContextInterfaceName)
	funcs := []codegen.FunctionSignature{
		{
			Name:    contextMethod,
			Params:  nil,
			Ret:     []jen.Code{jen.Map(jen.String()).String()},
			Comment: fmt.Sprintf("%s returns the JSONLD URIs required in the context string for this property and the specific values that are set. The value in the map is the alias used to import the property's value or values.", contextMethod),
		},
	}
	return codegen.NewInterface(
		pkg.Path(),
		jsonLDContextInterfaceName,
		funcs,
		comment)
}

// Property represents a property of an ActivityStreams type.
type Property interface {
	VocabName() string
	GetPublicPackage() Package
	PropertyName() string
	StructName() string
	InterfaceName() string
	SetKindFns(docName, idName, vocab string, kind *jen.Statement, deser *codegen.Method) error
	DeserializeFnName() string
	HasNaturalLanguageMap() bool
}

// TypeGenerator represents an ActivityStream type definition to generate in Go.
type TypeGenerator struct {
	vocabName         string
	vocabURI          *url.URL
	vocabAlias        string
	pm                *PackageManager
	typeName          string
	comment           string
	properties        map[string]Property
	withoutProperties map[string]Property
	rangeProperties   []Property
	extends           []*TypeGenerator
	disjoint          []*TypeGenerator
	typeless          bool
	extendedBy        []*TypeGenerator
	m                 *ManagerGenerator
	cacheOnce         sync.Once
	cachedStruct      *codegen.Struct
}

// NewTypeGenerator creates a new generator for a specific ActivityStreams Core
// or extension type. It will return an error if there are multiple properties
// have the same Name.
//
// The TypeGenerator should be in the second pass to construct, relying on the
// fact that properties have already been constructed.
//
// The extends and disjoint parameters are allowed to be nil. These lists must
// also have unique (non-duplicated) elements. Note that the disjoint entries
// will be set up bi-directionally properly; no need to go back to an existing
// TypeGenerator to set up the link correctly.
//
// The rangeProperties list is allowed to be nil. Any passed in will properly
// have their SetKindFns bookkeeping done.
//
// All TypeGenerators must be created before the Definition method is called, to
// ensure that type extension, in the inheritence sense, is properly set up.
//
// A ManagerGenerator must be created with this type before Definition is
// called, to ensure that the serialization functions are properly set up.
func NewTypeGenerator(vocabName string,
	vocabURI *url.URL,
	vocabAlias string,
	pm *PackageManager,
	typeName, comment string,
	properties, withoutProperties, rangeProperties []Property,
	extends, disjoint []*TypeGenerator,
	typeless bool) (*TypeGenerator, error) {
	t := &TypeGenerator{
		vocabName:         vocabName,
		vocabURI:          vocabURI,
		vocabAlias:        vocabAlias,
		pm:                pm,
		typeName:          typeName,
		comment:           comment,
		properties:        make(map[string]Property, len(properties)),
		withoutProperties: make(map[string]Property, len(withoutProperties)),
		rangeProperties:   rangeProperties,
		extends:           extends,
		disjoint:          disjoint,
		typeless:          typeless,
	}
	for _, property := range properties {
		if err := t.AddPropertyGenerator(property); err != nil {
			return nil, err
		}
	}
	for _, wop := range withoutProperties {
		if _, has := t.withoutProperties[wop.StructName()]; has {
			return nil, fmt.Errorf("type already has withoutproperty with name %q", wop.StructName())
		}
		t.withoutProperties[wop.StructName()] = wop
	}
	// Complete doubly-linked extends/extendedBy lists.
	for _, ext := range extends {
		ext.extendedBy = append(ext.extendedBy, t)
	}
	// Complete doubly-linked disjoint types.
	for _, disj := range disjoint {
		disj.disjoint = append(disj.disjoint, t)
	}
	return t, nil
}

// AddPropertyGenerator adds a property generator to this type. It must be
// called before Definition is called.
func (t *TypeGenerator) AddPropertyGenerator(property Property) error {
	if _, has := t.properties[property.StructName()]; has {
		return fmt.Errorf("type already has property with name %q", property.StructName())
	}
	t.properties[property.StructName()] = property
	return nil
}

// AddRangeProperty adds another property as having this type as a value. Must
// be called before Definition is called.
func (t *TypeGenerator) AddRangeProperty(property Property) {
	t.rangeProperties = append(t.rangeProperties, property)
}

// apply propagates the manager's functions referring to this type's
// implementation as if this type were a Kind.
//
// Prepares to use the manager for the Definition generation.
//
// Must be called before Definition is called on the properties.
func (t *TypeGenerator) apply(m *ManagerGenerator) error {
	t.m = m
	// Set up Kind functions for this type, on its range of properties as
	// well as the range of properties of those it is extending from.
	deser := m.getDeserializationMethodForType(t)
	kind := jen.Qual(t.PublicPackage().Path(), t.InterfaceName())
	// Refursively-applying function.
	var setKindsOnWhoseProps func(whichType *TypeGenerator) error
	// Map to ensure we only call each property once.
	propsSet := make(map[Property]bool)
	setKindsOnWhoseProps = func(whichType *TypeGenerator) error {
		// Apply this TypeGenerator's kinds to whichType's range of
		// properties.
		for _, p := range whichType.rangeProperties {
			if propsSet[p] {
				continue
			}
			// Kluge: convert.toIdentifier must match this!
			if e := p.SetKindFns(t.TypeName(), strings.Title(t.TypeName()), t.vocabName, kind, deser); e != nil {
				return e
			}
			propsSet[p] = true
		}
		// Recursively apply this TypeGenerator's kinds to the parents
		// of whichType.
		for _, extendParent := range whichType.extends {
			if e := setKindsOnWhoseProps(extendParent); e != nil {
				return e
			}
		}
		return nil
	}
	// Begin the recursing by applying to this type's range of properties.
	return setKindsOnWhoseProps(t)
}

// VocabName returns this TypeGenerator's vocabulary name.
func (t *TypeGenerator) VocabName() string {
	return t.vocabName
}

// Package gets this TypeGenerator's Private Package.
func (t *TypeGenerator) PrivatePackage() Package {
	return t.pm.PrivatePackage()
}

// Package gets this TypeGenerator's Public Package.
func (t *TypeGenerator) PublicPackage() Package {
	return t.pm.PublicPackage()
}

// Comment returns the comment for this type.
func (t *TypeGenerator) Comments() string {
	return t.comment
}

// TypeName returns the ActivityStreams name for this type.
func (t *TypeGenerator) TypeName() string {
	return t.typeName
}

// StructName returns the Go name for this type.
func (t *TypeGenerator) StructName() string {
	return fmt.Sprintf("%s%s", t.VocabName(), t.typeName)
}

// InterfaceName returns the interface name for this type.
func (t *TypeGenerator) InterfaceName() string {
	return t.StructName()
}

// Extends returns the generators of types that this ActivityStreams type
// extends from.
func (t *TypeGenerator) Extends() []*TypeGenerator {
	return t.extends
}

// ExtendedBy returns the generators of types that extend from this
// ActivityStreams type.
func (t *TypeGenerator) ExtendedBy() []*TypeGenerator {
	return t.extendedBy
}

// Disjoint returns the generators of types that this ActivityStreams type is
// disjoint to.
func (t *TypeGenerator) Disjoint() []*TypeGenerator {
	return t.disjoint
}

// Properties returns the Properties of this type, mapped by their property
// name.
func (t *TypeGenerator) Properties() map[string]Property {
	return t.properties
}

// WithoutProperties returns the properties that do not apply to this type,
// mapped by their property name.
func (t *TypeGenerator) WithoutProperties() map[string]Property {
	return t.withoutProperties
}

// extendsFnName determines the name of the Extends function, which
// determines if this ActivityStreams type extends another one.
func (t *TypeGenerator) extendsFnName() string {
	return fmt.Sprintf("%s%s", t.StructName(), extendsMethod)
}

// extendedByFnName determines the name of the ExtendedBy function, which
// determines if another ActivityStreams type extends this one.
func (t *TypeGenerator) extendedByFnName() string {
	return fmt.Sprintf("%s%s", t.TypeName(), extendedByMethod)
}

// isATypeFnName determines the name of the IsA function, which determines if
// this Type is the same as the other one or if another ActivityStreams type
// extends this one.
func (t *TypeGenerator) isATypeFnName() string {
	return fmt.Sprintf("%s%s", isAMethod, t.TypeName())
}

// disjointWithFnName determines the name of the DisjointWith function, which
// determines if another ActivityStreams type is disjoint with this one.
func (t *TypeGenerator) disjointWithFnName() string {
	return fmt.Sprintf("%s%s", t.TypeName(), disjointWithMethod)
}

// deserializationFnName determines the name of the deserialize function for
// this type.
func (t *TypeGenerator) deserializationFnName() string {
	return fmt.Sprintf("%s%s", deserializeFnName, t.TypeName())
}

// InterfaceDefinition creates the interface of this type in the specified
// package.
//
// Requires ManagerGenerator to have been created.
func (t *TypeGenerator) InterfaceDefinition(pkg Package) *codegen.Interface {
	s := t.Definition()
	return s.ToInterface(pkg.Path(), t.InterfaceName(), t.Comments())
}

// Definition generates the golang code for this ActivityStreams type.
func (t *TypeGenerator) Definition() *codegen.Struct {
	t.cacheOnce.Do(func() {
		members := t.members()
		ser := t.serializationMethod()
		less := t.lessMethod()
		get := t.getUnknownMethod()
		deser := t.deserializationFn()
		extendsFn, extendsMethod := t.extendsDefinition()
		getters := t.allGetters()
		setters := t.allSetters()
		constructor := t.constructorFn()
		ctxMethods := t.contextMethods()
		t.cachedStruct = codegen.NewStruct(
			t.Comments(),
			t.StructName(),
			append(append(append(
				[]*codegen.Method{
					t.nameDefinition(),
					t.vocabURIDefinition(),
					extendsMethod,
					ser,
					less,
					get,
				},
				ctxMethods...),
				getters...),
				setters...,
			),
			[]*codegen.Function{
				constructor,
				t.isATypeDefinition(),
				t.extendedByDefinition(),
				extendsFn,
				t.disjointWithDefinition(),
				deser,
			},
			members)
	})
	return t.cachedStruct
}

// sortedProperty is a slice of Properties that implements the Sort interface.
type sortedProperty []Property

// Less compares the property names.
func (s sortedProperty) Less(i, j int) bool {
	return s[i].PropertyName() < s[j].PropertyName()
}

// Swap reorders two elements.
func (s sortedProperty) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Len returns the length of this slice of properties.
func (s sortedProperty) Len() int {
	return len(s)
}

// allProperties returns all properties that this type contains, accounting for
// the extended types and without properties.
func (t *TypeGenerator) allProperties() []Property {
	p := t.properties
	// Properties of parents that are extended, minus DoesNotApplyTo
	var extends map[*TypeGenerator]string
	extends = t.getAllParentExtends(extends, t)
	for ext := range extends {
		for k, v := range ext.Properties() {
			p[k] = v
		}
	}
	for ext := range extends {
		for k := range ext.WithoutProperties() {
			delete(p, k)
		}
	}
	for k := range t.WithoutProperties() {
		delete(p, k)
	}
	// Sort the properties into a stable order -- this is important for
	// stability in comparisons such as LessThan in order to be able to
	// easily explain the "in property lexicographical" order.
	//
	// Also improves readability in the implementation code.
	sortedP := make(sortedProperty, 0, len(p))
	for _, property := range p {
		sortedP = append(sortedP, property)
	}
	sort.Sort(sortedP)
	return sortedP
}

// memberName returns the member name for this property.
func (*TypeGenerator) memberName(p Property) string {
	return fmt.Sprintf(
		"%s%s",
		p.VocabName(),
		strings.Title(p.PropertyName()))
}

// members returns all the properties this type has as its members.
func (t *TypeGenerator) members() (members []jen.Code) {
	p := t.allProperties()
	// Convert to jen.Code
	members = make([]jen.Code, 0, len(p))
	for _, property := range p {
		members = append(members, jen.Id(t.memberName(property)).Qual(property.GetPublicPackage().Path(), property.InterfaceName()))
	}
	// TODO: Normalize alias of properties when setting properties.
	members = append(members, jen.Id(aliasMember).String())
	members = append(members, jen.Id(unknownMember).Map(jen.String()).Interface())
	return
}

// nameDefinition generates the golang method for returning the ActivityStreams
// type name.
func (t *TypeGenerator) nameDefinition() *codegen.Method {
	return codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		typeNameMethod,
		t.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.String()},
		[]jen.Code{
			jen.Return(jen.Lit(t.TypeName())),
		},
		fmt.Sprintf("%s returns the name of this type.", typeNameMethod))
}

// vocabURIDefinition generates the golang method for returning this type's
// vocabulary URI as a string.
func (t *TypeGenerator) vocabURIDefinition() *codegen.Method {
	return codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		vocabURIMethod,
		t.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.String()},
		[]jen.Code{
			jen.Return(jen.Lit(t.vocabURI.String())),
		},
		fmt.Sprintf("%s returns the vocabulary's URI as a string.", vocabURIMethod))

}

// getAllParentExtends recursively determines all the parent types that this
// type extends from.
func (t *TypeGenerator) getAllParentExtends(s map[*TypeGenerator]string, tg *TypeGenerator) map[*TypeGenerator]string {
	if s == nil {
		s = make(map[*TypeGenerator]string)
	}
	for _, e := range tg.Extends() {
		s[e] = e.TypeName()
		s = t.getAllParentExtends(s, e)
	}
	return s
}

// extendsDefinition generates the golang function for determining if this
// ActivityStreams type extends another type. It requires the Type interface.
func (t *TypeGenerator) extendsDefinition() (*codegen.Function, *codegen.Method) {
	var extends map[*TypeGenerator]string
	extends = t.getAllParentExtends(extends, t)
	// Sort the rootin' tootin' thing, eliminating noise upon regeneration..
	extendsStr := make([]string, 0, len(extends))
	for _, name := range extends {
		extendsStr = append(extendsStr, name)
	}
	sort.Strings(extendsStr)
	extensions := make([]jen.Code, 0, len(extendsStr))
	for _, name := range extendsStr {
		extensions = append(extensions, jen.Lit(name))
	}
	impl := []jen.Code{jen.Comment("Shortcut implementation: this does not extend anything."), jen.Return(jen.False())}
	if len(extensions) > 0 {
		impl = []jen.Code{jen.Id("extensions").Op(":=").Index().String().Values(extensions...),
			jen.For(jen.List(
				jen.Id("_"),
				jen.Id("ext"),
			).Op(":=").Range().Id("extensions")).Block(
				jen.If(
					jen.Id("ext").Op("==").Id("other").Dot(typeNameMethod).Call(),
				).Block(
					jen.Return(jen.True()),
				),
			),
			jen.Return(jen.False())}
	}
	f := codegen.NewCommentedFunction(
		t.PrivatePackage().Path(),
		t.extendsFnName(),
		[]jen.Code{jen.Id("other").Qual(t.PublicPackage().Path(), typeInterfaceName)},
		[]jen.Code{jen.Bool()},
		impl,
		fmt.Sprintf("%s returns true if the %s type extends from the other type.", t.extendsFnName(), t.TypeName()))
	m := codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		extendingMethod,
		t.StructName(),
		[]jen.Code{jen.Id("other").Qual(t.PublicPackage().Path(), typeInterfaceName)},
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			jen.Return(
				jen.Id(t.extendsFnName()).Call(jen.Id("other")),
			),
		},
		fmt.Sprintf("%s returns true if the %s type extends from the other type.", extendingMethod, t.TypeName()))
	return f, m
}

// getAllChildrenExtendBy recursivley determines all the child types that this
// type is extended by.
func (t *TypeGenerator) getAllChildrenExtendedBy(s []string, tg *TypeGenerator) []string {
	for _, e := range tg.ExtendedBy() {
		s = append(s, e.TypeName())
		s = t.getAllChildrenExtendedBy(s, e)
	}
	sort.Strings(s)
	return s
}

// isATypeDefinition generates the golang function for determining if another
// ActivityStreams type is this type or extends this type. It requires the Type
// interface.
func (t *TypeGenerator) isATypeDefinition() *codegen.Function {
	return codegen.NewCommentedFunction(
		t.PrivatePackage().Path(),
		t.isATypeFnName(),
		[]jen.Code{jen.Id("other").Qual(t.PublicPackage().Path(), typeInterfaceName)},
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			jen.If(
				jen.Id("other").Dot(typeNameMethod).Call().Op("==").Lit(t.TypeName()),
			).Block(
				jen.Return(jen.True()),
			),
			jen.Return(
				t.extendedByDefinition().Call(
					jen.Id("other"),
				),
			),
		},
		fmt.Sprintf("%s returns true if the other provided type is the %s type or extends from the %s type.", t.isATypeFnName(), t.TypeName(), t.TypeName()))
}

// extendedByDefinition generates the golang function for determining if
// another ActivityStreams type extends this type. It requires the Type
// interface.
func (t *TypeGenerator) extendedByDefinition() *codegen.Function {
	extendNames := t.getAllChildrenExtendedBy(nil, t)
	extensions := make([]jen.Code, len(extendNames))
	for i, e := range extendNames {
		extensions[i] = jen.Lit(e)
	}
	impl := []jen.Code{jen.Comment("Shortcut implementation: is not extended by anything."), jen.Return(jen.False())}
	if len(extensions) > 0 {
		impl = []jen.Code{jen.Id("extensions").Op(":=").Index().String().Values(extensions...),
			jen.For(jen.List(
				jen.Id("_"),
				jen.Id("ext"),
			).Op(":=").Range().Id("extensions")).Block(
				jen.If(
					jen.Id("ext").Op("==").Id("other").Dot(typeNameMethod).Call(),
				).Block(
					jen.Return(jen.True()),
				),
			),
			jen.Return(jen.False())}
	}
	return codegen.NewCommentedFunction(
		t.PrivatePackage().Path(),
		t.extendedByFnName(),
		[]jen.Code{jen.Id("other").Qual(t.PublicPackage().Path(), typeInterfaceName)},
		[]jen.Code{jen.Bool()},
		impl,
		fmt.Sprintf("%s returns true if the other provided type extends from the %s type. Note that it returns false if the types are the same; see the %q variant instead.", t.extendedByFnName(), t.TypeName(), t.isATypeFnName()))
}

// getAllChildrenDisjointWith recursivley determines all the child types that this
// type is disjoint with.
func (t *TypeGenerator) getAllDisjointWith() (s []string) {
	var extends map[*TypeGenerator]string
	extends = t.getAllParentExtends(extends, t)
	extends[t] = t.TypeName()
	for tg := range extends {
		for _, e := range tg.Disjoint() {
			s = append(s, e.TypeName())
			// Get all the disjoint type's children.
			s = t.getAllChildrenExtendedBy(s, e)
		}
	}
	return s
}

// disjointWithDefinition generates the golang function for determining if
// another ActivityStreams type is disjoint with this type. It requires the Type
// interface.
func (t *TypeGenerator) disjointWithDefinition() *codegen.Function {
	disjointNames := t.getAllDisjointWith()
	disjointWith := make([]jen.Code, len(disjointNames))
	for i, d := range disjointNames {
		disjointWith[i] = jen.Lit(d)
	}
	impl := []jen.Code{jen.Comment("Shortcut implementation: is not disjoint with anything."), jen.Return(jen.False())}
	if len(disjointWith) > 0 {
		impl = []jen.Code{jen.Id("disjointWith").Op(":=").Index().String().Values(disjointWith...),
			jen.For(jen.List(
				jen.Id("_"),
				jen.Id("disjoint"),
			).Op(":=").Range().Id("disjointWith")).Block(
				jen.If(
					jen.Id("disjoint").Op("==").Id("other").Dot(typeNameMethod).Call(),
				).Block(
					jen.Return(jen.True()),
				),
			),
			jen.Return(jen.False())}
	}
	return codegen.NewCommentedFunction(
		t.PrivatePackage().Path(),
		t.disjointWithFnName(),
		[]jen.Code{jen.Id("other").Qual(t.PublicPackage().Path(), typeInterfaceName)},
		[]jen.Code{jen.Bool()},
		impl,
		fmt.Sprintf("%s returns true if the other provided type is disjoint with the %s type.", t.disjointWithFnName(), t.TypeName()))
}

// serializationMethod returns the method needed to serialize a TypeGenerator as
// a property.
func (t *TypeGenerator) serializationMethod() (ser *codegen.Method) {
	serCode := jen.Commentf("Begin: Serialize known properties").Line()
	for _, prop := range t.allProperties() {
		serCode.Add(
			jen.Commentf("Maybe serialize property %q", prop.PropertyName()).Line(),
			jen.If(
				jen.Id(codegen.This()).Dot(t.memberName(prop)).Op("!=").Nil(),
			).Block(
				jen.If(
					jen.List(
						jen.Id("i"),
						jen.Err(),
					).Op(":=").Id(codegen.This()).Dot(t.memberName(prop)).Dot(serializeMethod).Call(),
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(jen.Nil(), jen.Err()),
				).Else().If(
					jen.Id("i").Op("!=").Nil(),
				).Block(
					jen.Id("m").Index(jen.Id(codegen.This()).Dot(t.memberName(prop)).Dot(nameMethod).Call()).Op("=").Id("i"),
				),
			).Line())
	}
	serCode = serCode.Commentf("End: Serialize known properties").Line()
	unknownCode := jen.Commentf("Begin: Serialize unknown properties").Line().For(
		jen.List(
			jen.Id("k"),
			jen.Id("v"),
		).Op(":=").Range().Id(codegen.This()).Dot(unknownMember),
	).Block(
		jen.Commentf("To be safe, ensure we aren't overwriting a known property"),
		jen.If(
			jen.List(
				jen.Id("_"),
				jen.Id("has"),
			).Op(":=").Id("m").Index(jen.Id("k")),
			jen.Op("!").Id("has"),
		).Block(
			jen.Id("m").Index(jen.Id("k")).Op("=").Id("v"),
		),
	).Line().Commentf("End: Serialize unknown properties").Line()
	header := jen.Id("m").Op(":=").Make(
		jen.Map(jen.String()).Interface(),
	)
	if !t.typeless {
		header = jen.Empty().Add(
			jen.Id("m").Op(":=").Make(
				jen.Map(jen.String()).Interface(),
			).Line(),
			jen.Id("typeName").Op(":=").Lit(t.TypeName()).Line(),
			jen.If(
				jen.Len(jen.Id(codegen.This()).Dot(aliasMember)).Op(">").Lit(0),
			).Block(
				jen.Id("typeName").Op("=").Id(codegen.This()).Dot(aliasMember).Op("+").Lit(":").Op("+").Lit(t.TypeName()),
			).Line(),
			jen.Id("m").Index(jen.Lit("type")).Op("=").Id("typeName"),
		)
	}
	ser = codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		serializeMethodName,
		t.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Map(jen.String()).Interface(), jen.Error()},
		[]jen.Code{
			header,
			serCode,
			unknownCode,
			jen.Return(jen.Id("m"), jen.Nil()),
		},
		fmt.Sprintf("%s converts this into an interface representation suitable for marshalling into a text or binary format.", serializeMethodName))
	return
}

// lessMethod returns the method needed to compare a type with another type.
func (t *TypeGenerator) lessMethod() (less *codegen.Method) {
	lessCode := jen.Commentf("Begin: Compare known properties").Line()
	for _, prop := range t.allProperties() {
		lessCode = lessCode.Add(
			jen.Commentf("Compare property %q", prop.PropertyName()).Line(),
			jen.If(
				jen.List(
					jen.Id("lhs"),
					jen.Id("rhs"),
				).Op(":=").List(
					jen.Id(codegen.This()).Dot(t.memberName(prop)),
					jen.Id("o").Dot(
						fmt.Sprintf(getMethodFormat, t.memberName(prop)),
					).Call(),
				),
				jen.Id("lhs").Op("!=").Nil().Op("&&").Id("rhs").Op("!=").Nil(),
			).Block(
				jen.If(
					jen.Id("lhs").Dot(compareLessMethod).Call(
						jen.Id("rhs"),
					),
				).Block(
					jen.Return(jen.True()),
				).Else().If(
					jen.Id("rhs").Dot(compareLessMethod).Call(
						jen.Id("lhs"),
					),
				).Block(
					jen.Return(jen.False()),
				),
			).Else().If(
				jen.Id("lhs").Op("==").Nil().Op("&&").Id("rhs").Op("!=").Nil(),
			).Block(
				jen.Commentf("Nil is less than anything else"),
				jen.Return(jen.True()),
			).Else().If(
				jen.Id("rhs").Op("!=").Nil().Op("&&").Id("rhs").Op("==").Nil(),
			).Block(
				jen.Commentf("Anything else is greater than nil"),
				jen.Return(jen.False()),
			),
			jen.Commentf("Else: Both are nil"),
			jen.Line())
	}
	lessCode = lessCode.Commentf("End: Compare known properties").Line()
	unknownCode := jen.Commentf("Begin: Compare unknown properties (only by number of them)").Line().If(
		jen.Len(
			jen.Id(codegen.This()).Dot(unknownMember),
		).Op("<").Len(
			jen.Id("o").Dot(getUnknownMethod).Call(),
		),
	).Block(
		jen.Return(jen.True()),
	).Else().If(
		jen.Len(
			jen.Id("o").Dot(getUnknownMethod).Call(),
		).Op("<").Len(
			jen.Id(codegen.This()).Dot(unknownMember),
		),
	).Block(
		jen.Return(jen.False()),
	).Commentf("End: Compare unknown properties (only by number of them)").Line()
	less = codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		compareLessMethod,
		t.StructName(),
		[]jen.Code{
			jen.Id("o").Qual(t.PublicPackage().Path(), t.InterfaceName()),
		},
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			lessCode,
			unknownCode,
			jen.Commentf("All properties are the same."),
			jen.Return(jen.False()),
		},
		fmt.Sprintf("%s computes if this %s is lesser, with an arbitrary but stable determination.", compareLessMethod, t.TypeName()))
	return
}

// deserializationFn returns free function reference that can be used to
// treat a TypeGenerator as another property's Kind.
func (t *TypeGenerator) deserializationFn() (deser *codegen.Function) {
	deserCode := jen.Commentf("Begin: Known property deserialization").Line()
	for _, prop := range t.allProperties() {
		deserMethod := t.m.getDeserializationMethodForProperty(prop)
		deserCode = deserCode.Add(
			jen.If(
				jen.List(
					jen.Id("p"),
					jen.Err(),
				).Op(":=").Add(deserMethod.On(managerInitName()).Call().Call(jen.Id("m"), jen.Id("aliasMap"))),
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Return(jen.Nil(), jen.Err()),
			).Else().If(
				jen.Id("p").Op("!=").Nil(),
			).Block(
				jen.Id(codegen.This()).Dot(t.memberName(prop)).Op("=").Id("p"),
			).Line())
	}
	deserCode = deserCode.Commentf("End: Known property deserialization").Line()
	knownProps := jen.Commentf("Begin: Code that ensures a property name is unknown").Line()
	for i, prop := range t.allProperties() {
		if i > 0 {
			knownProps = knownProps.Else()
		}
		knownProps = knownProps.If(
			jen.Id("k").Op("==").Lit(prop.PropertyName()),
		).Block(
			jen.Continue(),
		)
		if prop.HasNaturalLanguageMap() {
			knownProps = knownProps.Else().If(
				jen.Id("k").Op("==").Lit(prop.PropertyName() + "Map"),
			).Block(
				jen.Continue(),
			)
		}
	}
	knownProps = knownProps.Commentf("End: Code that ensures a property name is unknown").Line()
	unknownCode := jen.Commentf("Begin: Unknown deserialization").Line().For(
		jen.List(
			jen.Id("k"),
			jen.Id("v"),
		).Op(":=").Range().Id("m"),
	).Block(
		knownProps,
		jen.Id(codegen.This()).Dot(unknownMember).Index(jen.Id("k")).Op("=").Id("v"),
	).Line().Commentf("End: Unknown deserialization").Line()

	// Type vs typeless, typed needs an "aliasPrefix"
	header := jen.Empty().Add(
		jen.Id("alias").Op(":=").Lit("").Line(),
		jen.If(
			jen.List(
				jen.Id("a"),
				jen.Id("ok"),
			).Op(":=").Id("aliasMap").Index(jen.Lit(t.vocabURI.String())),
			jen.Id("ok"),
		).Block(
			jen.Id("alias").Op("=").Id("a"),
		).Line(),
		jen.Id(codegen.This()).Op(":=").Op("&").Id(t.StructName()).Values(jen.Dict{
			jen.Id(aliasMember):   jen.Id("alias"),
			jen.Id(unknownMember): jen.Make(jen.Map(jen.String()).Interface()),
		}),
	)
	typed := jen.Empty()
	if !t.typeless {
		header = jen.Empty().Add(
			jen.Id("alias").Op(":=").Lit("").Line(),
			jen.Id("aliasPrefix").Op(":=").Lit("").Line(),
			jen.If(
				jen.List(
					jen.Id("a"),
					jen.Id("ok"),
				).Op(":=").Id("aliasMap").Index(jen.Lit(t.vocabURI.String())),
				jen.Id("ok"),
			).Block(
				jen.Id("alias").Op("=").Id("a"),
				jen.Id("aliasPrefix").Op("=").Id("a").Op("+").Lit(":"),
			).Line(),
			jen.Id(codegen.This()).Op(":=").Op("&").Id(t.StructName()).Values(jen.Dict{
				jen.Id(aliasMember):   jen.Id("alias"),
				jen.Id(unknownMember): jen.Make(jen.Map(jen.String()).Interface()),
			}),
		)
		typed.Add(
			jen.If(
				jen.List(
					jen.Id("typeValue"),
					jen.Id("ok"),
				).Op(":=").Id("m").Index(jen.Lit("type")),
				jen.Op("!").Id("ok"),
			).Block(
				jen.Return(
					jen.Nil(),
					jen.Qual("fmt", "Errorf").Call(jen.Lit("no \"type\" property in map")),
				),
			).Else().If(
				jen.List(
					jen.Id("typeString"),
					jen.Id("ok"),
				).Op(":=").Id("typeValue").Assert(jen.String()),
				jen.Id("ok"),
			).Block(
				jen.Id("typeName").Op(":=").Qual("strings", "TrimPrefix").Call(
					jen.Id("typeString"),
					jen.Id("aliasPrefix"),
				),
				jen.If(
					jen.Id("typeName").Op("!=").Lit(t.TypeName()),
				).Block(
					jen.Return(
						jen.Nil(),
						jen.Qual("fmt", "Errorf").Call(jen.Lit("\"type\" property is not of %q type: %s"), jen.Lit(t.TypeName()), jen.Id("typeName")),
					),
				),
				jen.Commentf("Fall through, success in finding a proper Type"),
			).Else().If(
				jen.List(
					jen.Id("arrType"),
					jen.Id("ok"),
				).Op(":=").Id("typeValue").Assert(jen.Index().Interface()),
				jen.Id("ok"),
			).Block(
				jen.Id("found").Op(":=").False(),
				jen.For(
					jen.List(
						jen.Id("_"),
						jen.Id("elemVal"),
					).Op(":=").Range().Id("arrType"),
				).Block(
					jen.If(
						jen.List(
							jen.Id("typeString"),
							jen.Id("ok"),
						).Op(":=").Id("elemVal").Assert(jen.String()),
						jen.Id("ok").Op("&&").Qual("strings", "TrimPrefix").Call(
							jen.Id("typeString"),
							jen.Id("aliasPrefix"),
						).Op("==").Lit(t.TypeName()),
					).Block(
						jen.Id("found").Op("=").True(),
						jen.Break(),
					),
				),
				jen.If(
					jen.Op("!").Id("found"),
				).Block(
					jen.Return(
						jen.Nil(),
						jen.Qual("fmt", "Errorf").Call(jen.Lit("could not find a \"type\" property of value %q"), jen.Lit(t.TypeName())),
					),
				),
				jen.Commentf("Fall through, success in finding a proper Type"),
			).Else().Block(
				jen.Return(
					jen.Nil(),
					jen.Qual("fmt", "Errorf").Call(jen.Lit("\"type\" property is unrecognized type: %T"), jen.Id("typeValue")),
				),
			),
		)
	}
	deser = codegen.NewCommentedFunction(
		t.PrivatePackage().Path(),
		t.deserializationFnName(),
		[]jen.Code{jen.Id("m").Map(jen.String()).Interface(), jen.Id("aliasMap").Map(jen.String()).String()},
		[]jen.Code{jen.Op("*").Id(t.StructName()), jen.Error()},
		[]jen.Code{
			header,
			typed,
			deserCode,
			unknownCode,
			jen.Return(jen.Id(codegen.This()), jen.Nil()),
		},
		fmt.Sprintf("%s creates a %s from a map representation that has been unmarshalled from a text or binary format.", t.deserializationFnName(), t.TypeName()))
	return
}

// getUnknownMethod returns the GetUnknown helper used to compare which type is
// LessThan. This method is API-leaky and shouldn't be used by normal app
// developers.
func (t *TypeGenerator) getUnknownMethod() (get *codegen.Method) {
	get = codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		getUnknownMethod,
		t.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Map(jen.String()).Interface()},
		[]jen.Code{
			jen.Return(jen.Id(codegen.This()).Dot(unknownMember)),
		},
		fmt.Sprintf(
			"%s returns the unknown properties for the %s type. Note that this should not be used by app developers. It is only used to help determine which implementation is LessThan the other. Developers who are creating a different implementation of this type's interface can use this method in their LessThan implementation, but routine ActivityPub applications should not use this to bypass the code generation tool.",
			getUnknownMethod,
			t.TypeName()))
	return
}

// allGetters returns all property Getters for this type.
func (t *TypeGenerator) allGetters() (m []*codegen.Method) {
	for _, property := range t.allProperties() {
		m = append(m, codegen.NewCommentedValueMethod(
			t.PrivatePackage().Path(),
			fmt.Sprintf(getMethodFormat, t.memberName(property)),
			t.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Qual(property.GetPublicPackage().Path(), property.InterfaceName())},
			[]jen.Code{
				jen.Return(
					jen.Id(codegen.This()).Dot(t.memberName(property)),
				),
			},
			fmt.Sprintf(getMethodFormat+" returns the %q property if it exists, and nil otherwise.", t.memberName(property), property.PropertyName())))
	}
	return
}

// allSetters returns all property Setters for this type.
func (t *TypeGenerator) allSetters() (m []*codegen.Method) {
	for _, property := range t.allProperties() {
		m = append(m, codegen.NewCommentedPointerMethod(
			t.PrivatePackage().Path(),
			fmt.Sprintf("Set%s", t.memberName(property)),
			t.StructName(),
			[]jen.Code{jen.Id("i").Qual(property.GetPublicPackage().Path(), property.InterfaceName())},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(codegen.This()).Dot(t.memberName(property)).Op("=").Id("i"),
			},
			fmt.Sprintf("Set%s sets the %q property.", t.memberName(property), property.PropertyName())))
	}
	return
}

// getAllManagerMethods returns all the manager methods used by this type.
func (t *TypeGenerator) getAllManagerMethods() (m []*codegen.Method) {
	for _, prop := range t.allProperties() {
		deserMethod := t.m.getDeserializationMethodForProperty(prop)
		m = append(m, deserMethod)
	}
	return m
}

// constructorFn creates a constructor for this type.
func (t *TypeGenerator) constructorFn() *codegen.Function {
	typeName := jen.Lit(t.TypeName())
	if len(t.vocabAlias) > 0 {
		typeName = jen.Lit(t.vocabAlias).Op("+").Lit(":").Op("+").Lit(t.TypeName())
	}
	body := []jen.Code{
		jen.Return(
			jen.Op("&").Qual(t.PrivatePackage().Path(), t.StructName()).Values(
				jen.Dict{
					jen.Id(aliasMember):   jen.Lit(t.vocabAlias),
					jen.Id(unknownMember): jen.Make(jen.Map(jen.String()).Interface()),
				},
			),
		),
	}
	if !t.typeless {
		body = []jen.Code{
			jen.Id("typeProp").Op(":=").Id(typePropertyConstructorName()).Call(),
			jen.Id("typeProp").Dot("AppendXMLSchemaString").Call(typeName),
			jen.Return(
				jen.Op("&").Qual(t.PrivatePackage().Path(), t.StructName()).Values(
					jen.Dict{
						jen.Id(aliasMember):   jen.Lit(t.vocabAlias),
						jen.Id(unknownMember): jen.Make(jen.Map(jen.String()).Interface()),
						jen.Id(typeMember):    jen.Id("typeProp"),
					},
				),
			),
		}
	}
	return codegen.NewCommentedFunction(
		t.PrivatePackage().Path(),
		fmt.Sprintf("%s%s", constructorName, t.StructName()),
		/*params=*/ nil,
		[]jen.Code{
			jen.Op("*").Qual(t.PrivatePackage().Path(), t.StructName()),
		},
		body,
		fmt.Sprintf("%s%s creates a new %s type", constructorName, t.StructName(), t.TypeName()))
}

// contextMethod returns a map of the context's vocabulary
func (t *TypeGenerator) contextMethods() []*codegen.Method {
	helperName := fmt.Sprintf("helper%s", contextMethod)
	helper := codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		helperName,
		t.StructName(),
		[]jen.Code{jen.Id("i").Id(jsonLDContextInterfaceName), jen.Id("toMerge").Map(jen.String()).String()},
		[]jen.Code{jen.Map(jen.String()).String()},
		[]jen.Code{
			jen.If(
				jen.Id("i").Op("==").Nil(),
			).Block(
				jen.Return(jen.Id("toMerge")),
			),
			jen.For(
				jen.List(
					jen.Id("k"),
					jen.Id("v"),
				).Op(":=").Range().Id("i").Dot(contextMethod).Call(),
			).Block(
				jen.Commentf("Since the literal maps in this function are determined at\ncode-generation time, this loop should not overwrite an existing key with a\nnew value."),
				jen.Id("toMerge").Index(jen.Id("k")).Op("=").Id("v"),
			),
			jen.Return(jen.Id("toMerge")),
		},
		fmt.Sprintf("%s obtains the context uris and their aliases from a property, if it is not nil.", helperName))
	contextKind := jen.Id("m").Op(":=").Map(jen.String()).String().Values(
		jen.Dict{
			jen.Lit(t.vocabURI.String()): jen.Id(codegen.This()).Dot(aliasMember),
		},
	).Line()
	for _, property := range t.allProperties() {
		contextKind.Add(
			jen.Id("m").Op("=").Id(codegen.This()).Dot(helperName).Call(
				jen.Id(codegen.This()).Dot(t.memberName(property)),
				jen.Id("m")).Line())
	}
	ctxMethod := codegen.NewCommentedValueMethod(
		t.PrivatePackage().Path(),
		contextMethod,
		t.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Map(jen.String()).String()},
		[]jen.Code{
			contextKind,
			jen.Return(jen.Id("m")),
		},
		fmt.Sprintf("%s returns the JSONLD URIs required in the context string for this type and the specific properties that are set. The value in the map is the alias used to import the type and its properties.", contextMethod))
	return []*codegen.Method{helper, ctxMethod}
}
