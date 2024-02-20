package convert

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
	"github.com/superseriousbusiness/activity/astool/gen"
	"github.com/superseriousbusiness/activity/astool/rdf"
	"github.com/superseriousbusiness/activity/astool/rdf/xsd"
)

const (
	interfacePkg = "vocab"
)

// File is a code-generated file.
type File struct {
	// F is the code-generated contents of this file.
	F *jen.File
	// FileName is the name of this file to write.
	FileName string
	// Directory specifies the location to write this file.
	Directory string
}

// vocabulary is a set of code generators for the vocabulary.
type vocabulary struct {
	Name       string
	URI        *url.URL
	Values     map[string]*gen.Kind
	FProps     map[string]*gen.FunctionalPropertyGenerator
	NFProps    map[string]*gen.NonFunctionalPropertyGenerator
	Types      map[string]*gen.TypeGenerator
	Manager    *gen.ManagerGenerator
	References map[string]*vocabulary
}

// newVocabulary creates a vocabulary with maps already made.
func newVocabulary() vocabulary {
	return vocabulary{
		Values:     make(map[string]*gen.Kind),
		FProps:     make(map[string]*gen.FunctionalPropertyGenerator),
		NFProps:    make(map[string]*gen.NonFunctionalPropertyGenerator),
		Types:      make(map[string]*gen.TypeGenerator),
		References: make(map[string]*vocabulary),
	}
}

// allTypeArray converts all Types, including referenced Types, to an array.
func (v vocabulary) allTypeArray() []*gen.TypeGenerator {
	var typeArray []*gen.TypeGenerator
	for _, ref := range v.References {
		typeArray = append(typeArray, ref.typeArray()...)
	}
	typeArray = append(typeArray, v.typeArray()...)
	sort.Sort(sortableTypeGenerator(typeArray))
	return typeArray
}

// allPropArray converts all Properties, including referenced Properties, to an
// array.
func (v vocabulary) allPropArray() []*gen.PropertyGenerator {
	var propArray []*gen.PropertyGenerator
	for _, ref := range v.References {
		propArray = append(propArray, ref.propArray()...)
	}
	propArray = append(propArray, v.propArray()...)
	sort.Sort(sortablePropertyGenerator(propArray))
	return propArray
}

// allFuncPropArray converts all FProps, including referenced Properties, to an
// array.
func (v vocabulary) allFuncPropArray() []*gen.FunctionalPropertyGenerator {
	var funcPropArray []*gen.FunctionalPropertyGenerator
	for _, ref := range v.References {
		funcPropArray = append(funcPropArray, ref.funcPropArray()...)
	}
	funcPropArray = append(funcPropArray, v.funcPropArray()...)
	sort.Sort(sortableFuncPropertyGenerator(funcPropArray))
	return funcPropArray
}

// allNonFuncPropArray converts all NFProps, including referenced Properties, to
// an array.
func (v vocabulary) allNonFuncPropArray() []*gen.NonFunctionalPropertyGenerator {
	var nonFuncPropArray []*gen.NonFunctionalPropertyGenerator
	for _, ref := range v.References {
		nonFuncPropArray = append(nonFuncPropArray, ref.nonFuncPropArray()...)
	}
	nonFuncPropArray = append(nonFuncPropArray, v.nonFuncPropArray()...)
	sort.Sort(sortableNonFuncPropertyGenerator(nonFuncPropArray))
	return nonFuncPropArray
}

// typeArray converts Types to an array.
func (v vocabulary) typeArray() []*gen.TypeGenerator {
	tg := make([]*gen.TypeGenerator, 0, len(v.Types))
	for _, t := range v.Types {
		tg = append(tg, t)
	}
	sort.Sort(sortableTypeGenerator(tg))
	return tg
}

// propArray converts FProps and NFProps to a generic array.
func (v vocabulary) propArray() []*gen.PropertyGenerator {
	fp := make([]*gen.PropertyGenerator, 0, len(v.FProps)+len(v.NFProps))
	for _, f := range v.FProps {
		fp = append(fp, &f.PropertyGenerator)
	}
	for _, f := range v.NFProps {
		fp = append(fp, &f.PropertyGenerator)
	}
	sort.Sort(sortablePropertyGenerator(fp))
	return fp
}

// funcPropArray converts only FProps to an array.
func (v vocabulary) funcPropArray() []*gen.FunctionalPropertyGenerator {
	fp := make([]*gen.FunctionalPropertyGenerator, 0, len(v.FProps))
	for _, f := range v.FProps {
		fp = append(fp, f)
	}
	sort.Sort(sortableFuncPropertyGenerator(fp))
	return fp
}

// nonFuncPropArray converts NFProps to an array.
func (v vocabulary) nonFuncPropArray() []*gen.NonFunctionalPropertyGenerator {
	nfp := make([]*gen.NonFunctionalPropertyGenerator, 0, len(v.NFProps))
	for _, nf := range v.NFProps {
		nfp = append(nfp, nf)
	}
	sort.Sort(sortableNonFuncPropertyGenerator(nfp))
	return nfp
}

// rdfReferences properly accounts for HTTP and HTTPS lookups of specification
// URIs.
type mapReferences map[string]*vocabulary

// Get attempts to fetch a reference, returning an error if it cannot.
func (r mapReferences) Get(uri string) (*vocabulary, error) {
	http, https, err := rdf.ToHttpAndHttps(uri)
	if err != nil {
		return nil, err
	}
	if v, ok := r[http]; ok {
		return v, nil
	} else if v, ok := r[https]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("mapReferences does not have reference %s", uri)
	}
}

// rdfReferences properly accounts for HTTP and HTTPS lookups of specification
// URIs.
type rdfReferences map[string]*rdf.Vocabulary

// Get attempts to fetch a reference, returning an error if it cannot.
func (r rdfReferences) Get(uri string) (*rdf.Vocabulary, error) {
	http, https, err := rdf.ToHttpAndHttps(uri)
	if err != nil {
		return nil, err
	}
	if v, ok := r[http]; ok {
		return v, nil
	} else if v, ok := r[https]; ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("rdfReferences does not have reference %s", uri)
	}
}

// PackagePolicy governs what file directory structure to generate files in.
// Only affects types and properties in a vocabulary. Does not affect values.
type PackagePolicy int

const (
	// FlatUnderRoot puts all types and properties together for each
	// vocabulary.
	FlatUnderRoot PackagePolicy = iota
	// IndividualUnderRoot puts each type and each property into their own
	// package within each vocabulary.
	IndividualUnderRoot
)

