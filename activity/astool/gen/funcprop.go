package gen

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

const (
	iriMember = "iri"
)

// FunctionalPropertyGenerator produces Go code for properties that can have
// only one value. The resulting property is a struct type that can have one
// value that could be from multiple Kinds of values. If there is only one
// allowed Kind, then a smaller API is generated as a special case.
type FunctionalPropertyGenerator struct {
	PropertyGenerator
	cacheOnce    sync.Once
	cachedStruct *codegen.Struct
}

// NewFunctionalPropertyGenerator is a convenience constructor to create
// FunctionalPropertyGenerators.
//
// PropertyGenerators shoulf be in the first pass to construct, before types and
// other generators are constructed.
func NewFunctionalPropertyGenerator(vocabName string,
	vocabURI *url.URL,
	vocabAlias string,
	pm *PackageManager,
	name Identifier,
	comment string,
	kinds []Kind,
	hasNaturalLanguageMap bool) (*FunctionalPropertyGenerator, error) {
	// Ensure that the natural language map has the langString kind.
	if hasNaturalLanguageMap {
		found := false
		for _, k := range kinds {
			if k.Name.LowerName == "langString" {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Property has natural language map, but not an rdf:langString kind")
		}
	}
	return &FunctionalPropertyGenerator{
		PropertyGenerator: PropertyGenerator{
			vocabName:             vocabName,
			vocabURI:              vocabURI,
			vocabAlias:            vocabAlias,
			packageManager:        pm,
			hasNaturalLanguageMap: hasNaturalLanguageMap,
			name:                  name,
			comment:               comment,
			kinds:                 kinds,
		},
	}, nil
}

// InterfaceDefinition creates an interface definition in the provided package.
func (p *FunctionalPropertyGenerator) InterfaceDefinition(pkg Package) *codegen.Interface {
	s := p.Definition()
	return s.ToInterface(pkg.Path(), p.InterfaceName(), p.Comments())
}

// isSingleTypeDef determines whether a special-case API can be generated for
// one allowed Kind.
func (p *FunctionalPropertyGenerator) isSingleTypeDef() bool {
	return len(p.kinds) == 1
}

// Definition produces the Go Struct code definition, which can generate its Go
// implementations.
//
// The TypeGenerator apply must be called for all types before Definition is
// called.
func (p *FunctionalPropertyGenerator) Definition() *codegen.Struct {
	p.cacheOnce.Do(func() {
		if p.isSingleTypeDef() {
			p.cachedStruct = p.singleTypeDef()
		} else {
			p.cachedStruct = p.multiTypeDef()
		}
	})
	return p.cachedStruct
}

// clearNonLanguageMapMembers generates the code required to clear all values,
// including unknown values, from this property except for the natural language
// map. If this property can handle a natural language map, then it is up to the
// calling code to determine whether to set the 'langMapMember' to nil.
func (p *FunctionalPropertyGenerator) clearNonLanguageMapMembers() []jen.Code {
	if p.isSingleTypeDef() {
		return p.singleTypeClearNonLanguageMapMembers()
	} else {
		return p.multiTypeClearNonLanguageMapMembers()
	}
}

// singleTypeClearNonLanguageMapMembers generates code to clear all members for
// the special case single-Kind property.
func (p *FunctionalPropertyGenerator) singleTypeClearNonLanguageMapMembers() []jen.Code {
	clearCode := []jen.Code{
		jen.Id(codegen.This()).Dot(unknownMemberName).Op("=").Nil(),
	}
	if !p.hasURIKind() {
		clearCode = append(clearCode, jen.Id(codegen.This()).Dot(iriMember).Op("=").Nil())
	}
	if p.kinds[0].Nilable {
		clearCode = append(clearCode, jen.Id(codegen.This()).Dot(p.memberName(0)).Op("=").Nil())
	} else {
		clearCode = append(clearCode, jen.Id(codegen.This()).Dot(p.hasMemberName(0)).Op("=").False())
	}
	return clearCode
}

// multiTypeClearNonLanguageMapMembers generates code to clear all members for
// a property with multiple Kinds.
func (p *FunctionalPropertyGenerator) multiTypeClearNonLanguageMapMembers() []jen.Code {
	clearLine := make([]jen.Code, len(p.kinds)+2) // +2 for the unknown, and maybe language map
	for i, kind := range p.kinds {
		if kind.Nilable {
			clearLine[i] = jen.Id(codegen.This()).Dot(p.memberName(i)).Op("=").Nil()
		} else {
			clearLine[i] = jen.Id(codegen.This()).Dot(p.hasMemberName(i)).Op("=").False()
		}
	}
	clearLine = append(clearLine, jen.Id(codegen.This()).Dot(unknownMemberName).Op("=").Nil())
	if !p.hasURIKind() {
		clearLine = append(clearLine, jen.Id(codegen.This()).Dot(iriMember).Op("=").Nil())
	}
	return clearLine
}

// funcs produces the methods needed for the functional property.
func (p *FunctionalPropertyGenerator) funcs() []*codegen.Method {
	kindIndexFns := make([]jen.Code, 0, len(p.kinds)+1)
	for i := range p.kinds {
		kindIndexFns = append(kindIndexFns, jen.If(
			jen.Id(codegen.This()).Dot(p.isMethodName(i)).Call(),
		).Block(
			jen.Return(jen.Lit(i)),
		))
	}
	kindIndexFns = append(kindIndexFns,
		jen.If(
			jen.Id(codegen.This()).Dot(isIRIMethod).Call(),
		).Block(
			jen.Return(jen.Lit(iriKindIndex)),
		))
	methods := []*codegen.Method{
		p.contextMethod(),
		codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			kindIndexMethod,
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Int()},
			[]jen.Code{
				join(kindIndexFns),
				jen.Return(jen.Lit(noneOrUnknownKindIndex)),
			},
			fmt.Sprintf("%s computes an arbitrary value for indexing this kind of value. This is a leaky API detail only for folks looking to replace the go-fed implementation. Applications should not use this method.", kindIndexMethod),
		),
	}
	if p.hasTypeKind() {
		typeInterfaceCode := jen.Empty()
		for i, k := range p.kinds {
			if k.isValue() {
				continue
			}
			typeInterfaceCode = typeInterfaceCode.If(
				jen.Id(codegen.This()).Dot(p.isMethodName(i)).Call(),
			).Block(
				jen.Return(
					jen.Id(codegen.This()).Dot(p.getFnName(i)).Call(),
				),
			).Line()
		}
		// GetType
		methods = append(methods,
			codegen.NewCommentedValueMethod(
				p.GetPrivatePackage().Path(),
				fmt.Sprintf("Get%s", typeInterfaceName),
				p.StructName(),
				/*params=*/ nil,
				// Requires the property and type public path to be the same.
				[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), typeInterfaceName)},
				[]jen.Code{
					typeInterfaceCode,
					jen.Return(jen.Nil()),
				},
				fmt.Sprintf("Get%s returns the value in this property as a %s. Returns nil if the value is not an ActivityStreams type, such as an IRI or another value.", typeInterfaceName, typeInterfaceName)))
		// SetType
		setHandlers := jen.Empty()
		for i, k := range p.kinds {
			if k.isValue() {
				continue
			}
			setHandlers = setHandlers.If(
				jen.List(
					jen.Id("v"),
					jen.Id("ok"),
				).Op(":=").Id("t").Assert(k.ConcreteKind),
				jen.Id("ok"),
			).Block(
				jen.Id(codegen.This()).Dot(p.setFnName(i)).Call(jen.Id("v")),
				jen.Return(jen.Nil()),
			).Line()
		}
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				fmt.Sprintf("Set%s", typeInterfaceName),
				p.StructName(),
				// Requires the property and type public path to be the same.
				[]jen.Code{jen.Id("t").Qual(p.GetPublicPackage().Path(), typeInterfaceName)},
				[]jen.Code{jen.Error()},
				[]jen.Code{
					setHandlers,
					jen.Return(jen.Qual("fmt", "Errorf").Call(
						jen.Lit("illegal type to set on "+p.PropertyName()+" property: %T"),
						jen.Id("t"),
					)),
				},
				fmt.Sprintf("Set%s attempts to set the property for the arbitrary type. Returns an error if it is not a valid type to set on this property.", typeInterfaceName)))
	}
	if p.hasNaturalLanguageMap {
		// HasLanguage Method
		methods = append(methods,
			codegen.NewCommentedValueMethod(
				p.GetPrivatePackage().Path(),
				hasLanguageMethod,
				p.StructName(),
				[]jen.Code{jen.Id("bcp47").String()},
				[]jen.Code{jen.Bool()},
				[]jen.Code{
					jen.If(
						jen.Id(codegen.This()).Dot(langMapMember).Op("==").Nil(),
					).Block(
						jen.Return(jen.False()),
					).Else().Block(
						jen.List(
							jen.Id("_"),
							jen.Id("ok"),
						).Op(":=").Id(codegen.This()).Dot(langMapMember).Index(
							jen.Id("bcp47"),
						),
						jen.Return(jen.Id("ok")),
					),
				},
				fmt.Sprintf(
					"%s returns true if the natural language map has an entry for the specified BCP47 language code.",
					hasLanguageMethod,
				),
			))
		// GetLanguage Method
		methods = append(methods,
			codegen.NewCommentedValueMethod(
				p.GetPrivatePackage().Path(),
				getLanguageMethod,
				p.StructName(),
				[]jen.Code{jen.Id("bcp47").String()},
				[]jen.Code{jen.String()},
				[]jen.Code{
					jen.If(
						jen.Id(codegen.This()).Dot(langMapMember).Op("==").Nil(),
					).Block(
						jen.Return(jen.Lit("")),
					).Else().If(
						jen.List(
							jen.Id("v"),
							jen.Id("ok"),
						).Op(":=").Id(codegen.This()).Dot(langMapMember).Index(
							jen.Id("bcp47"),
						),
						jen.Id("ok"),
					).Block(
						jen.Return(jen.Id("v")),
					).Else().Block(
						jen.Return(jen.Lit("")),
					),
				},
				fmt.Sprintf(
					"%s returns the value for the specified BCP47 language code, or an empty string if it is either not a language map or no value is present.",
					getLanguageMethod,
				),
			))
		// SetLanguage Method
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				setLanguageMethod,
				p.StructName(),
				[]jen.Code{
					jen.Id("bcp47"),
					jen.Id("value").String(),
				},
				/*ret=*/ nil,
				append(p.clearNonLanguageMapMembers(),
					[]jen.Code{
						jen.If(
							jen.Id(codegen.This()).Dot(langMapMember).Op("==").Nil(),
						).Block(
							jen.Id(codegen.This()).Dot(langMapMember).Op("=").Make(
								jen.Map(jen.String()).String(),
							),
						),
						jen.Id(codegen.This()).Dot(langMapMember).Index(
							jen.Id("bcp47"),
						).Op("=").Id("value"),
					}...,
				),
				fmt.Sprintf(
					"%s sets the value for the specified BCP47 language code.",
					setLanguageMethod,
				),
			))
	}
	return methods
}

