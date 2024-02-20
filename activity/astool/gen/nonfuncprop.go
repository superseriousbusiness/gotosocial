package gen

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

const (
	propertiesName = "properties"
)

// NonFunctionalPropertyGenerator produces Go code for properties that can have
// more than one value. The resulting property is a type that is a list of
// iterators; each iterator is a concrete struct type. The property can be
// sorted and iterated over so individual elements can be inspected.
type NonFunctionalPropertyGenerator struct {
	PropertyGenerator
	cacheOnce    sync.Once
	cachedIter   *codegen.Struct
	cachedStruct *codegen.Struct
}

// NewNonFunctionalPropertyGenerator is a convenience constructor to create
// NonFunctionalPropertyGenerators.
//
// PropertyGenerators shoulf be in the first pass to construct, before types and
// other generators are constructed.
func NewNonFunctionalPropertyGenerator(vocabName string,
	vocabURI *url.URL,
	vocabAlias string,
	pm *PackageManager,
	name Identifier,
	comment string,
	kinds []Kind,
	hasNaturalLanguageMap bool) (*NonFunctionalPropertyGenerator, error) {
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
	return &NonFunctionalPropertyGenerator{
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

// InterfaceDefinitions creates interface definitions in the provided package.
func (p *NonFunctionalPropertyGenerator) InterfaceDefinitions(pkg Package) []*codegen.Interface {
	s, t := p.Definitions()
	return []*codegen.Interface{
		s.ToInterface(pkg.Path(), p.elementTypeGenerator().InterfaceName(), fmt.Sprintf("%s represents a single value for the %q property.", p.elementTypeGenerator().InterfaceName(), p.PropertyName())),
		t.ToInterface(pkg.Path(), p.InterfaceName(), p.Comments()),
	}
}

// Definitions produces the Go code definitions, which can generate their Go
// implementations. The struct is the iterator for various values of the
// property, which is defined by the type definition.
//
// The TypeGenerator apply must be called for all types before Definition is
// called.
func (p *NonFunctionalPropertyGenerator) Definitions() (*codegen.Struct, *codegen.Struct) {
	p.cacheOnce.Do(func() {
		var methods []*codegen.Method
		var funcs []*codegen.Function
		ser, deser := p.serializationFuncs()
		methods = append(methods, ser)
		funcs = append(funcs, deser)
		funcs = append(funcs, p.ConstructorFn())
		methods = append(methods, p.funcs()...)
		property := codegen.NewStruct(
			fmt.Sprintf("%s is the non-functional property %q. It is permitted to have one or more values, and of different value types.", p.StructName(), p.PropertyName()),
			p.StructName(),
			methods,
			funcs,
			[]jen.Code{
				jen.Id(propertiesName).Index().Op("*").Id(p.iteratorTypeName().CamelName),
				jen.Id(aliasMember).String(),
			})
		iterator := p.elementTypeGenerator().Definition()
		p.cachedIter, p.cachedStruct = iterator, property
	})
	return p.cachedIter, p.cachedStruct
}

// iteratorInterfaceName gets the interface name for the iterator.
func (p *NonFunctionalPropertyGenerator) iteratorInterfaceName() string {
	return strings.Title(p.iteratorTypeName().CamelName)
}

// elementTypeGenerator produces a FunctionalPropertyGenerator for the iterator
// type.
func (p *NonFunctionalPropertyGenerator) elementTypeGenerator() *FunctionalPropertyGenerator {
	return &FunctionalPropertyGenerator{
		PropertyGenerator: PropertyGenerator{
			vocabName:             p.vocabName,
			vocabURI:              p.vocabURI,
			vocabAlias:            p.vocabAlias,
			packageManager:        p.PropertyGenerator.packageManager,
			name:                  p.iteratorTypeName(),
			kinds:                 p.kinds,
			hasNaturalLanguageMap: p.PropertyGenerator.hasNaturalLanguageMap,
			asIterator:            true,
		},
	}
}

// funcs produces the methods needed for the NonFunctional property.
func (p *NonFunctionalPropertyGenerator) funcs() []*codegen.Method {
	var methods []*codegen.Method
	less := jen.Empty()
	for i, kind := range p.kinds {
		dict := jen.Dict{
			jen.Id(parentMemberName): jen.Id(codegen.This()),
			jen.Id(p.memberName(i)):  jen.Id("v"),
			jen.Id(aliasMember):      jen.Id(codegen.This()).Dot(aliasMember),
		}
		if !kind.Nilable {
			dict[jen.Id(p.hasMemberName(i))] = jen.True()
		}
		// Prepend Method
		prependDict := jen.Dict{}
		for k, v := range dict {
			prependDict[k] = v
		}
		prependDict[jen.Id(myIndexMemberName)] = jen.Lit(0)
		prependMethodName := fmt.Sprintf("%s%s%s", prependMethod, kind.Vocab, p.kindCamelName(i))
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				prependMethodName,
				p.StructName(),
				[]jen.Code{jen.Id("v").Add(kind.ConcreteKind)},
				/*ret=*/ nil,
				[]jen.Code{
					jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
						jen.Index().Op("*").Id(p.iteratorTypeName().CamelName).Values(
							jen.Values(prependDict),
						),
						jen.Id(codegen.This()).Dot(propertiesName).Op("..."),
					),
					jen.For(
						jen.Id("i").Op(":=").Lit(1),
						jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
						jen.Id("i").Op("++"),
					).Block(
						jen.Parens(
							jen.Id(codegen.This()).Dot(propertiesName),
						).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
					),
				},
				fmt.Sprintf("%s prepends a %s value to the front of a list of the property %q. Invalidates all iterators.", prependMethodName, kind.Name.LowerName, p.PropertyName())))
		// Insert Method
		insertDict := jen.Dict{}
		for k, v := range dict {
			insertDict[k] = v
		}
		insertDict[jen.Id(myIndexMemberName)] = jen.Id("idx")
		insertMethodName := fmt.Sprintf("%s%s%s", insertMethod, kind.Vocab, p.kindCamelName(i))
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				insertMethodName,
				p.StructName(),
				[]jen.Code{
					jen.Id("idx").Int(),
					jen.Id("v").Add(kind.ConcreteKind),
				},
				/*ret=*/ nil,
				[]jen.Code{
					jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
						jen.Id(codegen.This()).Dot(propertiesName),
						jen.Nil(),
					),
					jen.Copy(
						jen.Id(codegen.This()).Dot(propertiesName).Index(
							jen.Id("idx").Op("+").Lit(1),
							jen.Empty(),
						),
						jen.Id(codegen.This()).Dot(propertiesName).Index(
							jen.Id("idx"),
							jen.Empty(),
						),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Index(
						jen.Id("idx"),
					).Op("=").Op("&").Id(p.iteratorTypeName().CamelName).Values(
						insertDict,
					),
					jen.For(
						jen.Id("i").Op(":=").Id("idx"),
						jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
						jen.Id("i").Op("++"),
					).Block(
						jen.Parens(
							jen.Id(codegen.This()).Dot(propertiesName),
						).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
					),
				},
				fmt.Sprintf("%s inserts a %s value at the specified index for a property %q. Existing elements at that index and higher are shifted back once. Invalidates all iterators.", insertMethodName, kind.Name.LowerName, p.PropertyName())))
		// Append Method
		appendDict := jen.Dict{}
		for k, v := range dict {
			appendDict[k] = v
		}
		appendDict[jen.Id(myIndexMemberName)] = jen.Id(codegen.This()).Dot(lenMethod).Call()
		appendMethodName := fmt.Sprintf("%s%s%s", appendMethod, kind.Vocab, p.kindCamelName(i))
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				appendMethodName,
				p.StructName(),
				[]jen.Code{jen.Id("v").Add(kind.ConcreteKind)},
				/*ret=*/ nil,
				[]jen.Code{
					jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
						jen.Id(codegen.This()).Dot(propertiesName),
						jen.Op("&").Id(p.iteratorTypeName().CamelName).Values(
							appendDict,
						),
					),
				},
				fmt.Sprintf("%s appends a %s value to the back of a list of the property %q. Invalidates iterators that are traversing using %s.", appendMethodName, kind.Name.LowerName, p.PropertyName(), prevMethod)))
		// Set Method
		setDict := jen.Dict{}
		for k, v := range dict {
			setDict[k] = v
		}
		setDict[jen.Id(myIndexMemberName)] = jen.Id("idx")
		setMethodName := p.setFnName(i)
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				setMethodName,
				p.StructName(),
				[]jen.Code{jen.Id("idx").Int(), jen.Id("v").Add(kind.ConcreteKind)},
				/*ret=*/ nil,
				[]jen.Code{
					jen.Parens(jen.Id(codegen.This()).Dot(propertiesName)).Index(jen.Id("idx")).Dot(parentMemberName).Op("=").Nil(),
					jen.Parens(jen.Id(codegen.This()).Dot(propertiesName)).Index(jen.Id("idx")).Op("=").Op("&").Id(p.iteratorTypeName().CamelName).Values(
						setDict,
					),
				},
				fmt.Sprintf("%s sets a %s value to be at the specified index for the property %q. Panics if the index is out of bounds. Invalidates all iterators.", setMethodName, kind.Name.LowerName, p.PropertyName())))
		// Less logic
		if i > 0 {
			less.Else()
		}
		lessCall := kind.lessFnCode(jen.Id("lhs"), jen.Id("rhs"))
		less.If(
			jen.Id("idx1").Op("==").Lit(i),
		).Block(
			jen.Id("lhs").Op(":=").Id(codegen.This()).Dot(propertiesName).Index(jen.Id("i")).Dot(p.getFnName(i)).Call(),
			jen.Id("rhs").Op(":=").Id(codegen.This()).Dot(propertiesName).Index(jen.Id("j")).Dot(p.getFnName(i)).Call(),
			jen.Return(lessCall),
		)
	}
	// IRI Prepend, Insert, Append, Set, and Less logic
	methods = append(methods,
		codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			fmt.Sprintf("%sIRI", prependMethod),
			p.StructName(),
			[]jen.Code{jen.Id("v").Op("*").Qual("net/url", "URL")},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
					jen.Index().Op("*").Id(p.iteratorTypeName().CamelName).Values(
						jen.Values(jen.Dict{
							p.thisIRI():               jen.Id("v"),
							jen.Id(parentMemberName):  jen.Id(codegen.This()),
							jen.Id(myIndexMemberName): jen.Lit(0),
							jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
						}),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Op("..."),
				),
				jen.For(
					jen.Id("i").Op(":=").Lit(1),
					jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
					jen.Id("i").Op("++"),
				).Block(
					jen.Parens(
						jen.Id(codegen.This()).Dot(propertiesName),
					).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
				),
			},
			fmt.Sprintf("%sIRI prepends an IRI value to the front of a list of the property %q.", prependMethod, p.PropertyName())))
	methods = append(methods,
		codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			fmt.Sprintf("%sIRI", insertMethod),
			p.StructName(),
			[]jen.Code{
				jen.Id("idx").Int(),
				jen.Id("v").Op("*").Qual("net/url", "URL"),
			},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
					jen.Id(codegen.This()).Dot(propertiesName),
					jen.Nil(),
				),
				jen.Copy(
					jen.Id(codegen.This()).Dot(propertiesName).Index(
						jen.Id("idx").Op("+").Lit(1),
						jen.Empty(),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Index(
						jen.Id("idx"),
						jen.Empty(),
					),
				),
				jen.Id(codegen.This()).Dot(propertiesName).Index(
					jen.Id("idx"),
				).Op("=").Op("&").Id(p.iteratorTypeName().CamelName).Values(
					jen.Dict{
						p.thisIRI():               jen.Id("v"),
						jen.Id(parentMemberName):  jen.Id(codegen.This()),
						jen.Id(myIndexMemberName): jen.Id("idx"),
						jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
					},
				),
				jen.For(
					jen.Id("i").Op(":=").Id("idx"),
					jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
					jen.Id("i").Op("++"),
				).Block(
					jen.Parens(
						jen.Id(codegen.This()).Dot(propertiesName),
					).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
				),
			},
			fmt.Sprintf("%s inserts an IRI value at the specified index for a property %q. Existing elements at that index and higher are shifted back once. Invalidates all iterators.", insertMethod, p.PropertyName())))
	methods = append(methods,
		codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			fmt.Sprintf("%sIRI", appendMethod),
			p.StructName(),
			[]jen.Code{jen.Id("v").Op("*").Qual("net/url", "URL")},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
					jen.Id(codegen.This()).Dot(propertiesName),
					jen.Op("&").Id(p.iteratorTypeName().CamelName).Values(
						jen.Dict{
							p.thisIRI():               jen.Id("v"),
							jen.Id(parentMemberName):  jen.Id(codegen.This()),
							jen.Id(myIndexMemberName): jen.Id(codegen.This()).Dot(lenMethod).Call(),
							jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
						},
					),
				),
			},
			fmt.Sprintf("%sIRI appends an IRI value to the back of a list of the property %q", appendMethod, p.PropertyName())))
	methods = append(methods,
		codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			fmt.Sprintf("%sIRI", setMethod),
			p.StructName(),
			[]jen.Code{jen.Id("idx").Int(), jen.Id("v").Op("*").Qual("net/url", "URL")},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Parens(jen.Id(codegen.This()).Dot(propertiesName)).Index(jen.Id("idx")).Dot(parentMemberName).Op("=").Nil(),
				jen.Parens(jen.Id(codegen.This()).Dot(propertiesName)).Index(jen.Id("idx")).Op("=").Op("&").Id(p.iteratorTypeName().CamelName).Values(
					jen.Dict{
						p.thisIRI():               jen.Id("v"),
						jen.Id(parentMemberName):  jen.Id(codegen.This()),
						jen.Id(myIndexMemberName): jen.Id("idx"),
						jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
					},
				),
			},
			fmt.Sprintf("%sIRI sets an IRI value to be at the specified index for the property %q. Panics if the index is out of bounds.", setMethod, p.PropertyName())))
	less = less.Else().If(
		jen.Id("idx1").Op("==").Lit(iriKindIndex),
	).Block(
		jen.Id("lhs").Op(":=").Id(codegen.This()).Dot(propertiesName).Index(jen.Id("i")).Dot(getIRIMethod).Call(),
		jen.Id("rhs").Op(":=").Id(codegen.This()).Dot(propertiesName).Index(jen.Id("j")).Dot(getIRIMethod).Call(),
		jen.Return(
			jen.Id("lhs").Dot("String").Call().Op("<").Id("rhs").Dot("String").Call(),
		),
	)
	// Remove Method
	methods = append(methods,
		codegen.NewCommentedPointerMethod(
			p.GetPrivatePackage().Path(),
			removeMethod,
			p.StructName(),
			[]jen.Code{jen.Id("idx").Int()},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Parens(jen.Id(codegen.This()).Dot(propertiesName)).Index(jen.Id("idx")).Dot(parentMemberName).Op("=").Nil(),
				jen.Copy(
					jen.Parens(
						jen.Id(codegen.This()).Dot(propertiesName),
					).Index(
						jen.Id("idx"),
						jen.Empty(),
					),
					jen.Parens(
						jen.Id(codegen.This()).Dot(propertiesName),
					).Index(
						jen.Id("idx").Op("+").Lit(1),
						jen.Empty(),
					),
				),
				jen.Parens(
					jen.Id(codegen.This()).Dot(propertiesName),
				).Index(
					jen.Len(jen.Id(codegen.This()).Dot(propertiesName)).Op("-").Lit(1),
				).Op("=").Op("&").Id(p.iteratorTypeName().CamelName).Values(),
				jen.Id(codegen.This()).Dot(propertiesName).Op("=").Parens(
					jen.Id(codegen.This()).Dot(propertiesName),
				).Index(
					jen.Empty(),
					jen.Len(jen.Id(codegen.This()).Dot(propertiesName)).Op("-").Lit(1),
				),
				jen.For(
					jen.Id("i").Op(":=").Id("idx"),
					jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
					jen.Id("i").Op("++"),
				).Block(
					jen.Parens(
						jen.Id(codegen.This()).Dot(propertiesName),
					).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
				),
			},
			fmt.Sprintf("%s deletes an element at the specified index from a list of the property %q, regardless of its type. Panics if the index is out of bounds. Invalidates all iterators.", removeMethod, p.PropertyName())))
	// Len Method
	methods = append(methods,
		codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			lenMethod,
			p.StructName(),
			/*params=*/ nil,
			[]jen.Code{jen.Id("length").Int()},
			[]jen.Code{
				jen.Return(
					jen.Len(
						jen.Id(codegen.This()).Dot(propertiesName),
					),
				),
			},
			fmt.Sprintf("%s returns the number of values that exist for the %q property.", lenMethod, p.PropertyName())))
	// Swap Method
	methods = append(methods,
		codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			swapMethod,
			p.StructName(),
			[]jen.Code{
				jen.Id("i"),
				jen.Id("j").Int(),
			},
			/*ret=*/ nil,
			[]jen.Code{
				jen.List(
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("i")),
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("j")),
				).Op("=").List(
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("j")),
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("i")),
				),
			},
			fmt.Sprintf("%s swaps the location of values at two indices for the %q property.", swapMethod, p.PropertyName())))
	// Less Method
	methods = append(methods,
		codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			lessMethod,
			p.StructName(),
			[]jen.Code{
				jen.Id("i"),
				jen.Id("j").Int(),
			},
			[]jen.Code{jen.Bool()},
			[]jen.Code{
				jen.Id("idx1").Op(":=").Id(codegen.This()).Dot(kindIndexMethod).Call(jen.Id("i")),
				jen.Id("idx2").Op(":=").Id(codegen.This()).Dot(kindIndexMethod).Call(jen.Id("j")),
				jen.If(jen.Id("idx1").Op("<").Id("idx2")).Block(
					jen.Return(jen.True()),
				).Else().If(jen.Id("idx1").Op("==").Id("idx2")).Block(
					less,
				),
				jen.Return(jen.False()),
			},
			fmt.Sprintf("%s computes whether another property is less than this one. Mixing types results in a consistent but arbitrary ordering", lessMethod)))
	// KindIndex Method
	methods = append(methods,
		codegen.NewCommentedValueMethod(
			p.GetPrivatePackage().Path(),
			kindIndexMethod,
			p.StructName(),
			[]jen.Code{jen.Id("idx").Int()},
			[]jen.Code{jen.Int()},
			[]jen.Code{
				jen.Return(
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("idx")).Dot(kindIndexMethod).Call(),
				),
			},
			fmt.Sprintf("%s computes an arbitrary value for indexing this kind of value. This is a leaky API method specifically needed only for alternate implementations for go-fed. Applications should not use this method. Panics if the index is out of bounds.", kindIndexMethod)))
	// LessThan Method
	lessCode := jen.Empty().Add(
		jen.Id("l1").Op(":=").Id(codegen.This()).Dot(lenMethod).Call().Line(),
		jen.Id("l2").Op(":=").Id("o").Dot(lenMethod).Call().Line(),
		jen.Id("l").Op(":=").Id("l1").Line(),
		jen.If(
			jen.Id("l2").Op("<").Id("l1"),
		).Block(
			jen.Id("l").Op("=").Id("l2"),
		))
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		compareLessMethod,
		p.StructName(),
		[]jen.Code{jen.Id("o").Qual(p.GetPublicPackage().Path(), p.InterfaceName())},
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			lessCode,
			jen.For(
				jen.Id("i").Op(":=").Lit(0),
				jen.Id("i").Op("<").Id("l"),
				jen.Id("i").Op("++"),
			).Block(
				jen.If(
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("i")).Dot(compareLessMethod).Call(jen.Id("o").Dot(atMethodName).Call(jen.Id("i"))),
				).Block(
					jen.Return(jen.True()),
				).Else().If(
					jen.Id("o").Dot(atMethodName).Call(jen.Id("i")).Dot(compareLessMethod).Call(jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("i"))),
				).Block(
					jen.Return(jen.False()),
				),
			),
			jen.Return(jen.Id("l1").Op("<").Id("l2")),
		},
		fmt.Sprintf("%s compares two instances of this property with an arbitrary but stable comparison. Applications should not use this because it is only meant to help alternative implementations to go-fed to be able to normalize nonfunctional properties.", compareLessMethod),
	))
	// At Method
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		atMethodName,
		p.StructName(),
		[]jen.Code{jen.Id("index").Int()},
		[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), p.iteratorInterfaceName())},
		[]jen.Code{
			jen.Return(
				jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Id("index")),
			),
		},
		fmt.Sprintf("%s returns the property value for the specified index. Panics if the index is out of bounds.", atMethodName)))
	// Empty Method
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		emptyMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Bool()},
		[]jen.Code{
			jen.Return(
				jen.Id(codegen.This()).Dot(lenMethod).Call().Op("==").Lit(0),
			),
		},
		fmt.Sprintf("%s returns returns true if there are no elements.", emptyMethod)))
	// Begin Method
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		beginMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), p.iteratorInterfaceName())},
		[]jen.Code{
			jen.If(
				jen.Id(codegen.This()).Dot(emptyMethod).Call(),
			).Block(
				jen.Return(jen.Nil()),
			).Else().Block(
				jen.Return(
					jen.Id(codegen.This()).Dot(propertiesName).Index(jen.Lit(0)),
				),
			),
		},
		fmt.Sprintf("%s returns the first iterator, or nil if empty. Can be used with the iterator's %s method and this property's %s method to iterate from front to back through all values.", beginMethod, nextMethod, endMethod)))
	// End Method
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		endMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), p.iteratorInterfaceName())},
		[]jen.Code{
			jen.Return(jen.Nil()),
		},
		fmt.Sprintf("%s returns beyond-the-last iterator, which is nil. Can be used with the iterator's %s method and this property's %s method to iterate from front to back through all values.", endMethod, nextMethod, beginMethod)))
	// Context Method
	mDef := jen.Var().Id("m").Map(jen.String()).String()
	if p.vocabURI != nil {
		mDef = jen.Id("m").Op(":=").Map(jen.String()).String().Values(
			jen.Dict{
				jen.Lit(p.vocabURI.String()): jen.Id(codegen.This()).Dot(aliasMember),
			},
		)
	}
	methods = append(methods, codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		contextMethod,
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Map(jen.String()).String()},
		[]jen.Code{
			mDef,
			jen.For(
				jen.List(
					jen.Id("_"),
					jen.Id("elem"),
				).Op(":=").Range().Id(codegen.This()).Dot(propertiesName),
			).Block(
				jen.Id("child").Op(":=").Id("elem").Dot(contextMethod).Call(),
				jen.Commentf("Since the literal maps in this function are determined at\ncode-generation time, this loop should not overwrite an existing key with a\nnew value."),
				jen.For(
					jen.List(
						jen.Id("k"),
						jen.Id("v"),
					).Op(":=").Range().Id("child"),
				).Block(
					jen.Id("m").Index(jen.Id("k")).Op("=").Id("v"),
				),
			),
			jen.Return(jen.Id("m")),
		},
		fmt.Sprintf("%s returns the JSONLD URIs required in the context string for this property and the specific values that are set. The value in the map is the alias used to import the property's value or values.", contextMethod)))
	if p.hasTypeKind() {
		// SetType Method
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				fmt.Sprintf("%s%s", setMethod, typeInterfaceName),
				p.StructName(),
				// Requires the property and type public path to be the same.
				[]jen.Code{
					jen.Id("idx").Int(),
					jen.Id("t").Qual(p.GetPublicPackage().Path(), typeInterfaceName),
				},
				[]jen.Code{jen.Error()},
				[]jen.Code{
					jen.Id("n").Op(":=").Op("&").Id(
						p.iteratorTypeName().CamelName,
					).Values(
						jen.Dict{
							jen.Id(myIndexMemberName): jen.Id("idx"),
							jen.Id(parentMemberName):  jen.Id(codegen.This()),
							jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
						},
					),
					jen.If(
						jen.Err().Op(":=").Id("n").Dot(
							fmt.Sprintf("Set%s", typeInterfaceName),
						).Call(
							jen.Id("t"),
						),
						jen.Err().Op("!=").Nil(),
					).Block(
						jen.Return(jen.Err()),
					),
					jen.Parens(jen.Id(codegen.This()).Dot(propertiesName)).Index(jen.Id("idx")).Op("=").Id("n"),
					jen.Return(jen.Nil()),
				},
				fmt.Sprintf("%s%s sets an arbitrary type value to the specified index of the property %q. Invalidates all iterators. Returns an error if the type is not a valid one to set for this property. Panics if the index is out of bounds.", setMethod, typeInterfaceName, p.PropertyName())))
		// PrependType Method
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				fmt.Sprintf("%s%s", prependMethod, typeInterfaceName),
				p.StructName(),
				// Requires the property and type public path to be the same.
				[]jen.Code{jen.Id("t").Qual(p.GetPublicPackage().Path(), typeInterfaceName)},
				[]jen.Code{jen.Error()},
				[]jen.Code{
					jen.Id("n").Op(":=").Op("&").Id(
						p.iteratorTypeName().CamelName,
					).Values(
						jen.Dict{
							jen.Id(myIndexMemberName): jen.Lit(0),
							jen.Id(parentMemberName):  jen.Id(codegen.This()),
							jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
						},
					),
					jen.If(
						jen.Err().Op(":=").Id("n").Dot(
							fmt.Sprintf("Set%s", typeInterfaceName),
						).Call(
							jen.Id("t"),
						),
						jen.Err().Op("!=").Nil(),
					).Block(
						jen.Return(jen.Err()),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
						jen.Index().Op("*").Id(
							p.iteratorTypeName().CamelName,
						).Values(
							jen.Id("n"),
						),
						jen.Id(codegen.This()).Dot(propertiesName).Op("..."),
					),
					jen.For(
						jen.Id("i").Op(":=").Lit(1),
						jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
						jen.Id("i").Op("++"),
					).Block(
						jen.Parens(
							jen.Id(codegen.This()).Dot(propertiesName),
						).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
					),
					jen.Return(jen.Nil()),
				},
				fmt.Sprintf("%s%s prepends an arbitrary type value to the front of a list of the property %q. Invalidates all iterators. Returns an error if the type is not a valid one to set for this property.", prependMethod, typeInterfaceName, p.PropertyName())))
		// InsertType Method
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				fmt.Sprintf("%s%s", insertMethod, typeInterfaceName),
				p.StructName(),
				// Requires the property and type public path to be the same.
				[]jen.Code{
					jen.Id("idx").Int(),
					jen.Id("t").Qual(p.GetPublicPackage().Path(), typeInterfaceName),
				},
				[]jen.Code{jen.Error()},
				[]jen.Code{
					jen.Id("n").Op(":=").Op("&").Id(
						p.iteratorTypeName().CamelName,
					).Values(
						jen.Dict{
							jen.Id(myIndexMemberName): jen.Id("idx"),
							jen.Id(parentMemberName):  jen.Id(codegen.This()),
							jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
						},
					),
					jen.If(
						jen.Err().Op(":=").Id("n").Dot(
							fmt.Sprintf("Set%s", typeInterfaceName),
						).Call(
							jen.Id("t"),
						),
						jen.Err().Op("!=").Nil(),
					).Block(
						jen.Return(jen.Err()),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
						jen.Id(codegen.This()).Dot(propertiesName),
						jen.Nil(),
					),
					jen.Copy(
						jen.Id(codegen.This()).Dot(propertiesName).Index(
							jen.Id("idx").Op("+").Lit(1),
							jen.Empty(),
						),
						jen.Id(codegen.This()).Dot(propertiesName).Index(
							jen.Id("idx"),
							jen.Empty(),
						),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Index(
						jen.Id("idx"),
					).Op("=").Id("n"),
					jen.For(
						jen.Id("i").Op(":=").Id("idx"),
						jen.Id("i").Op("<").Id(codegen.This()).Dot(lenMethod).Call(),
						jen.Id("i").Op("++"),
					).Block(
						jen.Parens(
							jen.Id(codegen.This()).Dot(propertiesName),
						).Index(jen.Id("i")).Dot(myIndexMemberName).Op("=").Id("i"),
					),
					jen.Return(jen.Nil()),
				},
				fmt.Sprintf("%s%s prepends an arbitrary type value to the front of a list of the property %q. Invalidates all iterators. Returns an error if the type is not a valid one to set for this property.", prependMethod, typeInterfaceName, p.PropertyName())))

		// AppendType
		methods = append(methods,
			codegen.NewCommentedPointerMethod(
				p.GetPrivatePackage().Path(),
				fmt.Sprintf("%s%s", appendMethod, typeInterfaceName),
				p.StructName(),
				// Requires the property and type public path to be the same.
				[]jen.Code{jen.Id("t").Qual(p.GetPublicPackage().Path(), typeInterfaceName)},
				[]jen.Code{jen.Error()},
				[]jen.Code{
					jen.Id("n").Op(":=").Op("&").Id(
						p.iteratorTypeName().CamelName,
					).Values(
						jen.Dict{
							jen.Id(myIndexMemberName): jen.Id(codegen.This()).Dot(lenMethod).Call(),
							jen.Id(parentMemberName):  jen.Id(codegen.This()),
							jen.Id(aliasMember):       jen.Id(codegen.This()).Dot(aliasMember),
						},
					),
					jen.If(
						jen.Err().Op(":=").Id("n").Dot(
							fmt.Sprintf("Set%s", typeInterfaceName),
						).Call(
							jen.Id("t"),
						),
						jen.Err().Op("!=").Nil(),
					).Block(
						jen.Return(jen.Err()),
					),
					jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
						jen.Id(codegen.This()).Dot(propertiesName),
						jen.Id("n"),
					),
					jen.Return(jen.Nil()),
				},
				fmt.Sprintf("%s%s prepends an arbitrary type value to the front of a list of the property %q. Invalidates iterators that are traversing using %s. Returns an error if the type is not a valid one to set for this property.", prependMethod, typeInterfaceName, p.PropertyName(), prevMethod)))
	}
	methods = append(methods, p.commonMethods()...)
	methods = append(methods, p.nameMethod())
	return methods
}

