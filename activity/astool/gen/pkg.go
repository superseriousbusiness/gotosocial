package gen

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

// PackageManager manages the path and names of a package consisting of a public
// and a private portion.
type PackageManager struct {
	prefix  string
	root    string
	public  string
	private string
}

// NewPackageManager creates a package manager whose private implementation is
// in an "impl" subdirectory.
func NewPackageManager(prefix, root string) *PackageManager {
	pathPrefix := strings.Replace(prefix, string(os.PathSeparator), "/", -1)
	pathRoot := strings.Replace(root, string(os.PathSeparator), "/", -1)
	return &PackageManager{
		prefix:  pathPrefix,
		root:    pathRoot,
		public:  "",
		private: "impl",
	}
}

// PublicPackage returns the public package.
func (p *PackageManager) PublicPackage() Package {
	return p.toPackage(p.public, true)
}

// PrivatePackage returns the private package.
func (p *PackageManager) PrivatePackage() Package {
	return p.toPackage(p.private, false)
}

// Sub creates a PackageManager clone that manages a subdirectory.
func (p *PackageManager) Sub(name string) *PackageManager {
	s := name
	if len(p.root) > 0 {
		s = fmt.Sprintf("%s/%s", p.root, name)
	}
	return &PackageManager{
		prefix:  p.prefix,
		root:    s,
		public:  p.public,
		private: p.private,
	}
}

// SubPrivate creates a PackageManager clone where the private package is one
// subdirectory further.
func (p *PackageManager) SubPrivate(name string) *PackageManager {
	s := name
	if len(p.private) > 0 {
		s = fmt.Sprintf("%s/%s", p.private, name)
	}
	return &PackageManager{
		prefix:  p.prefix,
		root:    p.root,
		public:  p.public,
		private: s,
	}
}

// SubPublic creates a PackageManager clone where the public package is one
// subdirectory further.
func (p *PackageManager) SubPublic(name string) *PackageManager {
	s := name
	if len(p.public) > 0 {
		s = fmt.Sprintf("%s/%s", p.public, name)
	}
	return &PackageManager{
		prefix:  p.prefix,
		root:    p.root,
		public:  s,
		private: p.private,
	}
}

// toPackage returns the public or private Package managed by this
// PackageManager.
func (p *PackageManager) toPackage(suffix string, public bool) Package {
	var path string
	if len(p.root) > 0 && len(suffix) > 0 {
		path = strings.Join([]string{p.root, suffix}, "/")
	} else if len(suffix) > 0 {
		path = suffix
	} else if len(p.root) > 0 {
		path = p.root
	}
	s := strings.Split(path, "/")
	name := s[len(s)-1]
	return Package{
		prefix:   p.prefix,
		path:     path,
		name:     name,
		isPublic: public,
		parent:   p,
	}
}

// Package represents a Golang package.
type Package struct {
	prefix   string
	path     string
	name     string
	isPublic bool
	parent   *PackageManager
}

// Path is the GOPATH or module path to this package.
func (p Package) Path() string {
	path := p.prefix
	if len(p.path) > 0 {
		path += "/" + p.path
	}
	return path
}

// WriteDir obtains the relative directory this package should be written to,
// which may not be the same as Path. The calling code may not be running at the
// root of GOPATH.
func (p Package) WriteDir() string {
	return p.path
}

// Name returns the name of this package.
func (p Package) Name() string {
	return strings.Replace(p.name, "_", "", -1)
}

// IsPublic returns whether this package is intended to house public files for
// application developer use.
func (p Package) IsPublic() bool {
	return p.isPublic
}

// Parent returns the PackageManager managing this Package.
func (p Package) Parent() *PackageManager {
	return p.parent
}

const (
	managerInterfaceName           = "privateManager"
	setManagerFunctionName         = "SetManager"
	setTypePropertyConstructorName = "SetTypePropertyConstructor"
)