// serializationFuncs produces the Methods and Functions needed for a
// functional property to be serialized and deserialized to and from an
// encoding.
func (p *FunctionalPropertyGenerator) serializationFuncs() (*codegen.Method, *codegen.Function) {
	serializeFns := jen.Empty()
	for i, kind := range p.kinds {
		if i > 0 {
			serializeFns = serializeFns.Else()
		}
		serializeFns = serializeFns.If(
			jen.Id(codegen.This()).Dot(p.isMethodName(i)).Call(),
		)
		if kind.SerializeFn != nil {
			// This is a value that has a function that must be
			// called to serialize properly.
			serializeFns = serializeFns.Block(
				jen.Return(
					kind.SerializeFn.Clone().Call(
						jen.Id(codegen.This()).Dot(p.getFnName(i)).Call(),
					),
				),
			)
		} else {
			// This is a type with a Serialize method.
			serializeFns = serializeFns.Block(
				jen.Return(
					jen.Id(codegen.This()).Dot(p.getFnName(i)).Call().Dot(serializeMethodName).Call(),
				),
			)
		}
	}
	if !p.hasURIKind() {
		serializeFns = serializeFns.Else().If(
			jen.Id(codegen.This()).Dot(isIRIMethod).Call(),
		).Block(
			jen.Return(
				jen.Id(codegen.This()).Dot(iriMember).Dot("String").Call(),
				jen.Nil(),
			),
		)
	}
	serialize := codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		p.serializeFnName(),
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Interface(), jen.Error()},
		[]jen.Code{serializeFns, jen.Return(
			jen.Id(codegen.This()).Dot(unknownMemberName),
			jen.Nil(),
		)},
		fmt.Sprintf("%s converts this into an interface representation suitable for marshalling into a text or binary format. Applications should not need this function as most typical use cases serialize types instead of individual properties. It is exposed for alternatives to go-fed implementations to use.", p.serializeFnName()))
	valueDeserializeFns := jen.Empty()
	typeDeserializeFns := jen.Empty()
	foundValue := false
	foundType := false
	for i, kind := range p.kinds {
		values := jen.Dict{
			jen.Id(p.memberName(i)): jen.Id("v"),
			jen.Id(aliasMember):     jen.Id("alias"),
		}
		if !kind.Nilable {
			values[jen.Id(p.hasMemberName(i))] = jen.True()
		}
		tmp := jen.Empty()
		if kind.isValue() && foundValue {
			tmp = tmp.Else()
		} else if !kind.isValue() && foundType {
			tmp = tmp.Else()
		}
		variable := jen.Id("i")
		if !kind.isValue() {
			variable = jen.Id("m")
		}
		tmp = tmp.If(
			jen.List(
				jen.Id("v"),
				jen.Err(),
			).Op(":=").Add(kind.deserializeFnCode(variable, jen.Id("aliasMap"))),
			jen.Err().Op("==").Nil(),
		).Block(
			jen.Id(codegen.This()).Op(":=").Op("&").Id(p.StructName()).Values(
				values,
			),
			jen.Return(
				jen.Id(codegen.This()),
				jen.Nil(),
			),
		)
		if kind.isValue() {
			foundValue = true
			valueDeserializeFns = valueDeserializeFns.Add(tmp)
		} else {
			foundType = true
			typeDeserializeFns = typeDeserializeFns.Add(tmp)
		}
	}
	mapProperty := jen.Empty()
	if p.hasNaturalLanguageMap {
		mapProperty = jen.If(
			jen.Id("!ok"),
		).Block(
			jen.Commentf("Attempt to find the map instead."),
			jen.List(
				jen.Id("i"),
				jen.Id("ok"),
			).Op("=").Id("m").Index(
				jen.Id("propName").Op("+").Lit("Map"),
			),
		)
	}
	aliasBlock := jen.Empty()
	if p.vocabURI != nil {
		aliasBlock = jen.If(
			jen.List(
				jen.Id("a"),
				jen.Id("ok"),
			).Op(":=").Id("aliasMap").Index(jen.Lit(p.vocabURI.String())),
			jen.Id("ok"),
		).Block(
			jen.Id("alias").Op("=").Id("a"),
		)
	}
	var deserialize *codegen.Function
	if p.asIterator {
		deserialize = codegen.NewCommentedFunction(
			p.GetPrivatePackage().Path(),
			p.DeserializeFnName(),
			[]jen.Code{jen.Id("i").Interface(), jen.Id("aliasMap").Map(jen.String()).String()},
			[]jen.Code{jen.Op("*").Id(p.StructName()), jen.Error()},
			[]jen.Code{
				jen.Id("alias").Op(":=").Lit(""),
				aliasBlock,
				p.wrapDeserializeCode(valueDeserializeFns, typeDeserializeFns),
			},
			fmt.Sprintf("%s creates an iterator from an element that has been unmarshalled from a text or binary format.", p.DeserializeFnName()))
	} else {
		deserialize = codegen.NewCommentedFunction(
			p.GetPrivatePackage().Path(),
			p.DeserializeFnName(),
			[]jen.Code{jen.Id("m").Map(jen.String()).Interface(), jen.Id("aliasMap").Map(jen.String()).String()},
			[]jen.Code{jen.Op("*").Id(p.StructName()), jen.Error()},
			[]jen.Code{
				jen.Id("alias").Op(":=").Lit(""),
				aliasBlock,
				jen.Id("propName").Op(":=").Lit(p.PropertyName()),
				jen.If(
					jen.Len(jen.Id("alias")).Op(">").Lit(0),
				).Block(
					jen.Commentf("Use alias both to find the property, and set within the property."),
					jen.Id("propName").Op("=").Qual("fmt", "Sprintf").Call(
						jen.Lit("%s:%s"),
						jen.Id("alias"),
						jen.Lit(p.PropertyName()),
					),
				),
				jen.List(
					jen.Id("i"),
					jen.Id("ok"),
				).Op(":=").Id("m").Index(
					jen.Id("propName"),
				),
				mapProperty,
				jen.If(jen.Id("ok")).Block(
					p.wrapDeserializeCode(valueDeserializeFns, typeDeserializeFns),
				),
				jen.Return(
					jen.Nil(),
					jen.Nil(),
				),
			},
			fmt.Sprintf("%s creates a %q property from an interface representation that has been unmarshalled from a text or binary format.", p.DeserializeFnName(), p.PropertyName()))
	}
	return serialize, deserialize
}

