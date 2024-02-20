package gen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

const (
	managerName        = "Manager"
	managerInitVarName = "mgr"
)

// managerInitName returns the package variable name for the manager.
func managerInitName() string {
	return managerInitVarName
}

// Generates the ActivityStreamManager that handles the creation of
// ActivityStream Core, Extended, and any extension types.
//
// This is implicitly used by Application code, but Application code usually
// won't need to manually use this Manager.
//
// This also provides interfaces to break the recursive/cyclic dependencies
// between properties and types. The previous version of this tool did not
// attempt to solve this problem, and instead just created one big and bloated
// library in order to avoid having to break the dependence. This version of
// the tool instead will generate interfaces for all of the required types.
//
// This means that developers will only ever need to interact with these
// interfaces, and could switch out using this implementation for another one of
// their own choosing.
//
// Also note that the manager links against all the implementations to generate
// a comprehensive registry. So while individual properties and types are able
// to be compiled separately, this generated output will link against all of
// these libraries.
//
// TODO: Improve the code generation to examine specific Golang code to
// determine which types to actually generate, and prune the unneeded types.
// This would cut down on the bloat on a per-program basis.
type ManagerGenerator struct {
	pkg Package
	tg  []*TypeGenerator
	fp  []*FunctionalPropertyGenerator
	nfp []*NonFunctionalPropertyGenerator
	// Constructed at creation time. These rely on pointer stability,
	// which should happen as none of these generators are treated as
	// values.
	tgManagedMethods  map[*TypeGenerator]*managedMethods
	fpManagedMethods  map[*FunctionalPropertyGenerator]*managedMethods
	nfpManagedMethods map[*NonFunctionalPropertyGenerator]*managedMethods
}

// managedMethods caches the specific methods and interfaces mapped to specific
// properties and types.
type managedMethods struct {
	deserializor *codegen.Method
}

// NewManagerGenerator creates a new manager system.
//
// This generator should be constructed in the third pass, after types and
// property generators are all constructed.
func NewManagerGenerator(pkg Package,
	tg []*TypeGenerator,
	fp []*FunctionalPropertyGenerator,
	nfp []*NonFunctionalPropertyGenerator) (*ManagerGenerator, error) {
	mg := &ManagerGenerator{
		pkg:               pkg,
		tg:                tg,
		fp:                fp,
		nfp:               nfp,
		tgManagedMethods:  make(map[*TypeGenerator]*managedMethods, len(tg)),
		fpManagedMethods:  make(map[*FunctionalPropertyGenerator]*managedMethods, len(fp)),
		nfpManagedMethods: make(map[*NonFunctionalPropertyGenerator]*managedMethods, len(nfp)),
	}
	// Pass 1: Get all deserializor-like methods created. Further passes may
	// rely on already having this data available in the manager.
	for _, t := range tg {
		mg.tgManagedMethods[t] = &managedMethods{
			deserializor: mg.createDeserializationMethodForType(t),
		}
	}
	for _, p := range fp {
		mg.fpManagedMethods[p] = &managedMethods{
			deserializor: mg.createDeserializationMethodForFuncProperty(p),
		}
	}
	for _, p := range nfp {
		mg.nfpManagedMethods[p] = &managedMethods{
			deserializor: mg.createDeserializationMethodForNonFuncProperty(p),
		}
	}
	// Pass 2: Inform the type of this ManagerGenerator so that it can keep
	// all of its bookkeeping straight.
	for _, t := range tg {
		if e := t.apply(mg); e != nil {
			return nil, e
		}
	}
	return mg, nil
}

// getDeserializationMethodForType obtains the deserialization method for a
// type.
func (m *ManagerGenerator) getDeserializationMethodForType(t *TypeGenerator) *codegen.Method {
	return m.tgManagedMethods[t].deserializor
}

// getDeserializationMethodForProperty obtains the deserialization method for a
// property regardless whether it is functional or non-functional.
func (m *ManagerGenerator) getDeserializationMethodForProperty(p Property) *codegen.Method {
	switch v := p.(type) {
	case *FunctionalPropertyGenerator:
		return m.fpManagedMethods[v].deserializor
	case *NonFunctionalPropertyGenerator:
		return m.nfpManagedMethods[v].deserializor
	default:
		panic("unknown property type")
	}
}

