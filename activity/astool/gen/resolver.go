package gen

import (
	"fmt"
	"sync"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

const (
	contextJSONLDName                = "@context"
	typePropertyName                 = "type"
	jsonResolverStructName           = "JSONResolver"
	typeResolverStructName           = "TypeResolver"
	typePredicatedResolverStructName = "TypePredicatedResolver"
	resolveMethod                    = "Resolve"
	applyMethod                      = "Apply"
	activityStreamInterface          = "ActivityStreamsInterface"
	resolverInterface                = "Resolver"
	callbackMember                   = "callbacks"
	predicateMember                  = "predicate"
	delegateMember                   = "delegate"
	errorNoMatch                     = "ErrNoCallbackMatch"
	errorUnhandled                   = "ErrUnhandledType"
	errorPredicateUnmatched          = "ErrPredicateUnmatched"
	errorCannotTypeAssert            = "errCannotTypeAssertType"
	isUnFnName                       = "IsUnmatchedErr"
	toAliasMapFnName                 = "toAliasMap"
)

// ResolverGenerator generates the code required for the TypeResolver and the
// PredicateTypeResolver.
type ResolverGenerator struct {
	pkg                         Package
	types                       []*TypeGenerator
	manGen                      *ManagerGenerator
	cacheOnce                   sync.Once
	cachedJSON                  *codegen.Struct
	cachedTypePredicate         *codegen.Struct
	cachedType                  *codegen.Struct
	cachedErrNoMatch            jen.Code
	cachedErrUnhandled          jen.Code
	cachedErrPredicateUnmatched jen.Code
	cachedErrCannotTypeAssert   jen.Code
	cachedFns                   []*codegen.Function
	cachedASInterface           *codegen.Interface
	cachedResolverInterface     *codegen.Interface
}

// Creates a new ResolverGenerator for generating all the methods, functions,
// errors, interface, and struct types needed for them.
//
// Must be constructed after all TypeGenerators.
func NewResolverGenerator(
	tgs []*TypeGenerator,
	m *ManagerGenerator,
	pkg Package) *ResolverGenerator {
	return &ResolverGenerator{
		pkg:    pkg,
		types:  tgs,
		manGen: m,
	}
}

// Definition returns the TypeResolver and PredicateTypeResolver.
//
// This function signature is pure garbage and yet I keep heaping it on.
func (r *ResolverGenerator) Definition() (jsonRes, typeRes, typePredRes *codegen.Struct, errs []jen.Code, fns []*codegen.Function, iFaces []*codegen.Interface) {
	r.cacheOnce.Do(func() {
		r.cachedJSON = codegen.NewStruct(
			fmt.Sprintf("%s resolves a JSON-deserialized map into "+
				"its concrete ActivityStreams type", jsonResolverStructName),
			jsonResolverStructName,
			r.jsonResolverMethods(),
			append(r.resolverFunctions(jsonResolverStructName,
				"creates a new Resolver that takes a "+
					"JSON-deserialized generic map and determines "+
					"the correct concrete Go type. The callback "+
					"function is guaranteed to receive a value "+
					"whose underlying ActivityStreams type "+
					"matches the concrete interface name in its "+
					"signature. The callback functions must be of "+
					"the form:\n\n"+
					"  func(context.Context, <TypeInterface>) error\n\n"+
					"where TypeInterface is the code-generated "+
					"interface for an ActivityStream type. An "+
					"error is returned if a callback function "+
					"does not match this signature."),
				r.toAliasFunction()),
			r.resolverMembers())
		r.cachedType = codegen.NewStruct(
			fmt.Sprintf("%s resolves ActivityStreams values based "+
				"on their type name.", typeResolverStructName),
			typeResolverStructName,
			r.typeResolverMethods(),
			r.resolverFunctions(typeResolverStructName,
				"creates a new Resolver that examines the "+
					"type of an ActivityStream value to determine "+
					"what callback function to pass the concretely "+
					"typed value. The callback is guaranteed to "+
					"receive a value whose underlying "+
					"ActivityStreams type matches the concrete "+
					"interface name in its signature. The "+
					"callback functions must be "+
					"of the form:\n\n"+
					"  func(context.Context, <TypeInterface>) error\n\n"+
					"where TypeInterface is the code-generated "+
					"interface for an ActivityStream type. An "+
					"error is returned if a callback function "+
					"does not match this signature."),
			r.resolverMembers())
		r.cachedTypePredicate = codegen.NewStruct(
			fmt.Sprintf("%s resolves ActivityStreams values if "+
				"the value satisfies a predicate condition "+
				"based on its type.", typePredicatedResolverStructName),
			typePredicatedResolverStructName,
			r.typePredicatedResolverMethods(),
			r.predicateResolverFunctions(typePredicatedResolverStructName,
				"creates a new Resolver that applies a "+
					"predicate to an ActivityStreams value to "+
					"determine whether to Resolve or not. The "+
					"ActivityStreams value's type is examined "+
					"to determine if the predicate can apply "+
					"itself to the value. This guarantees the "+
					"predicate will receive a concrete value "+
					"whose underlying ActivityStreams type "+
					"matches the concrete interface name. "+
					"The predicate function must be of the form: \n\n"+
					"  func(context.Context, <TypeInterface>) (bool, error)\n\n"+
					"where TypeInterface is the code-generated "+
					"interface for an ActivityStreams type. An "+
					"error is returned if the predicate does "+
					"not match this signature."),
			r.predicateResolverMembers())
		r.cachedErrNoMatch = r.errorNoMatch()
		r.cachedErrUnhandled = r.errorUnhandled()
		r.cachedErrPredicateUnmatched = r.errorPredicateUnmatched()
		r.cachedErrCannotTypeAssert = r.errorCannotTypeAssert()
		r.cachedFns = r.fns()
		r.cachedASInterface = r.asInterface()
		r.cachedResolverInterface = r.resolverInterface()
	})
	return r.cachedJSON, r.cachedType, r.cachedTypePredicate, []jen.Code{
			r.cachedErrNoMatch,
			r.cachedErrUnhandled,
			r.cachedErrPredicateUnmatched,
			r.cachedErrCannotTypeAssert,
		}, r.cachedFns, []*codegen.Interface{
			r.cachedASInterface,
			r.cachedResolverInterface,
		}
}

// errorNoMatch returns the declaration for the ErrNoMatch global value.
func (r *ResolverGenerator) errorNoMatch() jen.Code {
	return jen.Commentf(
		"%s indicates a Resolver could not match the ActivityStreams value to a "+
			"callback function.",
		errorNoMatch,
	).Line().Var().Id(errorNoMatch).Error().Op("=").Qual("errors", "New").Call(jen.Lit("activity stream did not match the callback function"))
}

// errorUnhandled returns the declaration for the ErrUnhandled global value.
func (r *ResolverGenerator) errorUnhandled() jen.Code {
	return jen.Commentf(
		"%s indicates that an ActivityStreams value has a type that is "+
			"not handled by the code that has been generated.",
		errorUnhandled,
	).Line().Var().Id(errorUnhandled).Error().Op("=").Qual("errors", "New").Call(jen.Lit("activity stream did not match any known types"))
}

// errorCannotTypeAssert returns the declaration for the errCannotTypeAssert
// global value.
func (r *ResolverGenerator) errorCannotTypeAssert() jen.Code {
	return jen.Commentf(
		"%s indicates that the 'type' property returned by the "+
			"ActivityStreams value cannot be type-asserted to its "+
			"interface form.",
		errorCannotTypeAssert,
	).Line().Var().Id(errorCannotTypeAssert).Error().Op("=").Qual("errors", "New").Call(jen.Lit("activity stream type cannot be asserted to its interface"))
}

// errorPredicateUnmatched returns the declaration for the ErrPredicateUnmatched
// global value.
func (r *ResolverGenerator) errorPredicateUnmatched() jen.Code {
	return jen.Commentf(
		"%s indicates that a predicate is accepting a type or "+
			"interface that does not match an ActivityStreams value's "+
			"type or interface.",
		errorPredicateUnmatched,
	).Line().Var().Id(errorPredicateUnmatched).Error().Op("=").Qual("errors", "New").Call(jen.Lit("activity stream did not match type demanded by predicate"))
}

// fns returns all utility functions.
func (r *ResolverGenerator) fns() []*codegen.Function {
	allTypeFns := make([]jen.Code, 0)
	for _, t := range r.types {
		allTypeFns = append(allTypeFns, jen.Func().Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("i").Qual(t.PublicPackage().Path(), t.InterfaceName()),
		).Error().Block(
			jen.Id("t").Op("=").Id("i"),
			jen.Return(jen.Nil()),
		))
	}
	return []*codegen.Function{
		codegen.NewCommentedFunction(
			r.pkg.Path(),
			isUnFnName,
			[]jen.Code{
				jen.Err().Error(),
			},
			[]jen.Code{
				jen.Bool(),
			},
			[]jen.Code{
				jen.Return(
					jen.Err().Op("==").Id(errorPredicateUnmatched).Op(
						"||",
					).Err().Op("==").Id(errorUnhandled).Op(
						"||",
					).Err().Op("==").Id(errorNoMatch),
				),
			},
			fmt.Sprintf("%s is true when the error indicates that a Resolver was unsuccessful due to the ActivityStreams value not matching its callbacks or predicates.", isUnFnName)),
		codegen.NewCommentedFunction(
			r.types[0].PublicPackage().Path(),
			fmt.Sprintf("To%s", typeInterfaceName),
			[]jen.Code{
				jen.Id("c").Qual("context", "Context"),
				jen.Id("m").Map(jen.String()).Interface(),
			},
			[]jen.Code{
				jen.Id("t").Qual(r.types[0].PublicPackage().Path(), typeInterfaceName),
				jen.Err().Error(),
			},
			[]jen.Code{
				jen.Var().Id("r").Op("*").Qual(r.pkg.Path(), jsonResolverStructName),
				jen.List(
					jen.Id("r"),
					jen.Err(),
				).Op("=").Qual(
					r.pkg.Path(),
					fmt.Sprintf("%s%s", constructorName, jsonResolverStructName),
				).Call(
					jen.List(
						allTypeFns...,
					),
				),
				jen.If(
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(),
				),
				jen.Err().Op("=").Id("r").Dot(resolveMethod).Call(
					jen.Id("c"),
					jen.Id("m"),
				),
				jen.Return(),
			},
			fmt.Sprintf("To%s attempts to resolve the generic JSON map into a Type.", typeInterfaceName)),
	}
}