// singleTypeDef generates a special-case simplified API for a functional
// property that can only be a single Kind of value.
func (p *FunctionalPropertyGenerator) singleTypeDef() *codegen.Struct {
	var comment string
	var kindMembers []jen.Code
	if p.kinds[0].Nilable {
		comment = fmt.Sprintf("%s is the functional property %q. It is permitted to be a single nilable value type.", p.StructName(), p.PropertyName())
		if p.asIterator {
			comment = fmt.Sprintf("%s is an iterator for a property. It is permitted to be a single nilable value type.", p.StructName())
		}
		kindMembers = []jen.Code{
			jen.Id(p.memberName(0)).Add(p.kinds[0].ConcreteKind),
		}
	} else {
		comment = fmt.Sprintf("%s is the functional property %q. It is permitted to be a single default-valued value type.", p.StructName(), p.PropertyName())
		if p.asIterator {
			comment = fmt.Sprintf("%s is an iterator for a property. It is permitted to be a single default-valued value type.", p.StructName())
		}
		kindMembers = []jen.Code{
			jen.Id(p.memberName(0)).Add(p.kinds[0].ConcreteKind),
			jen.Id(p.hasMemberName(0)).Bool(),
		}
	}
	kindMembers = append(kindMembers, p.unknownMemberDef())
	if !p.hasURIKind() {
		kindMembers = append(kindMembers, p.iriMemberDef())
	}
	// TODO: Normalize alias of values when setting on this property.
	kindMembers = append(kindMembers, jen.Id(aliasMember).String())
	if p.asIterator {
		kindMembers = append(kindMembers, jen.Id(myIndexMemberName).Int())
		kindMembers = append(kindMembers, jen.Id(parentMemberName).Qual(p.GetPublicPackage().Path(), p.parentTypeInterfaceName()))
	}
	var methods []*codegen.Method
	var funcs []*codegen.Function
	ser, deser := p.serializationFuncs()
	methods = append(methods, ser)
	funcs = append(funcs, deser)
	funcs = append(funcs, p.ConstructorFn())
	methods = append(methods, p.singleTypeFuncs()...)
	methods = append(methods, p.funcs()...)
	methods = append(methods, p.commonMethods()...)
	methods = append(methods, p.nameMethod())
	return codegen.NewStruct(comment,
		p.StructName(),
		methods,
		funcs,
		kindMembers)
}