// Converter is responsible for taking the intermediate RDF understanding of one
// or more ActivityStreams-based specifications and converting it into a series
// of code-generated files.
//
// It supports generating code in two different styles, which greatly affects
// the package imports that application developers will use in their
// applications. While both are available, this results in code that may not be
// able to be built on some systems due to the compilation memory requirements.
// Therefore, final tooling is ecouraged to pick one and only one to use as
// there is no need for another dimension of code fragmentation. These styles
// are dicatated by the Converter's PackagePolicy.
//
// The generated code is separated into three locations: a "values" series of
// subdirectories, the subdirectories for the ActivityStream specification, and
// the "root" generated package.
//
// The "root" package is indended to be the sole package that all application
// developers use for non-interface types. It contains free-functions that aid
// in navigating the ActivityStreams hierarchy (such as extends and disjoints).
// It also contains a Resolver for deserialization or type dispatching.
//
// The specifications' generated code contains both interfaces and
// implementations. Developers' applications should only rely on the interfaces,
// which are used internally anyway.
type Converter struct {
	GenRoot       *gen.PackageManager
	PackagePolicy PackagePolicy
	// Properties stemming from JSONLD
	idProperty   *gen.FunctionalPropertyGenerator
	typeProperty *gen.NonFunctionalPropertyGenerator
}

// Convert turns a ParsedVocabulary into a set of code-generated files.
func (c *Converter) Convert(p *rdf.ParsedVocabulary) (f []*File, e error) {
	v := newVocabulary()
	done := make(map[string]bool)
	// Step 0: Create the "@id" and "@type" properties
	var xsdAnyUriKinds []gen.Kind
	var xsdStringKinds []gen.Kind
	var xsdAnyUri *url.URL
	var xsdString *url.URL
	xsdAnyUri, e = url.Parse(xsd.XmlSpec + xsd.AnyURISpec)
	if e != nil {
		return
	}
	xsdString, e = url.Parse(xsd.XmlSpec + xsd.StringSpec)
	if e != nil {
		return
	}
	xsdAnyUriKinds, e = c.propertyKinds(rdf.VocabularyProperty{
		Range: []rdf.VocabularyReference{{
			Name:  xsd.AnyURISpec,
			URI:   xsdAnyUri,
			Vocab: xsd.XmlSpec,
		}},
	}, v.Values, p.Vocab, p.References)
	if e != nil {
		return
	}
	xsdStringKinds, e = c.propertyKinds(rdf.VocabularyProperty{
		Range: []rdf.VocabularyReference{{
			Name:  xsd.StringSpec,
			URI:   xsdString,
			Vocab: xsd.XmlSpec,
		}},
	}, v.Values, p.Vocab, p.References)
	if e != nil {
		return
	}
	var idPkg, typePkg *gen.PackageManager
	idPkg, e = c.propertyPackageManager(rdf.VocabularyProperty{Name: "id"}, gen.JSONLDVocabName)
	if e != nil {
		return
	}
	c.idProperty, e = gen.NewIdProperty(idPkg, xsdAnyUriKinds[0])
	if e != nil {
		return
	}
	typePkg, e = c.propertyPackageManager(rdf.VocabularyProperty{Name: "type"}, gen.JSONLDVocabName)
	if e != nil {
		return
	}
	c.typeProperty, e = gen.NewTypeProperty(typePkg, xsdAnyUriKinds[0], xsdStringKinds[0])
	if e != nil {
		return
	}
	v.FProps[gen.JSONLDIdName] = c.idProperty
	v.NFProps[gen.JSONLDTypeName] = c.typeProperty
	// Step 1: Convert referenced specifications
	for i, vocabURI := range p.Order {
		refP := p.Clone()
		if i < len(p.Order)-1 {
			refP.Vocab = *refP.References[vocabURI]
			delete(refP.References, vocabURI)
		}

		var refV vocabulary
		refV, e = c.convertVocabulary(refP, v.References)
		if e != nil {
			return
		}
		// Order is full of references except the last, which is the
		// Vocab on ParsedVocabulary.
		if i < len(p.Order)-1 {
			v.References[vocabURI] = &refV
		} else {
			refV.References = v.References
			v = refV
		}
		done[vocabURI] = true
	}
	// Step 2: Intrinsically-known vocabularies won't appear in p.Order
	recurV, err := c.convertReferenceVocabularyRecursively(done, p, v.References)
	if err != nil {
		e = err
		return
	}
	for k, rv := range recurV {
		v.References[k] = rv
	}
	// Step 3: Create code-wide generators
	e = c.convertGenRoot(&v)
	if e != nil {
		return
	}
	// Step 4: Use the code generators to build the resulting code-generated
	// files.
	f, e = c.convertToFiles(v)
	return
}

// convertReferenceVocabularyRecursively will convert all references nested in
// all vocabularies and results in a flattened converted map.
func (c *Converter) convertReferenceVocabularyRecursively(skip map[string]bool, p *rdf.ParsedVocabulary, refs map[string]*vocabulary) (v map[string]*vocabulary, e error) {
	v = make(map[string]*vocabulary)
	for k := range p.References {
		if skip[k] {
			continue
		}
		refP := p.Clone()
		refP.Vocab = *refP.References[k]
		delete(refP.References, k)
		var refV vocabulary
		refV, e = c.convertVocabulary(refP, refs)
		if e != nil {
			return
		}
		v[k] = &refV
		// Recur
		var recurVocab map[string]*vocabulary
		recurVocab, e = c.convertReferenceVocabularyRecursively(skip, refP, refs)
		if e != nil {
			return
		}
		// Flatten
		for k, refVocab := range recurVocab {
			v[k] = refVocab
		}
	}
	return
}