// TypePackageGenerator manages generating one-time files needed for types.
type TypePackageGenerator struct {
	typeVocabName string
	m             *ManagerGenerator
	typeProperty  *PropertyGenerator
}

// NewTypePackageGenerator creates a new TypePackageGenerator.
func NewTypePackageGenerator(
	typeVocabName string,
	m *ManagerGenerator,
	typeProperty *NonFunctionalPropertyGenerator) *TypePackageGenerator {
	return &TypePackageGenerator{
		typeVocabName: typeVocabName,
		m:             m,
		typeProperty:  &typeProperty.PropertyGenerator,
	}
}

// PublicDefinitions creates the public-facing code generated definitions needed
// once per package.
//
// Precondition: The passed-in generators are the complete set of type
// generators within a package. Must satisfy: len(tgs) > 0.
func (t *TypePackageGenerator) PublicDefinitions(tgs []*TypeGenerator) (typeI *codegen.Interface) {
	return publicTypeDefinitions(tgs)
}

// PrivateDefinitions creates the private code generated definitions needed once
// per package.
//
// Precondition: The passed-in generators are the complete set of type
// generators within a package. len(tgs) > 0
func (t *TypePackageGenerator) PrivateDefinitions(tgs []*TypeGenerator) ([]*jen.Statement, []*codegen.Interface, []*codegen.Function) {
	pkg := tgs[0].PrivatePackage()
	s, i, f := privateManagerHookDefinitions(pkg, tgs, nil)
	interfaces := []*codegen.Interface{i, ContextInterface(pkg)}
	cv, setCv := privateTypePropertyConstructor(pkg, toPublicConstructor(t.typeVocabName, t.m, t.typeProperty))
	return []*jen.Statement{s, cv}, interfaces, []*codegen.Function{f, setCv}
}

// PropertyPackageGenerator manages generating one-time files needed for
// properties.
type PropertyPackageGenerator struct{}

// NewPropertyPackageGenerator creates a new PropertyPackageGenerator.
func NewPropertyPackageGenerator() *PropertyPackageGenerator {
	return &PropertyPackageGenerator{}
}

// PrivateDefinitions creates the private code generated definitions needed once
// per package.
//
// Precondition: The passed-in generators are the complete set of type
// generators within a package. len(pgs) > 0
func (p *PropertyPackageGenerator) PrivateDefinitions(pgs []*PropertyGenerator) (*jen.Statement, *codegen.Interface, *codegen.Function) {
	return privateManagerHookDefinitions(pgs[0].GetPrivatePackage(), nil, pgs)
}

// PackageGenerator maanges generating one-time files needed for both type and
// property implementations.
type PackageGenerator struct {
	typeVocabName string
	m             *ManagerGenerator
	typeProperty  *PropertyGenerator
}

// NewPackageGenerator creates a new PackageGenerator.
func NewPackageGenerator(typeVocabName string, m *ManagerGenerator, typeProperty *NonFunctionalPropertyGenerator) *PackageGenerator {
	return &PackageGenerator{
		typeVocabName: typeVocabName,
		m:             m,
		typeProperty:  &typeProperty.PropertyGenerator,
	}
}

// InitDefinitions returns the root init function needed to inject proper global
// package-private variables needed at runtime. This is the dependency injection
// into the implementation.
func (t *PackageGenerator) InitDefinitions(pkg Package, tgs []*TypeGenerator, pgs []*PropertyGenerator) (globalManager *jen.Statement, init *codegen.Function) {
	return genInit(pkg, tgs, pgs, toPublicConstructor(t.typeVocabName, t.m, t.typeProperty))
}

// RootDefinitions creates functions needed at the root level of the package declarations.
func (t *PackageGenerator) RootDefinitions(vocabName string, tgs []*TypeGenerator, pgs []*PropertyGenerator) (typeCtors, propCtors, ext, disj, extBy, isA []*codegen.Function) {
	return rootDefinitions(vocabName, t.m, tgs, pgs)
}