// singleTypeFuncs generates the special-case simplified methods for a
// functional property with exactly one Kind of value.
func (p *FunctionalPropertyGenerator) singleTypeFuncs() []*codegen.Method {
	var methods []*codegen.Method
	// HasAny Method
	isLine := jen.Id(codegen.This()).Dot(p.isMethodName(0)).Call()
	if !p.hasURIKind() {
		isLine.Op("||").Id(codegen.This()).Dot(iriMember).Op("!=").Nil()
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		hasAnyMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Bool()},
		[]jen.Code{jen.Return(isLine)},
		fmt.Sprintf("%s returns true if the value or IRI is set.", hasAnyMethod),
	))
	// Is Method
	hasComment := fmt.Sprintf("%s returns true if this property is set and not an IRI.", p.isMethodName(0))
	if p.hasNaturalLanguageMap {
		hasComment = fmt.Sprintf(
			"%s returns true if this property is set and is not a natural language map. When true, the %s and %s methods may be used to access and set this property. To determine if the property was set as a natural language map, use the %s method instead.",
			p.isMethodName(0),
			getMethod,
			p.setFnName(0),
			isLanguageMapMethod,
		)
	}
	if p.kinds[0].Nilable {
		methods = append(methods, codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			p.isMethodName(0),
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Bool()},
			[]jen.Code{jen.Return(jen.Id(codegen.This()).Dot(p.memberName(0)).Op("!=").Nil())},
			hasComment,
		))
	} else {
		methods = append(methods, codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			p.isMethodName(0),
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Bool()},
			[]jen.Code{jen.Return(jen.Id(codegen.This()).Dot(p.hasMemberName(0)))},
			hasComment,
		))
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		isIRIMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Bool()},
		[]jen.Code{jen.Return(p.thisIRI().Op("!=").Nil())},
		fmt.Sprintf("%s returns true if this property is an IRI.", isIRIMethod),
	))
	// Get Method
	getComment := fmt.Sprintf("%s returns the value of this property. When %s returns false, %s will return any arbitrary value.", getMethod, p.isMethodName(0), getMethod)
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		p.getFnName(0),
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{p.kinds[0].ConcreteKind},
		[]jen.Code{jen.Return(jen.Id(codegen.This()).Dot(p.memberName(0)))},
		getComment,
	))
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		getIRIMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Op("*").Qual("net/url", "URL")},
		[]jen.Code{jen.Return(p.thisIRI())},
		fmt.Sprintf("%s returns the IRI of this property. When %s returns false, %s will return any arbitrary value.", getIRIMethod, isIRIMethod, getIRIMethod),
	))
	// Set Method
	setComment := fmt.Sprintf("%s sets the value of this property. Calling %s afterwards will return true.", p.setFnName(0), p.isMethodName(0))
	if p.hasNaturalLanguageMap {
		setComment = fmt.Sprintf(
			"%s sets the value of this property and clears the natural language map. Calling %s afterwards will return true. Calling %s afterwards returns false.",
			p.setFnName(0),
			p.isMethodName(0),
			isLanguageMapMethod,
		)
	}
	if p.kinds[0].Nilable {
		methods = append(methods, codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			p.setFnName(0),
			p.StructName(),
			[]jen.Code{jen.Id("v").Add(p.kinds[0].ConcreteKind)},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(codegen.This()).Dot(p.clearMethodName()).Call(),
				jen.Id(codegen.This()).Dot(p.memberName(0)).Op("=").Id("v"),
			},
			setComment,
		))
	} else {
		methods = append(methods, codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			p.setFnName(0),
			p.StructName(),
			[]jen.Code{jen.Id("v").Add(p.kinds[0].ConcreteKind)},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(codegen.This()).Dot(p.clearMethodName()).Call(),
				jen.Id(codegen.This()).Dot(p.memberName(0)).Op("=").Id("v"),
				jen.Id(codegen.This()).Dot(p.hasMemberName(0)).Op("=").True(),
			},
			setComment,
		))
	}
	methods = append(methods, codegen.NewCommentedPointerMethod(
		p.GetPrivatePackage().Path(),
		setIRIMethod,
		p.StructName(),
		[]jen.Code{jen.Id("v").Op("*").Qual("net/url", "URL")},
		/*ret=*/ nil,
		[]jen.Code{
			jen.Id(codegen.This()).Dot(p.clearMethodName()).Call(),
			p.thisIRISetFn(),
		},
		fmt.Sprintf("%s sets the value of this property. Calling %s afterwards will return true.", setIRIMethod, isIRIMethod),
	))
	// Clear Method
	clearComment := fmt.Sprintf("%s ensures no value of this property is set. Calling %s afterwards will return false.", p.clearMethodName(), p.isMethodName(0))
	clearCode := p.singleTypeClearNonLanguageMapMembers()
	if p.hasNaturalLanguageMap {
		clearComment = fmt.Sprintf(
			"%s ensures no value and no language map for this property is set. Calling %s or %s afterwards will return false.",
			p.clearMethodName(),
			p.isMethodName(0),
			isLanguageMapMethod,
		)
		clearCode = append(clearCode, jen.Id(codegen.This()).Dot(langMapMember).Op("=").Nil())
	}
	methods = append(methods, codegen.NewCommentedPointerMethod(
		p.GetPrivatePackage().Path(),
		p.clearMethodName(),
		p.StructName(),
		/*params=*/ nil,
		/*ret=*/ nil,
		clearCode,
		clearComment,
	))
	// LessThan Method
	lessCode := p.kinds[0].lessFnCode(jen.Id(codegen.This()).Dot(p.getFnName(0)).Call(), jen.Id("o").Dot(p.getFnName(0)).Call())
	iriCmp := jen.Empty()
	if !p.hasURIKind() {
		iriCmp = iriCmp.Add(
			jen.Commentf("LessThan comparison for if either or both are IRIs.").Line(),
			jen.If(
				jen.Id(codegen.This()).Dot(isIRIMethod).Call().Op("&&").Id("o").Dot(isIRIMethod).Call(),
			).Block(
				jen.Return(
					jen.Id(codegen.This()).Dot(iriMember).Dot("String").Call().Op("<").Id("o").Dot(getIRIMethod).Call().Dot("String").Call(),
				),
			).Else())
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		compareLessMethod,
		p.StructName(),
		[]jen.Code{jen.Id("o").Qual(p.GetPublicPackage().Path(), p.InterfaceName())},
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			iriCmp.If(
				jen.Id(codegen.This()).Dot(isIRIMethod).Call(),
			).Block(
				jen.Commentf("IRIs are always less than other values, none, or unknowns"),
				jen.Return(jen.True()),
			).Else().If(
				jen.Id("o").Dot(isIRIMethod).Call(),
			).Block(
				jen.Commentf("This other, none, or unknown value is always greater than IRIs"),
				jen.Return(jen.False()),
			),
			jen.Commentf("LessThan comparison for the single value or unknown value."),
			jen.If(
				jen.Op("!").Id(codegen.This()).Dot(p.isMethodName(0)).Call().Op("&&").Op("!").Id("o").Dot(p.isMethodName(0)).Call(),
			).Block(
				jen.Commentf("Both are unknowns."),
				jen.Return(jen.False()),
			).Else().If(
				jen.Id(codegen.This()).Dot(p.isMethodName(0)).Call().Op("&&").Op("!").Id("o").Dot(p.isMethodName(0)).Call(),
			).Block(
				jen.Commentf("Values are always greater than unknown values."),
				jen.Return(jen.False()),
			).Else().If(
				jen.Op("!").Id(codegen.This()).Dot(p.isMethodName(0)).Call().Op("&&").Id("o").Dot(p.isMethodName(0)).Call(),
			).Block(
				jen.Commentf("Unknowns are always less than known values."),
				jen.Return(jen.True()),
			).Else().Block(
				jen.Commentf("Actual comparison."),
				jen.Return(lessCode),
			),
		},
		fmt.Sprintf("%s compares two instances of this property with an arbitrary but stable comparison. Applications should not use this because it is only meant to help alternative implementations to go-fed to be able to normalize nonfunctional properties.", compareLessMethod),
	))
	return methods
}