// jsonResolverMethods returns the methods for the TypeResolver.
func (r *ResolverGenerator) jsonResolverMethods() (m []*codegen.Method) {
	aliasToId := make(map[string]string)
	aliasFetching := jen.Empty()
	impl := jen.Empty()
	for i, t := range r.types {
		if i > 0 {
			impl = impl.Else()
		}
		// Get the vocab URI in http and https forms
		vocabHttps := *t.vocabURI
		vocabHttps.Scheme = "https"
		vocabHttp := vocabHttps
		vocabHttp.Scheme = "http"
		// Determine if we've already generated the code for fetching
		// the alias for this vocabulary.
		if _, ok := aliasToId[vocabHttps.String()]; !ok {
			// If not, generate the code.
			vocabId := t.vocabName + "Alias"
			aliasToId[vocabHttps.String()] = vocabId
			aliasFetching = aliasFetching.Add(
				jen.List(
					jen.Id(vocabId),
					jen.Id("ok"),
				).Op(":=").Id("aliasMap").Index(
					jen.Lit(vocabHttps.String()),
				),
			).Line().Add(
				jen.If(
					jen.Op("!").Id("ok"),
				).Block(
					jen.Id(vocabId).Op("=").Id("aliasMap").Index(
						jen.Lit(vocabHttp.String()),
					),
				),
			).Line().Add(
				// If it is not empty post-pend with a ":".
				jen.If(
					jen.Len(jen.Id(vocabId)).Op(">").Lit(0),
				).Block(
					jen.Id(vocabId).Op("+=").Lit(":"),
				),
			).Line()
		}
		// Fetch the identifier holding the alias for this vocabulary,
		aliasId := aliasToId[vocabHttps.String()]
		impl = impl.If(
			jen.Id("typeString").Op("==").Id(aliasId).Op("+").Lit(t.TypeName()),
		).Block(
			jen.List(
				jen.Id("v"),
				jen.Err(),
			).Op(":=").Add(r.manGen.getDeserializationMethodForType(t).On(managerInitVarName).Call().Call(
				jen.Id("m"),
				jen.Id("aliasMap"),
			)),
			jen.If(
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Return(jen.Err()),
			),
			jen.For(
				jen.List(
					jen.Id("_"),
					jen.Id("i"),
				).Op(":=").Range().Id(codegen.This()).Dot(callbackMember),
			).Block(
				jen.If(
					jen.List(
						jen.Id("fn"),
						jen.Id("ok"),
					).Op(":=").Id("i").Assert(
						jen.Func().Parens(
							jen.List(
								jen.Qual("context", "Context"),
								jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
							),
						).Error(),
					),
					jen.Id("ok"),
				).Block(
					jen.Return(
						jen.Id("fn").Call(jen.Id("ctx"), jen.Id("v")),
					),
				),
			),
			jen.Return(
				jen.Id(errorNoMatch),
			),
		)
	}
	m = append(m, codegen.NewCommentedValueMethod(
		r.pkg.Path(),
		resolveMethod,
		jsonResolverStructName,
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("m").Map(jen.String()).Interface(),
		},
		[]jen.Code{
			jen.Error(),
		},
		[]jen.Code{
			jen.List(
				jen.Id("typeValue"),
				jen.Id("ok"),
			).Op(":=").Id("m").Index(jen.Lit(typePropertyName)),
			jen.If(
				jen.Op("!").Id("ok"),
			).Block(
				jen.Return(
					jen.Qual("fmt", "Errorf").Call(
						jen.Lit("cannot determine ActivityStreams type: 'type' property is missing"),
					),
				),
			),
			jen.List(
				jen.Id("rawContext"),
				jen.Id("ok"),
			).Op(":=").Id("m").Index(jen.Lit(contextJSONLDName)),
			jen.If(
				jen.Op("!").Id("ok"),
			).Block(
				jen.Return(
					jen.Qual("fmt", "Errorf").Call(
						jen.Lit("cannot determine ActivityStreams type: '@context' is missing"),
					),
				),
			),
			jen.Id("aliasMap").Op(":=").Id(toAliasMapFnName).Call(jen.Id("rawContext")),
			jen.Commentf("Begin: Private lambda to handle a single string %q value. Makes code generation easier.", typePropertyName),
			jen.Id("handleFn").Op(":=").Func().Parens(
				jen.Id("typeString").String(),
			).Error().Block(
				aliasFetching,
				impl.Else().Block(
					jen.Return(
						jen.Id(errorUnhandled),
					),
				),
			),
			jen.Commentf("End: Private lambda"),
			jen.If(
				jen.List(
					jen.Id("typeStr"),
					jen.Id("ok"),
				).Op(":=").Id("typeValue").Assert(jen.String()),
				jen.Id("ok"),
			).Block(
				jen.Return(
					jen.Id("handleFn").Call(jen.Id("typeStr")),
				),
			).Else().If(
				jen.List(
					jen.Id("typeIArr"),
					jen.Id("ok"),
				).Op(":=").Id("typeValue").Assert(jen.Index().Interface()),
				jen.Id("ok"),
			).Block(
				jen.For(
					jen.List(
						jen.Id("_"),
						jen.Id("typeI"),
					).Op(":=").Range().Id("typeIArr"),
				).Block(
					jen.If(
						jen.List(
							jen.Id("typeStr"),
							jen.Id("ok"),
						).Op(":=").Id("typeI").Assert(jen.String()),
						jen.Id("ok"),
					).Block(
						jen.If(
							jen.List(
								jen.Err(),
							).Op(":=").Id("handleFn").Call(jen.Id("typeStr")),
							jen.Err().Op("==").Nil(),
						).Block(
							jen.Return(jen.Nil()),
						).Else().If(
							jen.Err().Op("==").Id(errorUnhandled),
						).Block(
							jen.Commentf("Keep trying other types: only if all fail do we return this error."),
							jen.Continue(),
						).Else().Block(
							jen.Return(jen.Err()),
						),
					),
				),
				jen.Return(
					jen.Id(errorUnhandled),
				),
			).Else().Block(
				jen.Return(
					jen.Id(errorUnhandled),
				),
			),
		},
		fmt.Sprintf("%s determines the ActivityStreams type of the payload, then applies the first callback function whose signature accepts the ActivityStreams value's type. This strictly assures that the callback function will only be passed ActivityStream objects whose type matches its interface. Returns an error if the ActivityStreams type does not match callbackers or is not a type handled by the generated code. If multiple types are present, it will check each one in order and apply only the first one. It returns an unhandled error for a multi-typed object if none of the types were able to be handled.", resolveMethod)))
	return
}