// Definition creates a manager implementation that works with the interface
// types required by the other PropertyGenerators and TypeGenerators for
// serializing and deserializing.
//
// Applications will implicitly use this manager and be isolated from the
// underlying specific go-fed implementation. If another alternative to go-fed
// were to be created, it could target those interfaces and be a drop-in
// replacement for an application.
//
// It is necessary to have this to acheive isolation without cyclic
// dependencies: types and properties can each belong in their own package (if
// desired) to minimize binary bloat.
func (m *ManagerGenerator) Definition() *codegen.Struct {
	var methods []*codegen.Method
	for _, tg := range m.tgManagedMethods {
		methods = append(methods, tg.deserializor)
	}
	for _, fp := range m.fpManagedMethods {
		methods = append(methods, fp.deserializor)
	}
	for _, nfp := range m.nfpManagedMethods {
		methods = append(methods, nfp.deserializor)
	}
	s := codegen.NewStruct(
		fmt.Sprintf("%s manages interface types and deserializations for use by generated code. Application code implicitly uses this manager at run-time to create concrete implementations of the interfaces.", managerName),
		managerName,
		methods,
		/*functions=*/ nil,
		/*members=*/ nil)
	return s
}

// createDeserializationMethodForType creates a new deserialization method for
// a type.
func (m *ManagerGenerator) createDeserializationMethodForType(tg *TypeGenerator) *codegen.Method {
	return m.createDeserializationMethod(
		tg.deserializationFnName(),
		tg.PublicPackage(),
		tg.PrivatePackage(),
		tg.InterfaceName(),
		tg.VocabName())
}

// createDeserializationMethodForFuncProperty creates a new deserialization
// method for a functional property.
func (m *ManagerGenerator) createDeserializationMethodForFuncProperty(fp *FunctionalPropertyGenerator) *codegen.Method {
	return m.createDeserializationMethod(
		fp.DeserializeFnName(),
		fp.GetPublicPackage(),
		fp.GetPrivatePackage(),
		fp.InterfaceName(),
		fp.VocabName())
}

// createDeserializationMethodForNonFuncProperty creates a new deserialization
// method for a non-functional property.
func (m *ManagerGenerator) createDeserializationMethodForNonFuncProperty(nfp *NonFunctionalPropertyGenerator) *codegen.Method {
	return m.createDeserializationMethod(
		nfp.DeserializeFnName(),
		nfp.GetPublicPackage(),
		nfp.GetPrivatePackage(),
		nfp.InterfaceName(),
		nfp.VocabName())
}

// createDeserializationMethod returns a function
func (m *ManagerGenerator) createDeserializationMethod(deserName string, pubPkg, privPkg Package, interfaceName, vocabName string) *codegen.Method {
	name := fmt.Sprintf("%s%s", deserName, vocabName)
	return codegen.NewCommentedValueMethod(
		m.pkg.Path(),
		name,
		managerName,
		/*param=*/ nil,
		[]jen.Code{
			jen.Func().Params(
				jen.Map(jen.String()).Interface(),
				jen.Map(jen.String()).String(),
			).Params(
				jen.Qual(pubPkg.Path(), interfaceName),
				jen.Error(),
			),
		},
		[]jen.Code{
			jen.Return(
				jen.Func().Params(
					jen.Id("m").Map(jen.String()).Interface(),
					jen.Id("aliasMap").Map(jen.String()).String(),
				).Params(
					jen.Qual(pubPkg.Path(), interfaceName),
					jen.Error(),
				).Block(
					jen.List(
						jen.Id("i"),
						jen.Err(),
					).Op(":=").Qual(privPkg.Path(), deserName).Call(jen.Id("m"), jen.Id("aliasMap")),
					jen.If(
						jen.Id("i").Op("==").Nil(),
					).Block(
						jen.Return(jen.Nil(), jen.Err()),
					),
					jen.Return(jen.List(
						jen.Id("i"),
						jen.Err(),
					)),
				),
			),
		},
		fmt.Sprintf("%s returns the deserialization method for the %q non-functional property in the vocabulary %q", name, interfaceName, vocabName))
}