// convertToFiles takes the generators for a vocabulary and maps them into a
// file structure.
func (c *Converter) convertToFiles(v vocabulary) (f []*File, e error) {
	pub := c.GenRoot.PublicPackage()
	// References
	for _, ref := range v.References {
		for _, v := range ref.Values {
			pkg := c.valuePackage(v)
			f = append(f, convertValue(pkg, v))
		}
		var files []*File
		files, e = c.toFiles(*ref)
		if e != nil {
			return
		}
		f = append(f, files...)
		files, e = c.rootFiles(pub, ref.Name, *ref, v.Manager)
		if e != nil {
			return
		}
		f = append(f, files...)
		pkgFiles, err := c.packageFiles(*ref, v.Manager)
		if err != nil {
			e = err
			return
		}
		f = append(f, pkgFiles...)
	}
	// JSONLD
	var files []*File
	files, e = c.jsonLDToFiles()
	if e != nil {
		return
	}
	f = append(f, files...)
	files, e = c.jsonLDRootFiles(pub, v.Manager)
	if e != nil {
		return
	}
	f = append(f, files...)
	// This vocabulary
	for _, v := range v.Values {
		pkg := c.valuePackage(v)
		f = append(f, convertValue(pkg, v))
	}
	files, e = c.toFiles(v)
	if e != nil {
		return
	}
	f = append(f, files...)
	files, e = c.rootFiles(pub, v.Name, v, v.Manager)
	if e != nil {
		return
	}
	f = append(f, files...)
	pkgFiles, err := c.packageFiles(v, v.Manager)
	if err != nil {
		e = err
		return
	}
	f = append(f, pkgFiles...)
	// Init file
	var file *File
	file, e = c.initFile(pub, v, v.Manager)
	if e != nil {
		return
	}
	f = append(f, file)
	// Manager
	jenFile := jen.NewFilePath(pub.Path())
	jenFile.Add(v.Manager.Definition().Definition())
	f = append(f, &File{
		F:         jenFile,
		FileName:  "gen_manager.go",
		Directory: pub.WriteDir(),
	})
	// JSONLD types
	var idFiles, typeFiles []*File
	idFiles, e = c.propertyPackageFiles(&c.idProperty.PropertyGenerator, gen.JSONLDVocabName)
	if e != nil {
		return
	}
	f = append(f, idFiles...)
	typeFiles, e = c.propertyPackageFiles(&c.typeProperty.PropertyGenerator, gen.JSONLDVocabName)
	if e != nil {
		return
	}
	f = append(f, typeFiles...)
	// Root Package Documentation
	rootDocFile := jen.NewFilePath(pub.Path())
	rootDocFile.PackageComment(gen.GenRootPackageComment(pub.Name()))
	f = append(f, &File{
		F:         rootDocFile,
		FileName:  "gen_doc.go",
		Directory: pub.WriteDir(),
	})
	// Constants
	files, e = c.constFiles(c.GenRoot.PublicPackage(), v.allTypeArray(), v.allPropArray())
	if e != nil {
		return
	}
	f = append(f, files...)
	// Resolvers
	files, e = c.resolverFiles(c.GenRoot.PublicPackage(), v.Manager, v)
	if e != nil {
		return
	}
	f = append(f, files...)
	return
}

// convertVocabulary works in a two-pass system: first converting all known
// properties, and then the types.
//
// Due to the fact that properties rely on the Kind abstraction, and both
// properties and types can be Kinds, this introduces tight coupling between
// the two so that callbacks can fill in missing links in data that isn't known
// beforehand (ex: how to serialize, deserialize, and compare types).
//
// This feels very hacky and could be decoupled using standard design patterns,
// but since there is no need, it isn't addressed now.
func (c *Converter) convertVocabulary(p *rdf.ParsedVocabulary, refs map[string]*vocabulary) (v vocabulary, e error) {
	v = newVocabulary()
	v.Name = p.Vocab.Name
	v.URI = p.Vocab.URI
	for k, val := range p.Vocab.Values {
		v.Values[k] = c.convertValue(p.Vocab.Name, val)
	}
	for k, prop := range p.Vocab.Properties {
		if prop.Functional {
			v.FProps[k], e = c.convertFunctionalProperty(prop, v.Values, p.Vocab, p.References, refs)
		} else {
			v.NFProps[k], e = c.convertNonFunctionalProperty(prop, v.Values, p.Vocab, p.References, refs)
		}
		if e != nil {
			return
		}
	}
	if c.typeProperty == nil {
		e = fmt.Errorf("convertVocabulary: could not find \"type\" property and its constructor")
		return
	}
	// Instead of building a dependency tree, naively keep iterating through
	// 'allTypes' until it is empty (good) or we get stuck (return error).
	allTypes := make([]rdf.VocabularyType, 0, len(p.Vocab.Types))
	for _, t := range p.Vocab.Types {
		allTypes = append(allTypes, t)
	}
	for {
		if len(allTypes) == 0 {
			break
		}
		stuck := true
		for i, t := range allTypes {
			if c.allExtendsAreIn(p.Vocab.Registry, t, v.Types, refs) {
				var tg *gen.TypeGenerator
				tg, e = c.convertType(t, p.Vocab, v.FProps, v.NFProps, v.Types, refs)
				if e != nil {
					return
				}
				v.Types[t.Name] = tg
				stuck = false
				// Delete the one we just did.
				allTypes[i] = allTypes[len(allTypes)-1]
				allTypes = allTypes[:len(allTypes)-1]
				break
			}
		}
		if stuck {
			e = fmt.Errorf("converting gen got stuck in dependency cycle")
			return
		}
	}
	return
}

// convertGenRoot creates code-wide code generators.
func (c *Converter) convertGenRoot(v *vocabulary) (e error) {
	v.Manager, e = gen.NewManagerGenerator(
		c.GenRoot.PublicPackage(),
		v.allTypeArray(),
		append(v.allFuncPropArray(), c.idProperty),
		append(v.allNonFuncPropArray(), c.typeProperty))
	return
}

// existingProperty attempts to find an existing Property in a referred
// vocabulary.
//
// Returns all nils if the property has been defined in a later vocabulary.
func (c *Converter) existingProperty(registry *rdf.RDFRegistry, r rdf.VocabularyReference, genRefs map[string]*vocabulary) (g gen.Property, e error) {
	var url string
	url, e = registry.ResolveAlias(r.Vocab)
	if e != nil {
		return
	}
	mapRef := mapReferences(genRefs)
	var refVocab *vocabulary
	refVocab, e = mapRef.Get(url)
	if e != nil {
		// Forwarded property -- a later specification added this
		// property to a type defined in an earlier specification.
		e = nil
		return
	}
	// Ugly.
	if p, ok := refVocab.FProps[r.Name]; !ok {
		if p, ok := refVocab.NFProps[r.Name]; !ok {
			e = fmt.Errorf("refVocab %s cannot find %s", r.Vocab, r.Name)
			return
		} else {
			g = p
			return
		}
	} else {
		g = p
		return
	}
}