// typeResolverMethods returns the methods for the TypeResolver.
func (r *ResolverGenerator) typeResolverMethods() (m []*codegen.Method) {
	impl := jen.Empty()
	for i, t := range r.types {
		if i > 0 {
			impl = impl.Else()
		}
		impl = impl.If(
			jen.Id("o").Dot(vocabURIMethod).Call().Op("==").Lit(t.vocabURI.String()).Op(
				"&&",
			).Id("o").Dot(typeNameMethod).Call().Op("==").Lit(t.TypeName()),
		).Block(
			jen.If(
				jen.List(
					jen.Id("fn"),
					jen.Id("ok"),
				).Op(":=").Id("i").Assert(
					jen.Func().Parens(
						jen.List(
							jen.Qual("context", "Context"),
							jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
						),
					).Error(),
				),
				jen.Id("ok"),
			).Block(
				jen.If(
					jen.List(
						jen.Id("v"),
						jen.Id("ok"),
					).Op(":=").Id("o").Assert(
						jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
					),
					jen.Id("ok"),
				).Block(
					jen.Return(
						jen.Id("fn").Call(jen.Id("ctx"), jen.Id("v")),
					),
				).Else().Block(
					jen.Commentf("This occurs when the value is either not a go-fed type and is improperly satisfying various interfaces, or there is a bug in the go-fed generated code."),
					jen.Return(
						jen.Id(errorCannotTypeAssert),
					),
				),
			),
		)
	}
	m = append(m, codegen.NewCommentedValueMethod(
		r.pkg.Path(),
		resolveMethod,
		typeResolverStructName,
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("o").Id(activityStreamInterface),
		},
		[]jen.Code{
			jen.Error(),
		},
		[]jen.Code{
			jen.For(
				jen.List(
					jen.Id("_"),
					jen.Id("i"),
				).Op(":=").Range().Id(codegen.This()).Dot(callbackMember),
			).Block(
				impl.Else().Block(
					jen.Return(
						jen.Id(errorUnhandled),
					),
				),
			),
			jen.Return(
				jen.Id(errorNoMatch),
			),
		},
		fmt.Sprintf("%s applies the first callback function whose signature accepts the ActivityStreams value's type. This strictly assures that the callback function will only be passed ActivityStream objects whose type matches its interface. Returns an error if the ActivityStreams type does not match callbackers, is not a type handled by the generated code, or the value passed in is not go-fed compatible.", resolveMethod)))
	return
}