// PublicDefinitions creates the public-facing code generated definitions needed
// once per package.
//
// Precondition: The passed-in generators are the complete set of type
// generators within a package.
func (t *PackageGenerator) PublicDefinitions(tgs []*TypeGenerator) *codegen.Interface {
	return publicTypeDefinitions(tgs)
}

// PrivateDefinitions creates the private code generated definitions needed once
// per package.
//
// Precondition: The passed-in generators are the complete set of type
// generators within a package. One of tgs or pgs has at least one value.
func (t *PackageGenerator) PrivateDefinitions(tgs []*TypeGenerator, pgs []*PropertyGenerator) ([]*jen.Statement, []*codegen.Interface, []*codegen.Function) {
	var pkg Package
	if len(tgs) > 0 {
		pkg = tgs[0].PrivatePackage()
	} else {
		pkg = pgs[0].GetPrivatePackage()
	}
	s, i, f := privateManagerHookDefinitions(pkg, tgs, pgs)
	interfaces := []*codegen.Interface{i, ContextInterface(pkg)}
	cv, setCv := privateTypePropertyConstructor(pkg, toPublicConstructor(t.typeVocabName, t.m, t.typeProperty))
	return []*jen.Statement{s, cv}, interfaces, []*codegen.Function{f, setCv}
}

// privateTypePropertyConstructor creates common code needed by types to hook
// the type property constructor into this package at init time without
// statically linking to a specific implementation.
func privateTypePropertyConstructor(pkg Package, typePropertyConstructor *codegen.Function) (ctrVar *jen.Statement, setCtrVar *codegen.Function) {
	sig := typePropertyConstructor.ToFunctionSignature().Signature()
	ctrVar = jen.Var().Id(typePropertyConstructorName()).Add(sig)
	setCtrVar = codegen.NewCommentedFunction(
		pkg.Path(),
		setTypePropertyConstructorName,
		[]jen.Code{
			jen.Id("f").Add(sig),
		},
		/*ret=*/ nil,
		[]jen.Code{
			jen.Id(typePropertyConstructorName()).Op("=").Id("f"),
		},
		fmt.Sprintf("%s sets the \"type\" property's constructor in the package-global variable. For internal use only, do not use as part of Application behavior. Must be called at golang init time. Permits ActivityStreams types to correctly set their \"type\" property at construction time, so users don't have to remember to do so each time. It is dependency injected so other go-fed compatible implementations could inject their own type.", setTypePropertyConstructorName))
	return
}

// privateManagerHookDefinitions creates common code needed by types and
// properties to properly hook in the manager at initialization time.
func privateManagerHookDefinitions(pkg Package, tgs []*TypeGenerator, pgs []*PropertyGenerator) (mgrVar *jen.Statement, mgrI *codegen.Interface, setMgrFn *codegen.Function) {
	fnsMap := make(map[string]codegen.FunctionSignature)
	for _, tg := range tgs {
		for _, m := range tg.getAllManagerMethods() {
			v := m.ToFunctionSignature()
			fnsMap[v.Name] = v
		}
	}
	for _, pg := range pgs {
		for _, m := range pg.getAllManagerMethods() {
			v := m.ToFunctionSignature()
			fnsMap[v.Name] = v
		}
	}
	var fns []codegen.FunctionSignature
	for _, v := range fnsMap {
		fns = append(fns, v)
	}
	path := pkg.Path()
	return jen.Var().Id(managerInitName()).Id(managerInterfaceName),
		codegen.NewInterface(path,
			managerInterfaceName,
			fns,
			fmt.Sprintf("%s abstracts the code-generated manager that provides access to concrete implementations.", managerInterfaceName)),
		codegen.NewCommentedFunction(path,
			setManagerFunctionName,
			[]jen.Code{
				jen.Id("m").Id(managerInterfaceName),
			},
			/*ret=*/ nil,
			[]jen.Code{
				jen.Id(managerInitName()).Op("=").Id("m"),
			},
			fmt.Sprintf("%s sets the manager package-global variable. For internal use only, do not use as part of Application behavior. Must be called at golang init time.", setManagerFunctionName))
}