// convertType turns the rdf.VocabularyType into a TypeGenerator.
//
// Precondition: The types it extends from and the properties it references are
// already converted into their applicable generators.
func (c *Converter) convertType(t rdf.VocabularyType,
	v rdf.Vocabulary,
	existingFProps map[string]*gen.FunctionalPropertyGenerator,
	existingNFProps map[string]*gen.NonFunctionalPropertyGenerator,
	existingTypes map[string]*gen.TypeGenerator,
	genRefs map[string]*vocabulary) (tg *gen.TypeGenerator, e error) {
	// Determine the gen package name
	var pm *gen.PackageManager
	pm, e = c.typePackageManager(t, v.Name)
	if e != nil {
		return
	}
	// Determine the properties for this type
	var p []gen.Property
	for _, prop := range t.Properties {
		if len(prop.Vocab) != 0 {
			var py gen.Property
			py, e = c.existingProperty(v.Registry, prop, genRefs)
			if e != nil {
				return
			}
			if py != nil {
				p = append(p, py)
			}
		} else {
			var property gen.Property
			var ok bool
			property, ok = existingFProps[prop.Name]
			if !ok {
				property, ok = existingNFProps[prop.Name]
				if !ok {
					e = fmt.Errorf("cannot find property with name: %s", prop.Name)
					return
				}
			}
			p = append(p, property)
		}
	}
	// Determine WithoutProperties for this type
	var wop []gen.Property
	for _, prop := range t.WithoutProperties {
		if len(prop.Vocab) != 0 {
			var py gen.Property
			py, e = c.existingProperty(v.Registry, prop, genRefs)
			if e != nil {
				return
			}
			if py != nil {
				wop = append(wop, py)
			}
		} else {
			var property gen.Property
			var ok bool
			property, ok = existingFProps[prop.Name]
			if !ok {
				property, ok = existingNFProps[prop.Name]
				if !ok {
					e = fmt.Errorf("cannot find property with name: %s", prop.Name)
					return
				}
			}
			wop = append(wop, property)
		}
	}
	// Special case: this type is actually typeless, so ignore the "type"
	// property.
	if t.IsTypeless() {
		wop = append(wop, c.typeProperty)
	}
	// Determine what this type extends
	var ext []*gen.TypeGenerator
	for _, ex := range t.Extends {
		if len(ex.Vocab) != 0 {
			var t *gen.TypeGenerator
			t, e = existingType(v.Registry, ex, genRefs)
			if e != nil {
				return
			}
			ext = append(ext, t)
		} else {
			ext = append(ext, existingTypes[ex.Name])
		}
	}
	// Apply disjoint if both sides are available because the TypeGenerator
	// does not know the entire vocabulary, so cannot do this lookup and
	// create this connection for us.
	var disjoint []*gen.TypeGenerator
	for _, disj := range t.DisjointWith {
		if len(disj.Vocab) != 0 {
			var t *gen.TypeGenerator
			t, e = existingType(v.Registry, disj, genRefs)
			if e != nil {
				return
			}
			disjoint = append(disjoint, t)
		} else if disjointType, ok := existingTypes[disj.Name]; ok {
			disjoint = append(disjoint, disjointType)
		}
	}
	// Pass in properties whose range is this type so it can build
	// references properly.
	//
	// Note that the Kinds container on properties contains both types and
	// values.
	name := c.convertTypeToName(t)
	var rangeProps []gen.Property
	for _, prop := range existingFProps {
		for _, kind := range prop.GetKinds() {
			// TODO: Rename "LowerName" since the type's name is
			// actually title case.
			if kind.Name.LowerName == name {
				rangeProps = append(rangeProps, prop)
			}
		}
	}
	for _, prop := range existingNFProps {
		for _, kind := range prop.GetKinds() {
			if kind.Name.LowerName == name {
				rangeProps = append(rangeProps, prop)
			}
		}
	}
	var examples []string
	for _, ex := range t.Examples {
		examples = append(examples, asComment(ex))
	}
	comment := t.Notes
	if len(examples) > 0 {
		comment = fmt.Sprintf("%s\n\n%s", comment, strings.Join(examples, "\n\n"))
	}
	// Always include the type and id JSONLD properties
	p = append(p, []gen.Property{c.typeProperty, c.idProperty}...)
	tg, e = gen.NewTypeGenerator(
		v.GetName(),
		v.URI,
		v.GetWellKnownAlias(),
		pm,
		name,
		comment,
		p,
		wop,
		rangeProps,
		ext,
		disjoint,
		t.IsTypeless())
	return
}

// convertFunctionalProperty converts an rdf.VocabularyProperty that is
// functional (can only have one value) into a FunctionalPropertyGenerator.
func (c *Converter) convertFunctionalProperty(p rdf.VocabularyProperty,
	kinds map[string]*gen.Kind,
	v rdf.Vocabulary,
	refs map[string]*rdf.Vocabulary,
	genRefs map[string]*vocabulary) (fp *gen.FunctionalPropertyGenerator, e error) {
	var k []gen.Kind
	k, e = c.propertyKinds(p, kinds, v, refs)
	if e != nil {
		return
	}
	var pm *gen.PackageManager
	pm, e = c.propertyPackageManager(p, v.Name)
	if e != nil {
		return
	}
	var examples []string
	for _, ex := range p.Examples {
		examples = append(examples, asComment(ex))
	}
	comment := p.Notes
	if len(examples) > 0 {
		comment = fmt.Sprintf("%s\n\n%s", comment, strings.Join(examples, "\n\n"))
	}
	fp, e = gen.NewFunctionalPropertyGenerator(
		v.GetName(),
		v.URI,
		v.GetWellKnownAlias(),
		pm,
		toIdentifier(p),
		comment,
		k,
		p.NaturalLanguageMap)
	if e != nil {
		return
	}
	e = backPopulateProperty(v.Registry, p, genRefs, fp)
	if e != nil {
		return
	}
	return
}

// convertNonFunctionalProperty converts an rdf.VocabularyProperty that is
// non-functional (can have multiple values) into a
// NonFunctionalPropertyGenerator.
func (c *Converter) convertNonFunctionalProperty(p rdf.VocabularyProperty,
	kinds map[string]*gen.Kind,
	v rdf.Vocabulary,
	refs map[string]*rdf.Vocabulary,
	genRefs map[string]*vocabulary) (nfp *gen.NonFunctionalPropertyGenerator, e error) {
	var k []gen.Kind
	k, e = c.propertyKinds(p, kinds, v, refs)
	if e != nil {
		return
	}
	var pm *gen.PackageManager
	pm, e = c.propertyPackageManager(p, v.Name)
	if e != nil {
		return
	}
	var examples []string
	for _, ex := range p.Examples {
		examples = append(examples, asComment(ex))
	}
	comment := p.Notes
	if len(examples) > 0 {
		comment = fmt.Sprintf("%s\n\n%s", comment, strings.Join(examples, "\n\n"))
	}
	nfp, e = gen.NewNonFunctionalPropertyGenerator(
		v.GetName(),
		v.URI,
		v.GetWellKnownAlias(),
		pm,
		toIdentifier(p),
		comment,
		k,
		p.NaturalLanguageMap)
	if e != nil {
		return
	}
	e = backPopulateProperty(v.Registry, p, genRefs, nfp)
	if e != nil {
		return
	}
	return
}

// convertValue turns a rdf.VocabularyValue into a Kind.
func (c *Converter) convertValue(vocabName string, v rdf.VocabularyValue) *gen.Kind {
	s := v.SerializeFn.CloneToPackage(c.vocabValuePackage(v).Path())
	d := v.DeserializeFn.CloneToPackage(c.vocabValuePackage(v).Path())
	l := v.LessFn.CloneToPackage(c.vocabValuePackage(v).Path())
	// Name must use toIdentifier for vocabValuePackage and valuePackage to
	// be the same.
	id := toIdentifier(v)
	return gen.NewKindForValue(id.LowerName,
		id.CamelName,
		vocabName,
		v.DefinitionType,
		v.IsNilable,
		v.IsURI,
		s,
		d,
		l)
}