// typePredicatedResolverMethods returns the methods for the TypePredicatedResolver.
func (r *ResolverGenerator) typePredicatedResolverMethods() (m []*codegen.Method) {
	impl := jen.Empty()
	for i, t := range r.types {
		if i > 0 {
			impl = impl.Else()
		}
		impl = impl.If(
			jen.Id("o").Dot(vocabURIMethod).Call().Op("==").Lit(t.vocabURI.String()).Op(
				"&&",
			).Id("o").Dot(typeNameMethod).Call().Op("==").Lit(t.TypeName()),
		).Block(
			jen.If(
				jen.List(
					jen.Id("fn"),
					jen.Id("ok"),
				).Op(":=").Id(codegen.This()).Dot(predicateMember).Assert(
					jen.Func().Parens(
						jen.List(
							jen.Qual("context", "Context"),
							jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
						),
					).Parens(
						jen.List(
							jen.Bool(),
							jen.Error(),
						),
					),
				),
				jen.Id("ok"),
			).Block(
				jen.If(
					jen.List(
						jen.Id("v"),
						jen.Id("ok"),
					).Op(":=").Id("o").Assert(
						jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
					),
					jen.Id("ok"),
				).Block(
					jen.List(
						jen.Id("predicatePasses"),
						jen.Err(),
					).Op("=").Id("fn").Call(jen.Id("ctx"), jen.Id("v")),
				).Else().Block(
					jen.Commentf("This occurs when the value is either not a go-fed type and is improperly satisfying various interfaces, or there is a bug in the go-fed generated code."),
					jen.Return(
						jen.False(),
						jen.Id(errorCannotTypeAssert),
					),
				),
			).Else().Block(
				jen.Return(
					jen.False(),
					jen.Id(errorPredicateUnmatched),
				),
			),
		)
	}
	m = append(m, codegen.NewCommentedValueMethod(
		r.pkg.Path(),
		applyMethod,
		typePredicatedResolverStructName,
		[]jen.Code{
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("o").Id(activityStreamInterface),
		},
		[]jen.Code{
			jen.Bool(),
			jen.Error(),
		},
		[]jen.Code{
			jen.Var().Id("predicatePasses").Bool(),
			jen.Var().Err().Error(),
			impl.Else().Block(
				jen.Return(
					jen.False(),
					jen.Id(errorUnhandled),
				),
			),
			jen.If(
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Return(
					jen.Id("predicatePasses"),
					jen.Err(),
				),
			),
			jen.If(
				jen.Id("predicatePasses"),
			).Block(
				jen.Return(
					jen.True(),
					jen.Id(codegen.This()).Dot(delegateMember).Dot(resolveMethod).Call(
						jen.Id("ctx"),
						jen.Id("o"),
					),
				),
			).Else().Block(
				jen.Return(
					jen.False(),
					jen.Nil(),
				),
			),
		},
		fmt.Sprintf("%s uses a predicate to determine whether to resolve the ActivityStreams value. The predicate's signature is matched with the ActivityStreams value's type. This strictly assures that the predicate will only be passed ActivityStream objects whose type matches its interface. Returns an error if the ActivityStreams type does not match the predicate, is not a type handled by the generated code, or the resolver returns an error. Returns true if the predicate returned true.", applyMethod)))
	return
}