// multiTypeDef generates an API for a functional property that can be multiple
// Kinds of value.
func (p *FunctionalPropertyGenerator) multiTypeDef() *codegen.Struct {
	kindMembers := make([]jen.Code, 0, len(p.kinds))
	for i, kind := range p.kinds {
		if kind.Nilable {
			kindMembers = append(kindMembers, jen.Id(p.memberName(i)).Add(p.kinds[i].ConcreteKind))
		} else {
			kindMembers = append(kindMembers, jen.Id(p.memberName(i)).Add(p.kinds[i].ConcreteKind))
			kindMembers = append(kindMembers, jen.Id(p.hasMemberName(i)).Bool())
		}
	}
	kindMembers = append(kindMembers, p.unknownMemberDef())
	if !p.hasURIKind() {
		kindMembers = append(kindMembers, p.iriMemberDef())
	}
	kindMembers = append(kindMembers, jen.Id(aliasMember).String())
	explanation := "At most, one type of value can be present, or none at all. Setting a value will clear the other types of values so that only one of the 'Is' methods will return true. It is possible to clear all values, so that this property is empty."
	comment := fmt.Sprintf(
		"%s is the functional property %q. It is permitted to be one of multiple value types. %s",
		p.StructName(),
		p.PropertyName(),
		explanation,
	)
	if p.asIterator {
		comment = fmt.Sprintf(
			"%s is an iterator for a property. It is permitted to be one of multiple value types. %s",
			p.StructName(),
			explanation,
		)
		kindMembers = append(kindMembers, jen.Id(myIndexMemberName).Int())
		kindMembers = append(kindMembers, jen.Id(parentMemberName).Qual(p.GetPublicPackage().Path(), p.parentTypeInterfaceName()))
	}
	var methods []*codegen.Method
	var funcs []*codegen.Function
	ser, deser := p.serializationFuncs()
	methods = append(methods, ser)
	funcs = append(funcs, deser)
	funcs = append(funcs, p.ConstructorFn())
	methods = append(methods, p.multiTypeFuncs()...)
	methods = append(methods, p.funcs()...)
	methods = append(methods, p.commonMethods()...)
	methods = append(methods, p.nameMethod())
	return codegen.NewStruct(comment,
		p.StructName(),
		methods,
		funcs,
		kindMembers)
}