// convertTypeToName makes a Titled version of the VocabularyType's name.
func (c *Converter) convertTypeToName(v rdf.VocabularyType) string {
	return strings.Title(v.Name)
}

// propertyKinds determines what Kind names are referenced by the
// rdf.VocabularyProperty, which may rely on the parsing registry for these
// particular files.
func (c *Converter) propertyKinds(v rdf.VocabularyProperty,
	kinds map[string]*gen.Kind,
	vocab rdf.Vocabulary,
	refs map[string]*rdf.Vocabulary) (k []gen.Kind, e error) {
	for _, r := range v.Range {
		if len(r.Vocab) == 0 {
			if kind, ok := kinds[r.Name]; !ok {
				// It is a Type of the vocabulary
				if t, ok := vocab.Types[r.Name]; !ok {
					e = fmt.Errorf("cannot find own kind with name %q", r.Name)
					return
				} else {
					id := toIdentifier(t)
					kt := gen.NewKindForType(id.LowerName, id.CamelName, vocab.Name)
					k = append(k, *kt)
				}
			} else {
				// It is a Value of the vocabulary
				k = append(k, *kind)
			}
		} else {
			var url string
			url, e = vocab.Registry.ResolveAlias(r.Vocab)
			if e != nil {
				return
			}
			rdfRef := rdfReferences(refs)
			var refVocab *rdf.Vocabulary
			refVocab, e = rdfRef.Get(url)
			if e != nil {
				return
			}
			if val, ok := refVocab.Values[r.Name]; !ok {
				// It is a Type of the vocabulary instead
				if t, ok := refVocab.Types[r.Name]; !ok {
					e = fmt.Errorf("cannot find kind with name %q in %s", r.Name, url)
					return
				} else {
					id := toIdentifier(t)
					kt := gen.NewKindForType(id.LowerName, id.CamelName, refVocab.Name)
					k = append(k, *kt)
				}
			} else {
				// It is a Value of the vocabulary
				k = append(k, *c.convertValue(refVocab.Name, val))
			}
		}
	}
	return
}

// getValueRoot returns the subdirectory that contains value types.
func (c *Converter) getValueRoot() *gen.PackageManager {
	return c.GenRoot.Sub("values")
}

// valuePackage returns the subpackage for a value Kind.
//
// It must match the result of vocabValuePackage. Therefore, convertTypeToKind
// and convertValue must also use toIdentifier.
func (c *Converter) valuePackage(v *gen.Kind) gen.Package {
	return c.getValueRoot().Sub(v.Name.LowerName).PublicPackage()
}

// vocabValuePackage returns the subpackage for a value Kind.
//
// It must match the result of valuePackage. Therefore, convertTypeToKind and
// convertValue must also use toIdentifier.
func (c *Converter) vocabValuePackage(v rdf.VocabularyValue) gen.Package {
	return c.getValueRoot().Sub(toIdentifier(v).LowerName).PublicPackage()
}

// typePackageManager returns a package manager for an individual type. It may
// be the same as other types depending on the code generation policy.
func (c *Converter) typePackageManager(v typeNamer, vocabName string) (pkg *gen.PackageManager, e error) {
	return c.packageManager("type_"+v.TypeName(), vocabName)
}

// propertyPackageManager returns a package manager for an individual property.
// It may be the same as other types depending on the code generation policy.
func (c *Converter) propertyPackageManager(v propertyNamer, vocabName string) (pkg *gen.PackageManager, e error) {
	return c.packageManager("property_"+v.PropertyName(), vocabName)
}

// packageManager applies the code generation package policy and returns a
// PackageManager applicable for that policy.
//
// The FlatUnderRoot policy puts all properties and types together in a single
// public and single private package.
//
// The IndividualUnderRoot policy puts each property and each type in their own
// package, and each of those packages has their own public and private
// subpackages.
func (c *Converter) packageManager(s, vocabName string) (pkg *gen.PackageManager, e error) {
	s = strings.ToLower(s)
	switch c.PackagePolicy {
	case FlatUnderRoot:
		pkg = c.GenRoot.SubPublic(interfacePkg).SubPrivate(strings.ToLower(vocabName))
	case IndividualUnderRoot:
		pkg = c.GenRoot.SubPublic(interfacePkg).SubPrivate(strings.ToLower(vocabName)).SubPrivate(s)
	default:
		e = fmt.Errorf("unrecognized PackagePolicy: %v", c.PackagePolicy)
	}
	return
}