// resolverFunctions returns the functions for the TypeResolver.
func (r *ResolverGenerator) resolverFunctions(name, comment string) (f []*codegen.Function) {
	f = append(f, codegen.NewCommentedFunction(
		r.pkg.Path(),
		fmt.Sprintf("%s%s", constructorName, name),
		[]jen.Code{
			jen.Id("callbacks").Op("...").Interface(),
		},
		[]jen.Code{
			jen.Op("*").Id(name),
			jen.Error(),
		},
		[]jen.Code{
			jen.For(
				jen.List(
					jen.Id("_"),
					jen.Id("cb"),
				).Op(":=").Range().Id("callbacks"),
			).Block(
				jen.Commentf("Each callback function must satisfy one known function signature, or else we will generate a runtime error instead of silently fail."),
				jen.Switch(
					jen.Id("cb").Assert(jen.Type()),
				).Block(
					r.mustAssertToKnownTypes("cb"),
				),
			),
			jen.Return(
				jen.Op("&").Id(name).Values(
					jen.Dict{
						jen.Id(callbackMember): jen.Id("callbacks"),
					},
				),
				jen.Nil(),
			),
		},
		fmt.Sprintf("%s%s %s", constructorName, name, comment)))
	return
}

// predicateResolverFunctions returns the functions for the PredicateTypeResolver.
func (r *ResolverGenerator) predicateResolverFunctions(name, comment string) (f []*codegen.Function) {
	f = append(f, codegen.NewCommentedFunction(
		r.pkg.Path(),
		fmt.Sprintf("%s%s", constructorName, name),
		[]jen.Code{
			jen.Id("delegate").Id(resolverInterface),
			jen.Id("predicate").Interface(),
		},
		[]jen.Code{
			jen.Op("*").Id(name),
			jen.Error(),
		},
		[]jen.Code{
			jen.Commentf("The predicate must satisfy one known predicate function signature, or else we will generate a runtime error instead of silently fail."),
			jen.Switch(
				jen.Id("predicate").Assert(jen.Type()),
			).Block(
				r.mustAssertToKnownPredicate("predicate"),
			),
			jen.Return(
				jen.Op("&").Id(name).Values(
					jen.Dict{
						jen.Id(delegateMember):  jen.Id("delegate"),
						jen.Id(predicateMember): jen.Id("predicate"),
					},
				),
				jen.Nil(),
			),
		},
		fmt.Sprintf("%s%s %s", constructorName, name, comment)))
	return
}