// multiTypeFuncs generates the methods for a functional property with more than
// one Kind of value.
func (p *FunctionalPropertyGenerator) multiTypeFuncs() []*codegen.Method {
	var methods []*codegen.Method
	// HasAny Method
	isLine := make([]jen.Code, 0, len(p.kinds)+1)
	for i := range p.kinds {
		or := jen.Empty()
		if i < len(p.kinds)-1 {
			or = jen.Op("||")
		}
		isLine = append(isLine, jen.Id(codegen.This()).Dot(p.isMethodName(i)).Call().Add(or))
	}
	if !p.hasURIKind() {
		isLine[len(isLine)-1] = jen.Add(isLine[len(isLine)-1], jen.Op("||"))
		isLine = append(isLine, jen.Id(codegen.This()).Dot(iriMember).Op("!=").Nil())
	}
	hasAnyComment := fmt.Sprintf(
		"%s returns true if any of the different values is set.", hasAnyMethod,
	)
	if p.hasNaturalLanguageMap {
		hasAnyComment = fmt.Sprintf(
			"%s returns true if any of the values are set, except for the natural language map. When true, the specific has, getter, and setter methods may be used to determine what kind of value there is to access and set this property. To determine if the property was set as a natural language map, use the %s method instead.",
			hasAnyMethod,
			isLanguageMapMethod,
		)
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		hasAnyMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Bool()},
		[]jen.Code{jen.Return(join(isLine))},
		hasAnyComment,
	))
	// Clear Method
	clearComment := fmt.Sprintf(
		"%s ensures no value of this property is set. Calling %s or any of the 'Is' methods afterwards will return false.", p.clearMethodName(), hasAnyMethod,
	)
	clearLine := p.multiTypeClearNonLanguageMapMembers()
	if p.hasNaturalLanguageMap {
		clearComment = fmt.Sprintf(
			"%s ensures no value and no language map for this property is set. Calling %s or any of the 'Is' methods afterwards will return false.",
			p.clearMethodName(),
			hasAnyMethod,
		)
		clearLine = append(clearLine, jen.Id(codegen.This()).Dot(langMapMember).Op("=").Nil())
	}
	methods = append(methods, codegen.NewCommentedPointerMethod(
		p.GetPrivatePackage().Path(),
		p.clearMethodName(),
		p.StructName(),
		/*params=*/ nil,
		/*ret=*/ nil,
		clearLine,
		clearComment,
	))
	// Is Method
	for i, kind := range p.kinds {
		isComment := fmt.Sprintf(
			"%s returns true if this property has a type of %q. When true, use the %s and %s methods to access and set this property.",
			p.isMethodName(i),
			kind.Name.LowerName,
			p.getFnName(i),
			p.setFnName(i),
		)
		if p.hasNaturalLanguageMap {
			isComment = fmt.Sprintf(
				"%s. To determine if the property was set as a natural language map, use the %s method instead.",
				isComment,
				isLanguageMapMethod,
			)
		}
		if kind.Nilable {
			methods = append(methods, codegen.NewCommentedValueMethod(
				p.GetPrivatePackage().Path(),
				p.isMethodName(i),
				p.StructName(),
				/*params=*/ nil,
				[]jen.Code{jen.Bool()},
				[]jen.Code{jen.Return(jen.Id(codegen.This()).Dot(p.memberName(i)).Op("!=").Nil())},
				isComment,
			))
		} else {
			methods = append(methods, codegen.NewCommentedValueMethod(
				p.GetPrivatePackage().Path(),
				p.isMethodName(i),
				p.StructName(),
				/*params=*/ nil,
				[]jen.Code{jen.Bool()},
				[]jen.Code{jen.Return(jen.Id(codegen.This()).Dot(p.hasMemberName(i)))},
				isComment,
			))
		}
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		isIRIMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Bool()},
		[]jen.Code{jen.Return(p.thisIRI().Op("!=").Nil())},
		fmt.Sprintf(
			"%s returns true if this property is an IRI. When true, use %s and %s to access and set this property",
			isIRIMethod,
			getIRIMethod,
			setIRIMethod,
		)))
	// Set Method
	for i, kind := range p.kinds {
		setComment := fmt.Sprintf("%s sets the value of this property. Calling %s afterwards returns true.", p.setFnName(i), p.isMethodName(i))
		if p.hasNaturalLanguageMap {
			setComment = fmt.Sprintf(
				"%s sets the value of this property and clears the natural language map. Calling %s afterwards will return true. Calling %s afterwards returns false.",
				p.setFnName(i),
				p.isMethodName(i),
				isLanguageMapMethod,
			)
		}
		if kind.Nilable {
			methods = append(methods, codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				p.setFnName(i),
				p.StructName(),
				[]jen.Code{jen.Id("v").Add(kind.ConcreteKind)},
				/*ret=*/ nil,
				[]jen.Code{
					jen.Id(codegen.This()).Dot(p.clearMethodName()).Call(),
					jen.Id(codegen.This()).Dot(p.memberName(i)).Op("=").Id("v"),
				},
				setComment,
			))
		} else {
			methods = append(methods, codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				p.setFnName(i),
				p.StructName(),
				[]jen.Code{jen.Id("v").Add(kind.ConcreteKind)},
				/*ret=*/ nil,
				[]jen.Code{
					jen.Id(codegen.This()).Dot(p.clearMethodName()).Call(),
					jen.Id(codegen.This()).Dot(p.memberName(i)).Op("=").Id("v"),
					jen.Id(codegen.This()).Dot(p.hasMemberName(i)).Op("=").True(),
				},
				setComment,
			))
		}
	}
	methods = append(methods, codegen.NewCommentedPointerMethod(
		p.GetPrivatePackage().Path(),
		setIRIMethod,
		p.StructName(),
		[]jen.Code{jen.Id("v").Op("*").Qual("net/url", "URL")},
		/*ret=*/ nil,
		[]jen.Code{
			jen.Id(codegen.This()).Dot(p.clearMethodName()).Call(),
			p.thisIRISetFn(),
		},
		fmt.Sprintf("%s sets the value of this property. Calling %s afterwards returns true.", setIRIMethod, isIRIMethod),
	))
	// Get Method
	for i, kind := range p.kinds {
		getComment := fmt.Sprintf("%s returns the value of this property. When %s returns false, %s will return an arbitrary value.", p.getFnName(i), p.isMethodName(i), p.getFnName(i))
		methods = append(methods, codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			p.getFnName(i),
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Add(kind.ConcreteKind)},
			[]jen.Code{jen.Return(jen.Id(codegen.This()).Dot(p.memberName(i)))},
			getComment,
		))
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		getIRIMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Op("*").Qual("net/url", "URL")},
		[]jen.Code{jen.Return(p.thisIRI())},
		fmt.Sprintf("%s returns the IRI of this property. When %s returns false, %s will return an arbitrary value.", getIRIMethod, isIRIMethod, getIRIMethod),
	))
	// LessThan Method
	lessCode := jen.Empty().Add(
		jen.Id("idx1").Op(":=").Id(codegen.This()).Dot(kindIndexMethod).Call().Line(),
		jen.Id("idx2").Op(":=").Id("o").Dot(kindIndexMethod).Call().Line(),
		jen.If(jen.Id("idx1").Op("<").Id("idx2")).Block(
			jen.Return(jen.True()),
		).Else().If(jen.Id("idx1").Op(">").Id("idx2")).Block(
			jen.Return(jen.False()),
		))
	for i, kind := range p.kinds {
		lessCode.Add(
			jen.Else().If(
				jen.Id(codegen.This()).Dot(p.isMethodName(i)).Call(),
			).Block(
				jen.Return(kind.lessFnCode(jen.Id(codegen.This()).Dot(p.getFnName(i)).Call(), jen.Id("o").Dot(p.getFnName(i)).Call()))))
	}
	if !p.hasURIKind() {
		lessCode.Add(
			jen.Else().If(
				jen.Id(codegen.This()).Dot(isIRIMethod).Call(),
			).Block(
				jen.Return(
					jen.Id(codegen.This()).Dot(iriMember).Dot("String").Call().Op("<").Id("o").Dot(getIRIMethod).Call().Dot("String").Call(),
				),
			))
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		compareLessMethod,
		p.StructName(),
		[]jen.Code{jen.Id("o").Qual(p.GetPublicPackage().Path(), p.InterfaceName())},
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			lessCode,
			jen.Return(jen.False()),
		},
		fmt.Sprintf("%s compares two instances of this property with an arbitrary but stable comparison. Applications should not use this because it is only meant to help alternative implementations to go-fed to be able to normalize nonfunctional properties.", compareLessMethod),
	))
	return methods
}