// jsonLDRootFiles creates files that are applied for JSONLD.
//
// TODO: This function looks a lot like the next one (copy/paste). Deduplicate.
func (c *Converter) jsonLDRootFiles(pkg gen.Package, m *gen.ManagerGenerator) (f []*File, e error) {
	pg := gen.NewPackageGenerator(gen.JSONLDVocabName, m, c.typeProperty)
	_, propCtors, _, _, _, _ := pg.RootDefinitions(gen.JSONLDVocabName, []*gen.TypeGenerator{}, []*gen.PropertyGenerator{&c.typeProperty.PropertyGenerator, &c.idProperty.PropertyGenerator})
	lowerVocabName := strings.ToLower(gen.JSONLDVocabName)
	if file := funcsToFile(pkg, propCtors, fmt.Sprintf("gen_pkg_%s_property_constructors.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	return
}

// rootFiles creates files that are applied for all vocabularies. These files
// are the ones typically used by developers.
func (c *Converter) rootFiles(pkg gen.Package, vocabName string, v vocabulary, m *gen.ManagerGenerator) (f []*File, e error) {
	pg := gen.NewPackageGenerator(gen.JSONLDVocabName, m, c.typeProperty)
	typeCtors, propCtors, ext, disj, extBy, isA := pg.RootDefinitions(vocabName, v.typeArray(), v.propArray())
	lowerVocabName := strings.ToLower(vocabName)
	if file := funcsToFile(pkg, typeCtors, fmt.Sprintf("gen_pkg_%s_type_constructors.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	if file := funcsToFile(pkg, propCtors, fmt.Sprintf("gen_pkg_%s_property_constructors.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	if file := funcsToFile(pkg, ext, fmt.Sprintf("gen_pkg_%s_extends.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	if file := funcsToFile(pkg, disj, fmt.Sprintf("gen_pkg_%s_disjoint.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	if file := funcsToFile(pkg, extBy, fmt.Sprintf("gen_pkg_%s_extendedby.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	if file := funcsToFile(pkg, isA, fmt.Sprintf("gen_pkg_%s_isorextends.go", lowerVocabName)); file != nil {
		f = append(f, file)
	}
	return
}

// initFile creates the file with the init function that hooks together the
// runtime Manager.
func (c *Converter) initFile(pkg gen.Package, root vocabulary, m *gen.ManagerGenerator) (f *File, e error) {
	pg := gen.NewPackageGenerator(gen.JSONLDVocabName, m, c.typeProperty)
	globalVar, initFn := pg.InitDefinitions(pkg, root.allTypeArray(), root.allPropArray())
	initFile := jen.NewFilePath(pkg.Path())
	initFile.Add(globalVar).Line().Add(initFn.Definition()).Line()
	f = &File{
		F:         initFile,
		FileName:  "gen_init.go",
		Directory: pkg.WriteDir(),
	}
	return
}

// packageFiles generates package-level files necessary for types and properties
// within a single vocabulary.
//
// In the FlatUnderRoot policy, only one is needed in the public and private
// directories since all types and properties are co-located within the same
// package.
//
// In the IndividualUnderRoot policy, multiple are needed. Since each type and
// property are in their own package, one package file is needed in each of
// those types' and properties' subpackage.
func (c *Converter) packageFiles(v vocabulary, m *gen.ManagerGenerator) (f []*File, e error) {
	switch c.PackagePolicy {
	case FlatUnderRoot:
		pg := gen.NewPackageGenerator(gen.JSONLDVocabName, m, c.typeProperty)
		if tArr := v.typeArray(); len(tArr) > 0 {
			// Only need one for all types.
			pubI := pg.PublicDefinitions(tArr)
			// Public
			pub := v.typeArray()[0].PublicPackage()
			file := jen.NewFilePath(pub.Path())
			file.Add(pubI.Definition())
			f = append(f, &File{
				F:         file,
				FileName:  "gen_pkg.go",
				Directory: pub.WriteDir(),
			})
			// Public Package Documentation
			docFile := jen.NewFilePath(pub.Path())
			docFile.PackageComment(gen.VocabPackageComment(pub.Name(), v.Name))
			f = append(f, &File{
				F:         docFile,
				FileName:  "gen_doc.go",
				Directory: pub.WriteDir(),
			})
		}
		// Private
		if tArr, pArr := v.typeArray(), v.propArray(); len(tArr) > 0 || len(pArr) > 0 {
			s, i, fn := pg.PrivateDefinitions(tArr, pArr)
			var priv gen.Package
			if len(tArr) > 0 {
				priv = tArr[0].PrivatePackage()
			} else {
				priv = pArr[0].GetPrivatePackage()
			}
			file := jen.NewFilePath(priv.Path())
			for _, elem := range s {
				file.Add(elem).Line()
			}
			for _, elem := range i {
				file.Add(elem.Definition()).Line()
			}
			for _, elem := range fn {
				file.Add(elem.Definition()).Line()
			}
			f = append(f, &File{
				F:         file,
				FileName:  "gen_pkg.go",
				Directory: priv.WriteDir(),
			})
			// Private Package Documentation
			privDocFile := jen.NewFilePath(priv.Path())
			privDocFile.PackageComment(gen.PrivateFlatPackageComment(priv.Name(), v.Name))
			f = append(f, &File{
				F:         privDocFile,
				FileName:  "gen_doc.go",
				Directory: priv.WriteDir(),
			})
		}
	case IndividualUnderRoot:
		for _, tg := range v.Types {
			var file []*File
			file, e = c.typePackageFiles(tg, v.Name, m)
			if e != nil {
				return
			}
			f = append(f, file...)
		}
		for _, pg := range v.FProps {
			var file []*File
			file, e = c.propertyPackageFiles(&pg.PropertyGenerator, v.Name)
			if e != nil {
				return
			}
			f = append(f, file...)
		}
		for _, pg := range v.NFProps {
			var file []*File
			file, e = c.propertyPackageFiles(&pg.PropertyGenerator, v.Name)
			if e != nil {
				return
			}
			f = append(f, file...)
		}
	default:
		e = fmt.Errorf("unrecognized PackagePolicy: %v", c.PackagePolicy)
	}
	return
}

// typePackageFile creates the package-level files necessary for a type if it
// is being generated in its own package.
func (c *Converter) typePackageFiles(tg *gen.TypeGenerator, vocabName string, m *gen.ManagerGenerator) (f []*File, e error) {
	// Only need one for all types.
	tpg := gen.NewTypePackageGenerator(gen.JSONLDVocabName, m, c.typeProperty)
	pubI := tpg.PublicDefinitions([]*gen.TypeGenerator{tg})
	// Public
	pub := tg.PublicPackage()
	file := jen.NewFilePath(pub.Path())
	file.Add(pubI.Definition())
	f = append(f, &File{
		F:         file,
		FileName:  "gen_pkg.go",
		Directory: pub.WriteDir(),
	})
	// Public Package Documentation -- this may collide, but it's all the
	// same content.
	docFile := jen.NewFilePath(pub.Path())
	docFile.PackageComment(gen.VocabPackageComment(pub.Name(), vocabName))
	f = append(f, &File{
		F:         docFile,
		FileName:  "gen_doc.go",
		Directory: pub.WriteDir(),
	})
	// Private
	s, i, fn := tpg.PrivateDefinitions([]*gen.TypeGenerator{tg})
	priv := tg.PrivatePackage()
	file = jen.NewFilePath(priv.Path())
	for _, elem := range s {
		file.Add(elem).Line()
	}
	for _, elem := range i {
		file.Add(elem.Definition()).Line()
	}
	for _, elem := range fn {
		file.Add(elem.Definition()).Line()
	}
	f = append(f, &File{
		F:         file,
		FileName:  "gen_pkg.go",
		Directory: priv.WriteDir(),
	})
	// Private Package Documentation
	privDocFile := jen.NewFilePath(priv.Path())
	privDocFile.PackageComment(gen.PrivateIndividualTypePackageComment(priv.Name(), tg.TypeName()))
	f = append(f, &File{
		F:         privDocFile,
		FileName:  "gen_doc.go",
		Directory: priv.WriteDir(),
	})
	return
}

// propertyPackageFiles creates the package-level files necessary for a property
// if it is being generated in its own package.
func (c *Converter) propertyPackageFiles(pg *gen.PropertyGenerator, vocabName string) (f []*File, e error) {
	// Only need one for all types.
	ppg := gen.NewPropertyPackageGenerator()
	// Public Package Documentation -- this may collide, but it's all the
	// same content.
	pub := pg.GetPublicPackage()
	docFile := jen.NewFilePath(pub.Path())
	docFile.PackageComment(gen.VocabPackageComment(pub.Name(), vocabName))
	f = append(f, &File{
		F:         docFile,
		FileName:  "gen_doc.go",
		Directory: pub.WriteDir(),
	})
	// Private
	s, i, fn := ppg.PrivateDefinitions([]*gen.PropertyGenerator{pg})
	priv := pg.GetPrivatePackage()
	file := jen.NewFilePath(priv.Path())
	file.Add(
		s,
	).Line().Add(
		i.Definition(),
	).Line().Add(
		fn.Definition(),
	).Line()
	f = append(f, &File{
		F:         file,
		FileName:  "gen_pkg.go",
		Directory: priv.WriteDir(),
	})
	// Private Package Documentation
	privDocFile := jen.NewFilePath(priv.Path())
	privDocFile.PackageComment(gen.PrivateIndividualPropertyPackageComment(priv.Name(), pg.PropertyName()))
	f = append(f, &File{
		F:         privDocFile,
		FileName:  "gen_doc.go",
		Directory: priv.WriteDir(),
	})
	return
}

// resolverFiles creates the files necessary for the resolvers.
func (c *Converter) resolverFiles(pkg gen.Package, manGen *gen.ManagerGenerator, root vocabulary) (files []*File, e error) {
	rg := gen.NewResolverGenerator(root.allTypeArray(), manGen, pkg)
	jsonRes, typeRes, typePredRes, errDefs, fns, iFaces := rg.Definition()
	// Utils
	file := jen.NewFilePath(pkg.Path())
	for _, errDef := range errDefs {
		file.Add(errDef).Line()
	}
	for _, iFace := range iFaces {
		file.Add(iFace.Definition()).Line()
	}
	for _, fn := range fns {
		file.Add(fn.Definition()).Line()
	}
	files = append(files, &File{
		F:         file,
		FileName:  "gen_resolver_utils.go",
		Directory: pkg.WriteDir(),
	})
	// JSON resolver
	file = jen.NewFilePath(pkg.Path())
	file.Add(jsonRes.Definition())
	files = append(files, &File{
		F:         file,
		FileName:  "gen_json_resolver.go",
		Directory: pkg.WriteDir(),
	})
	// Type, not predicated
	file = jen.NewFilePath(pkg.Path())
	file.Add(typeRes.Definition())
	files = append(files, &File{
		F:         file,
		FileName:  "gen_type_resolver.go",
		Directory: pkg.WriteDir(),
	})
	// Type, Predicated
	file = jen.NewFilePath(pkg.Path())
	file.Add(typePredRes.Definition())
	files = append(files, &File{
		F:         file,
		FileName:  "gen_type_predicated_resolver.go",
		Directory: pkg.WriteDir(),
	})
	return
}

// constFiles creates the files for constants.
func (c *Converter) constFiles(pkg gen.Package, types []*gen.TypeGenerator, props []*gen.PropertyGenerator) (files []*File, e error) {
	consts := gen.GenerateConstants(types, props)
	file := jen.NewFilePath(pkg.Path())
	for _, elem := range consts {
		file.Add(elem).Line()
	}
	files = append(files, &File{
		F:         file,
		FileName:  "gen_consts.go",
		Directory: pkg.WriteDir(),
	})
	return files, e
}

// allExtendsAreIn determines if a VocabularyType's parents are all already
// converted to a TypeGenerator.
func (c *Converter) allExtendsAreIn(registry *rdf.RDFRegistry, t rdf.VocabularyType, v map[string]*gen.TypeGenerator, genRefs map[string]*vocabulary) bool {
	for _, e := range t.Extends {
		if len(e.Vocab) != 0 {
			_, err := existingType(registry, e, genRefs)
			return err == nil
		} else if _, ok := v[e.Name]; !ok {
			return false
		}
	}
	return true
}

// jsonLDToFiles converts id and type to files.
//
// TODO: This function and the next are a lot of shared code (copy/paste).
// Deduplicate it.
func (c *Converter) jsonLDToFiles() (f []*File, e error) {
	vName := strings.ToLower(gen.JSONLDVocabName)
	// type property
	var typePm *gen.PackageManager
	typePm, e = c.propertyPackageManager(rdf.VocabularyProperty{Name: "type"}, gen.JSONLDVocabName)
	if e != nil {
		return
	}
	// Implementation
	priv := typePm.PrivatePackage()
	file := jen.NewFilePath(priv.Path())
	s, t := c.typeProperty.Definitions()
	file.Add(s.Definition()).Line().Add(t.Definition())
	f = append(f, &File{
		F:         file,
		FileName:  fmt.Sprintf("gen_property_%s_%s.go", vName, c.typeProperty.PropertyName()),
		Directory: priv.WriteDir(),
	})
	// Interface
	pub := typePm.PublicPackage()
	file = jen.NewFilePath(pub.Path())
	for _, intf := range c.typeProperty.InterfaceDefinitions(typePm.PublicPackage()) {
		file.Add(intf.Definition()).Line()
	}
	f = append(f, &File{
		F:         file,
		FileName:  fmt.Sprintf("gen_property_%s_%s_interface.go", vName, c.typeProperty.PropertyName()),
		Directory: pub.WriteDir(),
	})

	// id property
	var idPm *gen.PackageManager
	idPm, e = c.propertyPackageManager(c.idProperty, gen.JSONLDVocabName)
	if e != nil {
		return
	}
	// Implementation
	priv = idPm.PrivatePackage()
	file = jen.NewFilePath(priv.Path())
	file.Add(c.idProperty.Definition().Definition())
	f = append(f, &File{
		F:         file,
		FileName:  fmt.Sprintf("gen_property_%s_%s.go", vName, c.idProperty.PropertyName()),
		Directory: priv.WriteDir(),
	})
	// Interface
	pub = idPm.PublicPackage()
	file = jen.NewFilePath(pub.Path())
	file.Add(c.idProperty.InterfaceDefinition(idPm.PublicPackage()).Definition())
	f = append(f, &File{
		F:         file,
		FileName:  fmt.Sprintf("gen_property_%s_%s_interface.go", vName, c.idProperty.PropertyName()),
		Directory: pub.WriteDir(),
	})
	return
}

// toFiles converts a vocabulary's types and properties to files.
func (c *Converter) toFiles(v vocabulary) (f []*File, e error) {
	vName := strings.ToLower(v.Name)
	for _, i := range v.FProps {
		var pm *gen.PackageManager
		pm, e = c.propertyPackageManager(i, v.Name)
		if e != nil {
			return
		}
		// Implementation
		priv := pm.PrivatePackage()
		file := jen.NewFilePath(priv.Path())
		file.Add(i.Definition().Definition())
		f = append(f, &File{
			F:         file,
			FileName:  fmt.Sprintf("gen_property_%s_%s.go", vName, i.PropertyName()),
			Directory: priv.WriteDir(),
		})
		// Interface
		pub := pm.PublicPackage()
		file = jen.NewFilePath(pub.Path())
		file.Add(i.InterfaceDefinition(pm.PublicPackage()).Definition())
		f = append(f, &File{
			F:         file,
			FileName:  fmt.Sprintf("gen_property_%s_%s_interface.go", vName, i.PropertyName()),
			Directory: pub.WriteDir(),
		})
	}
	// Non-Functional Properties
	for _, i := range v.NFProps {
		var pm *gen.PackageManager
		pm, e = c.propertyPackageManager(i, v.Name)
		if e != nil {
			return
		}
		// Implementation
		priv := pm.PrivatePackage()
		file := jen.NewFilePath(priv.Path())
		s, t := i.Definitions()
		file.Add(s.Definition()).Line().Add(t.Definition())
		f = append(f, &File{
			F:         file,
			FileName:  fmt.Sprintf("gen_property_%s_%s.go", vName, i.PropertyName()),
			Directory: priv.WriteDir(),
		})
		// Interface
		pub := pm.PublicPackage()
		file = jen.NewFilePath(pub.Path())
		for _, intf := range i.InterfaceDefinitions(pm.PublicPackage()) {
			file.Add(intf.Definition()).Line()
		}
		f = append(f, &File{
			F:         file,
			FileName:  fmt.Sprintf("gen_property_%s_%s_interface.go", vName, i.PropertyName()),
			Directory: pub.WriteDir(),
		})
	}
	// Types
	for _, i := range v.Types {
		var pm *gen.PackageManager
		pm, e = c.typePackageManager(i, v.Name)
		if e != nil {
			return
		}
		// Implementation
		priv := pm.PrivatePackage()
		file := jen.NewFilePath(priv.Path())
		file.Add(i.Definition().Definition())
		f = append(f, &File{
			F:         file,
			FileName:  fmt.Sprintf("gen_type_%s_%s.go", vName, strings.ToLower(i.TypeName())),
			Directory: priv.WriteDir(),
		})
		// Interface
		pub := pm.PublicPackage()
		file = jen.NewFilePath(pub.Path())
		file.Add(i.InterfaceDefinition(pm.PublicPackage()).Definition())
		f = append(f, &File{
			F:         file,
			FileName:  fmt.Sprintf("gen_type_%s_%s_interface.go", vName, strings.ToLower(i.TypeName())),
			Directory: pub.WriteDir(),
		})
	}
	return
}

// typeNamer bridges rdf.VocabularyType and gen.TypeGenerator.
type typeNamer interface {
	TypeName() string
}

var (
	_ typeNamer = &gen.TypeGenerator{}
	_ typeNamer = &rdf.VocabularyType{}
)

// propertyNamer bridges rdf.VocabularyProperty to the gen side of functional
// and non-functional property generators.
type propertyNamer interface {
	PropertyName() string
}

var (
	_ propertyNamer = &gen.FunctionalPropertyGenerator{}
	_ propertyNamer = &gen.NonFunctionalPropertyGenerator{}
	_ propertyNamer = &rdf.VocabularyProperty{}
)

// toIdentifier converts the name in the rdf package to a gen-compatible
// Identifier.
func toIdentifier(n rdf.NameGetter) gen.Identifier {
	return gen.Identifier{
		LowerName: n.GetName(),
		CamelName: strings.Title(n.GetName()),
	}
}

// convertValue converts a Kind value into a code-generated File.
func convertValue(pkg gen.Package, v *gen.Kind) *File {
	file := jen.NewFilePath(pkg.Path())
	file.Add(
		v.SerializeDef.Definition(),
	).Line().Add(
		v.DeserializeDef.Definition(),
	).Line().Add(
		v.LessDef.Definition())
	return &File{
		F:         file,
		FileName:  fmt.Sprintf("gen_%s.go", v.Name.LowerName),
		Directory: pkg.WriteDir(),
	}
}

// existingType attempts to find an existing Property in a referred vocabulary.
func existingType(registry *rdf.RDFRegistry, r rdf.VocabularyReference, genRefs map[string]*vocabulary) (g *gen.TypeGenerator, e error) {
	var url string
	url, e = registry.ResolveAlias(r.Vocab)
	if e != nil {
		return
	}
	mapRef := mapReferences(genRefs)
	var refVocab *vocabulary
	refVocab, e = mapRef.Get(url)
	if e != nil {
		return
	}
	if t, ok := refVocab.Types[r.Name]; !ok {
		e = fmt.Errorf("refVocab %s cannot find %s", r.Vocab, r.Name)
	} else {
		g = t
	}
	return
}

// funcsToFile is a helper converting an array of Functions into a single File
// in the specified Package.
func funcsToFile(pkg gen.Package, fns []*codegen.Function, filename string) *File {
	if len(fns) == 0 {
		return nil
	}
	file := jen.NewFilePath(pkg.Path())
	for _, fn := range fns {
		file.Add(fn.Definition()).Line()
	}
	return &File{
		F:         file,
		FileName:  filename,
		Directory: pkg.WriteDir(),
	}
}

// AsComment creates a Go-comment-compatible string out of an Example.
func asComment(v rdf.VocabularyExample) (s string) {
	if len(v.Name) > 0 && v.URI != nil {
		s = fmt.Sprintf("%s (%s):\n", v.Name, v.URI)
	} else if len(v.Name) > 0 {
		s = fmt.Sprintf("%s:\n", v.Name)
	} else if v.URI != nil {
		s = fmt.Sprintf("%s:\n", v.URI)
	}
	b, err := json.MarshalIndent(v.Example, "", "  ")
	if err != nil {
		panic(err)
	}
	ex := string(b)
	ex = strings.Replace(ex, "\n", "\n  ", -1)
	return fmt.Sprintf("%s  %s", s, ex)
}

// backPopulateProperty sets a new property generator onto existing types.
func backPopulateProperty(r *rdf.RDFRegistry, p rdf.VocabularyProperty, genRefs map[string]*vocabulary, fp gen.Property) (e error) {
	for _, dom := range p.Domain {
		// Within our own vocabulary -- skip
		if len(dom.Vocab) == 0 {
			continue
		}
		var url string
		url, e = r.ResolveAlias(dom.Vocab)
		if e != nil {
			return
		}
		mapRef := mapReferences(genRefs)
		var refVocab *vocabulary
		refVocab, e = mapRef.Get(url)
		if e != nil {
			// Continue -- this property should be added when the
			// type is being created for this extension.
			continue
		}
		if t, ok := refVocab.Types[dom.Name]; !ok {
			e = fmt.Errorf("back populate property with vocab %s cannot find %s", dom.Vocab, dom.Name)
			return
		} else {
			e = t.AddPropertyGenerator(fp)
			if e != nil {
				return
			}
		}
	}
	for _, ran := range p.Range {
		t, e := existingType(r, ran, genRefs)
		if e != nil {
			// Ignore -- it is either a part of this vocabulary, or
			// not a type (is a value instead).
			continue
		}
		t.AddRangeProperty(fp)
	}
	return
}