// resolverMembers returns the members for the TypeResolver.
func (r *ResolverGenerator) resolverMembers() (m []jen.Code) {
	m = append(m, jen.Id(callbackMember).Index().Interface())
	return
}

// predicateResolverMembers returns the members for the PredicateTypResolver.
func (r *ResolverGenerator) predicateResolverMembers() (m []jen.Code) {
	m = append(m, jen.Id(delegateMember).Id(resolverInterface))
	m = append(m, jen.Id(predicateMember).Interface())
	return
}

// mustAssertToKnownTypes creates the type assertion switch statement that will
// return an error if the parameter named does not match any of the expected
// function signatures.
func (r *ResolverGenerator) mustAssertToKnownTypes(paramName string) jen.Code {
	c := jen.Empty()
	for _, t := range r.types {
		c = c.Case(
			jen.Func().Parens(
				jen.List(
					jen.Qual("context", "Context"),
					jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
				),
			).Error(),
		).Block(
			jen.Commentf("Do nothing, this callback has a correct signature."),
		).Line()
	}
	c = c.Default().Block(
		jen.Return(
			jen.Nil(),
			jen.Qual("errors", "New").Call(jen.Lit("a callback function is of the wrong signature and would never be called")),
		),
	)
	return c
}

// mustAssertToKnownPredicate ensures the parameter name types-asserts to a
// known signature, or returns an error.
func (r *ResolverGenerator) mustAssertToKnownPredicate(paramName string) jen.Code {
	c := jen.Empty()
	for _, t := range r.types {
		c = c.Case(
			jen.Func().Parens(
				jen.List(
					jen.Qual("context", "Context"),
					jen.Qual(t.PublicPackage().Path(), t.InterfaceName()),
				),
			).Parens(
				jen.List(
					jen.Bool(),
					jen.Error(),
				),
			),
		).Block(
			jen.Commentf("Do nothing, this predicate has a correct signature."),
		).Line()
	}
	c = c.Default().Block(
		jen.Return(
			jen.Nil(),
			jen.Qual("errors", "New").Call(jen.Lit("the predicate function is of the wrong signature and would never be called")),
		),
	)
	return c
}