// unknownMemberDef returns the definition of a struct member that handles
// a property whose type is unknown.
func (p *FunctionalPropertyGenerator) unknownMemberDef() jen.Code {
	return jen.Id(unknownMemberName).Interface()
}

// iriMemberDef returns the definition of a struct member that handles
// a property whose type is an IRI.
func (p *FunctionalPropertyGenerator) iriMemberDef() jen.Code {
	return jen.Id(iriMember).Op("*").Qual("net/url", "URL")
}

// wrapDeserializeCode generates the "else if it's a []byte" code and IRI code
// used for deserializing unknown values.
func (p *FunctionalPropertyGenerator) wrapDeserializeCode(valueExisting, typeExisting jen.Code) *jen.Statement {
	iriCode := jen.Empty()
	if !p.hasURIKind() {
		iriCode = jen.If(
			jen.List(
				jen.Id("s"),
				jen.Id("ok"),
			).Op(":=").Id("i").Assert(jen.String()),
			jen.Id("ok"),
		).Block(
			// IRI
			jen.List(
				jen.Id("u"),
				jen.Err(),
			).Op(":=").Qual("net/url", "Parse").Call(jen.Id("s")),
			jen.Commentf("If error exists, don't error out -- skip this and treat as unknown string ([]byte) at worst"),
			jen.Commentf("Also, if no scheme exists, don't treat it as a URL -- net/url is greedy"),
			jen.If(jen.Err().Op("==").Nil().Op("&&").Len(jen.Id("u").Dot("Scheme")).Op(">").Lit(0)).Block(
				jen.Id(codegen.This()).Op(":=").Op("&").Id(p.StructName()).Values(
					jen.Dict{
						jen.Id(iriMember):   jen.Id("u"),
						jen.Id(aliasMember): jen.Id("alias"),
					},
				),
				jen.Return(
					jen.Id(codegen.This()),
					jen.Nil(),
				),
			),
		).Line()
	}
	if p.hasTypeKind() {
		iriCode = iriCode.If(
			jen.List(
				jen.Id("m"),
				jen.Id("ok"),
			).Op(":=").Id("i").Assert(jen.Map(jen.String()).Interface()),
			jen.Id("ok"),
		).Block(
			typeExisting,
		).Line()
	}
	if p.hasValueKind() {
		iriCode = iriCode.Add(valueExisting).Line()
	}
	iriCode = iriCode.Add(
		jen.Id(codegen.This()).Op(":=").Op("&").Id(p.StructName()).Values(
			jen.Dict{
				jen.Id(unknownMemberName): jen.Id("i"),
				jen.Id(aliasMember):       jen.Id("alias"),
			},
		),
		jen.Line(),
		jen.Return(
			jen.Id(codegen.This()),
			jen.Nil(),
		),
	)
	return iriCode
}