// publicTypeDefinitions creates common types needed by types for their public
// package.
//
// Requires tgs to not be empty.
func publicTypeDefinitions(tgs []*TypeGenerator) (typeI *codegen.Interface) {
	return TypeInterface(tgs[0].PublicPackage())
}

// rootDefinitions creates common functions needed at the root level of the
// package declarations.
func rootDefinitions(vocabName string, m *ManagerGenerator, tgs []*TypeGenerator, pgs []*PropertyGenerator) (typeCtors, propCtors, ext, disj, extBy, isA []*codegen.Function) {
	// Type constructors
	for _, tg := range tgs {
		typeCtors = append(typeCtors, codegen.NewCommentedFunction(
			m.pkg.Path(),
			fmt.Sprintf("New%s%s", vocabName, tg.TypeName()),
			/*params=*/ nil,
			[]jen.Code{jen.Qual(tg.PublicPackage().Path(), tg.InterfaceName())},
			[]jen.Code{
				jen.Return(
					tg.constructorFn().Call(),
				),
			},
			fmt.Sprintf("New%s%s creates a new %s", vocabName, tg.TypeName(), tg.InterfaceName())))
	}
	// Property Constructors
	for _, pg := range pgs {
		propCtors = append(propCtors, toPublicConstructor(vocabName, m, pg))
	}
	// Is
	for _, tg := range tgs {
		f := tg.isATypeDefinition()
		name := fmt.Sprintf("%s%s%s", isAMethod, vocabName, tg.TypeName())
		isA = append(isA, codegen.NewCommentedFunction(
			m.pkg.Path(),
			name,
			[]jen.Code{jen.Id("other").Qual(tg.PublicPackage().Path(), typeInterfaceName)},
			[]jen.Code{jen.Bool()},
			[]jen.Code{
				jen.Return(
					f.Call(jen.Id("other")),
				),
			},
			fmt.Sprintf("%s returns true if the other provided type is the %s type or extends from the %s type.", name, tg.TypeName(), tg.TypeName())))
	}
	// Extends
	for _, tg := range tgs {
		f, _ := tg.extendsDefinition()
		name := fmt.Sprintf("%s%s", vocabName, f.Name())
		ext = append(ext, codegen.NewCommentedFunction(
			m.pkg.Path(),
			name,
			[]jen.Code{jen.Id("other").Qual(tg.PublicPackage().Path(), typeInterfaceName)},
			[]jen.Code{jen.Bool()},
			[]jen.Code{
				jen.Return(
					f.Call(jen.Id("other")),
				),
			},
			fmt.Sprintf("%s returns true if %s extends from the other's type.", name, tg.TypeName())))
	}
	// DisjointWith
	for _, tg := range tgs {
		f := tg.disjointWithDefinition()
		name := fmt.Sprintf("%s%s", vocabName, f.Name())
		disj = append(disj, codegen.NewCommentedFunction(
			m.pkg.Path(),
			name,
			[]jen.Code{jen.Id("other").Qual(tg.PublicPackage().Path(), typeInterfaceName)},
			[]jen.Code{jen.Bool()},
			[]jen.Code{
				jen.Return(
					f.Call(jen.Id("other")),
				),
			},
			fmt.Sprintf("%s returns true if %s is disjoint with the other's type.", name, tg.TypeName())))
	}
	// ExtendedBy
	for _, tg := range tgs {
		f := tg.extendedByDefinition()
		name := fmt.Sprintf("%s%s", vocabName, f.Name())
		extBy = append(extBy, codegen.NewCommentedFunction(
			m.pkg.Path(),
			name,
			[]jen.Code{jen.Id("other").Qual(tg.PublicPackage().Path(), typeInterfaceName)},
			[]jen.Code{jen.Bool()},
			[]jen.Code{
				jen.Return(
					f.Call(jen.Id("other")),
				),
			},
			fmt.Sprintf("%s returns true if the other's type extends from %s. Note that it returns false if the types are the same; see the %q variant instead.", name, tg.TypeName(), isAMethod)))
	}
	return
}