// asInterface returns the ActivityStreamsInterface.
func (r *ResolverGenerator) asInterface() *codegen.Interface {
	return codegen.NewInterface(
		r.pkg.Path(),
		activityStreamInterface,
		[]codegen.FunctionSignature{
			{
				Name:    typeNameMethod,
				Params:  nil,
				Ret:     []jen.Code{jen.String()},
				Comment: fmt.Sprintf("%s returns the ActiivtyStreams value's type.", typeNameMethod),
			},
			{
				Name:    vocabURIMethod,
				Params:  nil,
				Ret:     []jen.Code{jen.String()},
				Comment: fmt.Sprintf("%s returns the vocabulary's URI as a string.", vocabURIMethod),
			},
		},
		fmt.Sprintf("%s represents any ActivityStream value code-generated by go-fed or compatible with the generated interfaces.", activityStreamInterface))
}

// resolverInterface returns the Resolver interface.
func (r *ResolverGenerator) resolverInterface() *codegen.Interface {
	return codegen.NewInterface(
		r.pkg.Path(),
		resolverInterface,
		[]codegen.FunctionSignature{
			{
				Name: resolveMethod,
				Params: []jen.Code{
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("o").Id(activityStreamInterface),
				},
				Ret: []jen.Code{
					jen.Error(),
				},
				Comment: fmt.Sprintf("%s will attempt to resolve an untyped ActivityStreams value into a Go concrete type.", resolveMethod),
			},
		},
		fmt.Sprintf("%s represents any %s.", resolverInterface, typeResolverStructName))
}