// contextMethod returns the Context method for this functional property.
func (p *FunctionalPropertyGenerator) contextMethod() *codegen.Method {
	contextKind := jen.Var().Id("child").Map(jen.String()).String().Line()
	hasIfStarted := false
	for i, kind := range p.kinds {
		// Skip raw values, as only types and any of their properties
		// will need to have LD context strings added.
		if kind.isValue() {
			continue
		}
		if hasIfStarted {
			contextKind = contextKind.Else()
		} else {
			hasIfStarted = true
		}
		contextKind.Add(
			jen.If(
				jen.Id(codegen.This()).Dot(p.isMethodName(i)).Call(),
			).Block(
				jen.Id("child").Op("=").Id(codegen.This()).Dot(p.getFnName(i)).Call().Dot(contextMethod).Call()))
	}
	mDef := jen.Var().Id("m").Map(jen.String()).String()
	if p.vocabURI != nil {
		mDef = jen.Id("m").Op(":=").Map(jen.String()).String().Values(
			jen.Dict{
				jen.Lit(p.vocabURI.String()): jen.Id(codegen.This()).Dot(aliasMember),
			},
		)
	}
	return codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		contextMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Map(jen.String()).String()},
		[]jen.Code{
			mDef,
			contextKind,
			jen.Commentf("Since the literal maps in this function are determined at\ncode-generation time, this loop should not overwrite an existing key with a\nnew value."),
			jen.For(
				jen.List(
					jen.Id("k"),
					jen.Id("v"),
				).Op(":=").Range().Id("child"),
			).Block(
				jen.Id("m").Index(jen.Id("k")).Op("=").Id("v"),
			),
			jen.Return(jen.Id("m")),
		},
		fmt.Sprintf("%s returns the JSONLD URIs required in the context string for this property and the specific values that are set. The value in the map is the alias used to import the property's value or values.", contextMethod))
}

// thisIRI returns the statement to access this IRI -- it may be an xsd:anyURI
// or another equivalent type.
func (p *FunctionalPropertyGenerator) thisIRI() *jen.Statement {
	if !p.hasURIKind() {
		return jen.Id(codegen.This()).Dot(iriMember)
	} else {
		for i, k := range p.kinds {
			if k.IsURI {
				return jen.Id(codegen.This()).Dot(p.memberName(i))
			}
		}
	}
	return nil
}

// thisIRI returns the statement to access this IRI -- it may be an xsd:anyURI
// or another equivalent type.
func (p *FunctionalPropertyGenerator) thisIRISetFn() *jen.Statement {
	if !p.hasURIKind() {
		return jen.Id(codegen.This()).Dot(iriMember).Op("=").Id("v")
	} else {
		for i, k := range p.kinds {
			if k.IsURI {
				return jen.Id(codegen.This()).Dot(p.setFnName(i)).Call(jen.Id("v"))
			}
		}
	}
	return nil
}

// hasValueKind returns true if this property has a Kind that is a Value.
func (p *FunctionalPropertyGenerator) hasValueKind() bool {
	for _, k := range p.kinds {
		if k.isValue() {
			return true
		}
	}
	return false
}

// nameMethod returns the Name method for this functional property.
func (p *FunctionalPropertyGenerator) nameMethod() *codegen.Method {
	nameImpl := jen.If(
		jen.Len(jen.Id(codegen.This()).Dot(aliasMember)).Op(">").Lit(0),
	).Block(
		jen.Return(
			jen.Id(codegen.This()).Dot(aliasMember).Op("+").Lit(":").Op("+").Lit(p.PropertyName()),
		),
	).Else().Block(
		jen.Return(
			jen.Lit(p.PropertyName()),
		),
	)
	if p.hasNaturalLanguageMap {
		nameImpl = jen.If(
			jen.Id(codegen.This()).Dot(isLanguageMapMethod).Call(),
		).Block(
			jen.Return(
				jen.Lit(p.PropertyName() + "Map"),
			),
		).Else().Block(
			jen.Return(
				jen.Lit(p.PropertyName()),
			),
		)
	}
	return codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		nameMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.String()},
		[]jen.Code{
			nameImpl,
		},
		fmt.Sprintf("%s returns the name of this property: %q.", nameMethod, p.PropertyName()),
	)
}