// serializationFuncs produces the Methods and Functions needed for a
// NonFunctional property to be serialized and deserialized to and from an
// encoding.
func (p *NonFunctionalPropertyGenerator) serializationFuncs() (*codegen.Method, *codegen.Function) {
	serialize := codegen.NewCommentedValueMethod(
		p.GetPrivatePackage().Path(),
		p.serializeFnName(),
		p.StructName(),
		/*params=*/ nil,
		[]jen.Code{jen.Interface(), jen.Error()},
		[]jen.Code{
			jen.Id("s").Op(":=").Make(
				jen.Index().Interface(),
				jen.Lit(0),
				jen.Len(jen.Id(codegen.This()).Dot(propertiesName)),
			),
			jen.For(
				jen.List(
					jen.Id("_"),
					jen.Id("iterator"),
				).Op(":=").Range().Id(codegen.This()).Dot(propertiesName),
			).Block(
				jen.If(
					jen.List(
						jen.Id("b"),
						jen.Err(),
					).Op(":=").Id("iterator").Dot(serializeIteratorMethod).Call(),
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(
						jen.Id("s"),
						jen.Err(),
					),
				).Else().Block(
					jen.Id("s").Op("=").Append(
						jen.Id("s"),
						jen.Id("b"),
					),
				),
			),
			jen.Commentf("Shortcut: if serializing one value, don't return an array -- pretty sure other Fediverse software would choke on a \"type\" value with array, for example."),
			jen.If(
				jen.Len(jen.Id("s")).Op("==").Lit(1),
			).Block(
				jen.Return(
					jen.Id("s").Index(jen.Lit(0)),
					jen.Nil(),
				),
			),
			jen.Return(
				jen.Id("s"),
				jen.Nil(),
			),
		},
		fmt.Sprintf("%s converts this into an interface representation suitable for marshalling into a text or binary format. Applications should not need this function as most typical use cases serialize types instead of individual properties. It is exposed for alternatives to go-fed implementations to use.", p.serializeFnName()))
	deserializeFn := func(variable string) jen.Code {
		return jen.If(
			jen.List(
				jen.Id("p"),
				jen.Err(),
			).Op(":=").Id(p.elementTypeGenerator().DeserializeFnName()).Call(
				jen.Id(variable),
				jen.Id("aliasMap"),
			),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Return(
				jen.Id(codegen.This()),
				jen.Err(),
			),
		).Else().If(
			jen.Id("p").Op("!=").Nil(),
		).Block(
			jen.Id(codegen.This()).Dot(propertiesName).Op("=").Append(
				jen.Id(codegen.This()).Dot(propertiesName),
				jen.Id("p"),
			),
		)
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
	deserialize := codegen.NewCommentedFunction(
		p.GetPrivatePackage().Path(),
		p.DeserializeFnName(),
		[]jen.Code{jen.Id("m").Map(jen.String()).Interface(), jen.Id("aliasMap").Map(jen.String()).String()},
		[]jen.Code{jen.Qual(p.GetPublicPackage().Path(), p.InterfaceName()), jen.Error()},
		[]jen.Code{
			jen.Id("alias").Op(":=").Lit(""),
			aliasBlock,
			jen.Id("propName").Op(":=").Lit(p.PropertyName()),
			jen.If(
				jen.Len(jen.Id("alias")).Op(">").Lit(0),
			).Block(
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
			jen.If(
				jen.Id("ok"),
			).Block(
				jen.Id(codegen.This()).Op(":=").Op("&").Id(p.StructName()).Values(
					jen.Dict{
						jen.Id(propertiesName): jen.Index().Op("*").Id(p.iteratorTypeName().CamelName).Values(),
						jen.Id(aliasMember):    jen.Id("alias"),
					},
				),
				jen.If(
					jen.List(
						jen.Id("list"),
						jen.Id("ok"),
					).Op(":=").Id("i").Assert(
						jen.Index().Interface(),
					),
					jen.Id("ok"),
				).Block(
					jen.For(
						jen.List(
							jen.Id("_"),
							jen.Id("iterator"),
						).Op(":=").Range().Id("list"),
					).Block(
						deserializeFn("iterator"),
					),
				).Else().Block(
					deserializeFn("i"),
				),
				jen.Commentf("Set up the properties for iteration."),
				jen.For(
					jen.List(jen.Id("idx"), jen.Id("ele")).Op(":=").Range().Id(codegen.This()).Dot(propertiesName),
				).Block(
					jen.Id("ele").Dot(parentMemberName).Op("=").Id(codegen.This()),
					jen.Id("ele").Dot(myIndexMemberName).Op("=").Id("idx"),
				),
				jen.Return(
					jen.Id(codegen.This()),
					jen.Nil(),
				),
			),
			jen.Return(
				jen.Nil(),
				jen.Nil(),
			),
		},
		fmt.Sprintf("%s creates a %q property from an interface representation that has been unmarshalled from a text or binary format.", p.DeserializeFnName(), p.PropertyName()))
	return serialize, deserialize
}

// thisIRI returns the member to access this IRI -- it may be an xsd:anyURI
// or another equivalent type.
func (p *NonFunctionalPropertyGenerator) thisIRI() *jen.Statement {
	if !p.hasURIKind() {
		return jen.Id(iriMember)
	} else {
		for i, k := range p.kinds {
			if k.IsURI {
				return jen.Id(p.memberName(i))
			}
		}
	}
	return nil
}

// nameMethod returns the Name method for this non-functional property.
func (p *NonFunctionalPropertyGenerator) nameMethod() *codegen.Method {
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
			jen.Id(codegen.This()).Dot(lenMethod).Call().Op("==").Lit(1).Op(
				"&&",
			).Id(codegen.This()).Dot(atMethodName).Call(jen.Lit(0)).Dot(isLanguageMapMethod).Call(),
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
		fmt.Sprintf("%s returns the name of this property (%q) with any alias.", nameMethod, p.PropertyName()),
	)
}