// toAliasFunction returns the toAliasMap function
func (r *ResolverGenerator) toAliasFunction() *codegen.Function {
	return codegen.NewCommentedFunction(
		r.pkg.Path(),
		toAliasMapFnName,
		[]jen.Code{
			jen.Id("i").Interface(),
		},
		[]jen.Code{
			jen.Id("m").Map(jen.String()).String(),
		},
		[]jen.Code{
			jen.Id("m").Op("=").Make(
				jen.Map(jen.String()).String(),
			),
			jen.Id("toHttpHttpsFn").Op(":=").Func().Parens(
				jen.Id("s").String(),
			).Parens(
				jen.List(
					jen.Id("ok").Bool(),
					jen.Id("http"),
					jen.Id("https").String(),
				),
			).Block(
				jen.If(
					jen.Qual("strings", "HasPrefix").Call(
						jen.Id("s"),
						jen.Lit("http://"),
					),
				).Block(
					jen.Id("ok").Op("=").True(),
					jen.Id("http").Op("=").Id("s"),
					jen.Id("https").Op("=").Lit("https").Op("+").Qual("strings", "TrimPrefix").Call(
						jen.Id("s"),
						jen.Lit("http"),
					),
				).Else().If(
					jen.Qual("strings", "HasPrefix").Call(
						jen.Id("s"),
						jen.Lit("https://"),
					),
				).Block(
					jen.Id("ok").Op("=").True(),
					jen.Id("https").Op("=").Id("s"),
					jen.Id("http").Op("=").Lit("http").Op("+").Qual("strings", "TrimPrefix").Call(
						jen.Id("s"),
						jen.Lit("https"),
					),
				),
				jen.Return(),
			),
			jen.Switch(jen.Id("v").Op(":=").Id("i").Assert(jen.Type())).Block(
				jen.Case(jen.String()).Block(
					jen.Commentf("Single entry, no alias."),
					jen.If(
						jen.List(
							jen.Id("ok"),
							jen.Id("http"),
							jen.Id("https"),
						).Op(":=").Id("toHttpHttpsFn").Call(jen.Id("v")),
						jen.Id("ok"),
					).Block(
						jen.Id("m").Index(
							jen.Id("http"),
						).Op("=").Lit(""),
						jen.Id("m").Index(
							jen.Id("https"),
						).Op("=").Lit(""),
					).Else().Block(
						jen.Id("m").Index(
							jen.Id("v"),
						).Op("=").Lit(""),
					),
				),
				jen.Case(jen.Index().Interface()).Block(
					jen.Commentf("Recursively apply."),
					jen.For(
						jen.List(
							jen.Id("_"),
							jen.Id("elem"),
						).Op(":=").Range().Id("v"),
					).Block(
						jen.Id("r").Op(":=").Id(toAliasMapFnName).Call(
							jen.Id("elem"),
						),
						jen.For(
							jen.List(
								jen.Id("k"),
								jen.Id("val"),
							).Op(":=").Range().Id("r"),
						).Block(
							jen.Id("m").Index(
								jen.Id("k"),
							).Op("=").Id("val"),
						),
					),
				),
				jen.Case(jen.Map(jen.String()).Interface()).Block(
					jen.Commentf("Map any aliases."),
					jen.For(
						jen.List(
							jen.Id("k"),
							jen.Id("val"),
						).Op(":=").Range().Id("v"),
					).Block(
						jen.Commentf("Only handle string aliases."),
						jen.Switch(jen.Id("conc").Op(":=").Id("val").Assert(jen.Type())).Block(
							jen.Case(jen.String()).Block(
								jen.Id("m").Index(
									jen.Id("k"),
								).Op("=").Id("conc"),
							),
						),
					),
				),
			),
			jen.Return(),
		},
		fmt.Sprintf("%s converts a JSONLD context into a map of vocabulary name to alias.", toAliasMapFnName))
}