// init generates the code that implements the init calls per-type and
// per-property package, so that the Manager is injected at runtime.
func genInit(pkg Package,
	tgs []*TypeGenerator,
	pgs []*PropertyGenerator,
	typePropertyConstructor *codegen.Function) (globalManager *jen.Statement, init *codegen.Function) {
	// manager dependency injection inits
	globalManager = jen.Var().Id(managerInitName()).Op("*").Qual(pkg.Path(), managerName)
	callInitsMap := make(map[string]jen.Code, len(tgs)+len(pgs))
	callInitsSlice := make([]string, 0, len(tgs)+len(pgs))
	for _, tg := range tgs {
		key := tg.PrivatePackage().Path()
		callInitsMap[key] = jen.Qual(tg.PrivatePackage().Path(), setManagerFunctionName).Call(
			jen.Qual(pkg.Path(), managerInitName()),
		)
		callInitsSlice = append(callInitsSlice, key)
	}
	for _, pg := range pgs {
		key := pg.GetPrivatePackage().Path()
		callInitsMap[key] = jen.Qual(pg.GetPrivatePackage().Path(), setManagerFunctionName).Call(
			jen.Qual(pkg.Path(), managerInitName()),
		)
		callInitsSlice = append(callInitsSlice, key)
	}
	sort.Strings(callInitsSlice)
	callInits := make([]jen.Code, 0, len(callInitsSlice))
	for _, c := range callInitsSlice {
		callInits = append(callInits, callInitsMap[c])
	}
	// type property constructor injection inits.
	// Resets the inits map and slice from above, to
	// keep appending to the callInits result.
	callInitsMap = make(map[string]jen.Code, len(tgs))
	callInitsSlice = make([]string, 0, len(tgs))
	for _, tg := range tgs {
		key := tg.PrivatePackage().Path()
		callInitsMap[key] = jen.Qual(tg.PrivatePackage().Path(), setTypePropertyConstructorName).Call(
			typePropertyConstructor.QualifiedName(),
		)
		callInitsSlice = append(callInitsSlice, key)
	}
	sort.Strings(callInitsSlice)
	for _, c := range callInitsSlice {
		callInits = append(callInits, callInitsMap[c])
	}
	init = codegen.NewCommentedFunction(
		pkg.Path(),
		"init",
		/*params=*/ nil,
		/*ret=*/ nil,
		append([]jen.Code{
			jen.Qual(pkg.Path(), managerInitName()).Op("=").Op("&").Qual(pkg.Path(), managerName).Values(),
		}, callInits...),
		fmt.Sprintf("init handles the 'magic' of creating a %s and dependency-injecting it into every other code-generated package. This gives the implementations access to create any type needed to deserialize, without relying on the other specific concrete implementations. In order to replace a go-fed created type with your own, be sure to have the manager call your own implementation's deserialize functions instead of the built-in type. Finally, each implementation views the %s as an interface with only a subset of funcitons available. This means this %s implements the union of those interfaces.", managerName, managerName, managerName))
	return
}

// toPublicConstructor creates a public constructor function for the given
// property, vocab name, and manager.
func toPublicConstructor(vocabName string, m *ManagerGenerator, pg *PropertyGenerator) *codegen.Function {
	return codegen.NewCommentedFunction(
		m.pkg.Path(),
		fmt.Sprintf("New%s%sProperty", vocabName, strings.Title(pg.PropertyName())),
		/*params=*/ nil,
		[]jen.Code{jen.Qual(pg.GetPublicPackage().Path(), pg.InterfaceName())},
		[]jen.Code{
			jen.Return(
				pg.ConstructorFn().Call(),
			),
		},
		fmt.Sprintf("New%s%s creates a new %s", vocabName, pg.StructName(), pg.InterfaceName()))
}
