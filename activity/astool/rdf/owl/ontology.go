package owl

import (
	"fmt"
	"strings"

	"github.com/superseriousbusiness/activity/astool/rdf"
)

const (
	owlSpec                = "http://www.w3.org/2002/07/owl#"
	membersSpec            = "members"
	disjointWithSpec       = "disjointWith"
	unionOfSpec            = "unionOf"
	importsSpec            = "imports"
	ontologySpec           = "Ontology"
	classSpec              = "Class"
	objectPropertySpec     = "ObjectProperty"
	functionalPropertySpec = "FunctionalProperty"
)

// OWLOntology is an Ontology for OWL2.
type OWLOntology struct {
	alias string
}

// SpecURI returns the URI of the OWL specification.
func (o *OWLOntology) SpecURI() string {
	return owlSpec
}

// Load this ontology without an alias.
func (o *OWLOntology) Load() ([]rdf.RDFNode, error) {
	return o.LoadAsAlias("")
}

// LoadAsAlias loads the ontology with the alias.
func (o *OWLOntology) LoadAsAlias(s string) ([]rdf.RDFNode, error) {
	o.alias = s
	return []rdf.RDFNode{
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     membersSpec,
			Delegate: &members{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     disjointWithSpec,
			Delegate: &disjointWith{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     unionOfSpec,
			Delegate: &unionOf{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     importsSpec,
			Delegate: &imports{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     ontologySpec,
			Delegate: &ontology{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     classSpec,
			Delegate: &class{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     objectPropertySpec,
			Delegate: &objectProperty{},
		},
		&rdf.AliasedDelegate{
			Spec:     owlSpec,
			Alias:    s,
			Name:     functionalPropertySpec,
			Delegate: &functionalProperty{},
		},
	}, nil
}

// LoadSpecificAsAlias loads a specific ontology definition as an alias.
func (o *OWLOntology) LoadSpecificAsAlias(alias, name string) ([]rdf.RDFNode, error) {
	switch name {
	case membersSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &members{},
			},
		}, nil
	case disjointWithSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &disjointWith{},
			},
		}, nil
	case unionOfSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &unionOf{},
			},
		}, nil
	case importsSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &imports{},
			},
		}, nil
	case ontologySpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &ontology{},
			},
		}, nil
	case classSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &class{},
			},
		}, nil
	case objectPropertySpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &objectProperty{},
			},
		}, nil
	case functionalPropertySpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &functionalProperty{},
			},
		}, nil
	}
	return nil, fmt.Errorf("owl ontology cannot find %q to alias to %q", name, alias)
}

// LoadElement allows loading nodes to enable contexts containing a container
// with an index.
func (o *OWLOntology) LoadElement(name string, payload map[string]interface{}) ([]rdf.RDFNode, error) {
	// First, detect if an idValue exists
	var idValue interface{}
	var ok bool
	idValue, ok = payload[rdf.IdSpec]
	if !ok {
		idValue, ok = payload[rdf.IdActivityStreamsSpec]
	}
	if !ok {
		return nil, nil
	}
	vStr, ok := idValue.(string)
	if !ok {
		return nil, nil
	}
	// Now that we have a string idValue, handle the import use case
	if !rdf.IsKeyApplicable(vStr, owlSpec, o.alias, importsSpec) {
		return nil, nil
	}
	node := &rdf.AliasedDelegate{
		Spec:  "",
		Alias: "",
		Name:  name,
		// Need to set Delegate, based on below logic
	}
	for k, v := range payload {
		if k == rdf.IdSpec || k == rdf.IdActivityStreamsSpec {
			continue
		}
		switch k {
		case rdf.ContainerSpec:
			container := &rdf.ContainerLD{}
			node.Delegate = container
			// Ugly, maybe move out to its own function when needed
			if cValStr, ok := v.(string); !ok {
				return nil, fmt.Errorf("unhandled owl import @container to non-string type: %T", v)
			} else {
				switch cValStr {
				case rdf.IndexSpec:
					container.ContainsNode = &rdf.IndexLD{}
				default:
					return nil, fmt.Errorf("unhandled owl import @container to string type %s", cValStr)
				}
			}
		default:
			return nil, fmt.Errorf("unhandled owl import use case: %s", k)
		}
	}
	return []rdf.RDFNode{node}, nil
}

// GetByName returns a bare node.
func (o *OWLOntology) GetByName(name string) (rdf.RDFNode, error) {
	name = strings.TrimPrefix(name, o.SpecURI())
	switch name {
	case membersSpec:
		return &members{}, nil
	case disjointWithSpec:
		return &disjointWith{}, nil
	case unionOfSpec:
		return &unionOf{}, nil
	case importsSpec:
		return &imports{}, nil
	case ontologySpec:
		return &ontology{}, nil
	case classSpec:
		return &class{}, nil
	case objectPropertySpec:
		return &objectProperty{}, nil
	case functionalPropertySpec:
		return &functionalProperty{}, nil
	}
	return nil, fmt.Errorf("owl ontology could not find node for name %s", name)
}

var _ rdf.RDFNode = &members{}

// members represents owl:members.
type members struct {
	pushed bool
}

// Enter does nothing but returns an error if the context is not reset.
func (m *members) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	if !ctx.IsReset() {
		ctx.Push()
		m.pushed = true
	}
	return true, nil
}

// Exit adds a Vocabulary Type, Property or Value to Result.
func (m *members) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	// Finish adding the Current item to the resulting vocabulary
	if ctx.Current == nil {
		return true, fmt.Errorf("owl members exiting with nil Current")
	}
	switch v := ctx.Current.(type) {
	case *rdf.VocabularyType:
		if err := ctx.Result.Vocab.SetType(ctx.Name, v); err != nil {
			return true, err
		}
	case *rdf.VocabularyProperty:
		if err := ctx.Result.Vocab.SetProperty(ctx.Name, v); err != nil {
			return true, err
		}
	case *rdf.VocabularyValue:
		if err := ctx.Result.Vocab.SetValue(ctx.Name, v); err != nil {
			return true, err
		}
	default:
		return true, fmt.Errorf("owl members exiting with unhandled type: %T", ctx.Current)
	}
	if m.pushed {
		ctx.Pop()
	}
	return true, nil
}

// Apply returns an error.
func (m *members) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl members cannot be applied")
}

var _ rdf.RDFNode = &disjointWith{}

// disjointWith represents owl:disjointWith.
type disjointWith struct{}

// Enter ensures the Current is a Type, then pushes a Reference.
func (d *disjointWith) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	// Push the Current type aside, to build a Reference.
	if ctx.Current == nil {
		return true, fmt.Errorf("owl disjointWith enter given a nil Current")
	} else if _, ok := ctx.Current.(*rdf.VocabularyType); !ok {
		return true, fmt.Errorf("owl disjointWith enter not given a *rdf.VocabularyType")
	}
	ctx.Push()
	ctx.Current = &rdf.VocabularyReference{}
	return true, nil
}

// Exit pops the Reference and adds it to the Type's DisjointWith.
func (d *disjointWith) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	// Pop the Reference, put into the type.
	ref, ok := ctx.Current.(*rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("owl disjointWith exit not given a *rdf.VocabularyReference")
	}
	ctx.Pop()
	vType, ok := ctx.Current.(*rdf.VocabularyType)
	if !ok {
		return true, fmt.Errorf("owl disjointWith exit not given a *rdf.VocabularyType")
	}
	vType.DisjointWith = append(vType.DisjointWith, *ref)
	return true, nil
}

// Apply returns an error.
func (d *disjointWith) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl disjointWith cannot be applied")
}

var _ rdf.RDFNode = &unionOf{}

// unionOf represents owl:unionOf.
type unionOf struct {
	entered bool
}

// Enter pushes a single Reference.
func (u *unionOf) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	u.entered = true
	ctx.Push()
	ctx.Current = &rdf.VocabularyReference{}
	return true, nil
}

// Exit pops a Reference and appends it to Current, which is a slice of
// References.
func (u *unionOf) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	u.entered = false
	if ctx.Current == nil {
		return true, fmt.Errorf("owl unionOf exit given nil Current")
	}
	i := ctx.Current
	ctx.Pop()
	ref, ok := i.(*rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("owl unionOf exit not given *rdf.VocabularyReference")
	}
	arr, ok := ctx.Current.([]rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("owl unionOf exit's previous Current not given []rdf.VocabularyReference")
	}
	ctx.Current = append(arr, *ref)
	return true, nil
}

// Apply will either apply a value onto a current Reference (if it was entered
// due to being a JSON array), or will append a new reference to Current (which
// is a slice of references).
func (u *unionOf) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	s, ok := value.(string)
	if !ok {
		return true, fmt.Errorf("owl unionOf apply given non-string value")
	}
	strs := rdf.SplitAlias(s)
	ref := &rdf.VocabularyReference{}
	if len(strs) == 1 {
		ref.Name = strs[0]
	} else if len(strs) == 2 {
		ref.Name = strs[1]
		ref.Vocab = strs[0]
	} else {
		return true, fmt.Errorf("owl unionOf apply bad SplitAlias")
	}
	if u.entered {
		in, ok := ctx.Current.(*rdf.VocabularyReference)
		if !ok {
			return true, fmt.Errorf("owl unionOf apply's Current not given *rdf.VocabularyReference: %T", ctx.Current)
		}
		in.Name = ref.Name
		in.Vocab = ref.Vocab
	} else {
		arr, ok := ctx.Current.([]rdf.VocabularyReference)
		if !ok {
			return true, fmt.Errorf("owl unionOf apply's Current not given []rdf.VocabularyReference: %T", ctx.Current)
		}
		ctx.Current = append(arr, *ref)
	}
	return true, nil
}

var _ rdf.RDFNode = &imports{}

// imports does nothing but returns errors. It should instead be handled by
// special cases in an Ontology's LoadElement.
//
// Overall, this is a pain to implement. If these errors are seen, then I am
// about to have a really not-fun day.
type imports struct{}

// Enter returns an error.
func (i *imports) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl imports cannot be entered")
}

// Exit returns an error.
func (i *imports) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl imports cannot be entered")
}

// Apply returns an error.
func (i *imports) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl imports cannot be entered")
}

var _ rdf.RDFNode = &ontology{}

// ontology does nothing.
type ontology struct{}

// Enter returns an error.
func (o *ontology) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl ontology cannot be entered")
}

// Exit returns an error.
func (o *ontology) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl ontology cannot be exited")
}

// Apply does nothing.
func (o *ontology) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Current = &ctx.Result.Vocab
	return true, nil
}

var _ rdf.RDFNode = &class{}

// class prepares a new Type on Current, unless Reference has already been
// prepared.
type class struct{}

// Enter returns an error.
func (c *class) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl class cannot be entered")
}

// Exit returns an error.
func (c *class) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl class cannot be exited")
}

// Apply sets a Type on Current, unless a Reference is already set.
func (c *class) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	// Prepare a new VocabularyType in the context, unless it is a
	// reference already prepared.
	if ctx.IsReset() {
		ctx.Current = &rdf.VocabularyType{}
	} else if _, ok := ctx.Current.(*rdf.VocabularyReference); ok {
		return true, nil
	} else if _, ok := ctx.Current.([]rdf.VocabularyReference); ok {
		return true, nil
	} else {
		return true, fmt.Errorf("owl class applied with non-reset ctx and not a vocab reference: %T", ctx.Current)
	}
	return true, nil
}

var _ rdf.RDFNode = &objectProperty{}

// objectProperty is owl:objectProperty
type objectProperty struct{}

// Enter returns an error.
func (o *objectProperty) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl objectProperty cannot be entered")
}

// Exit returns an error.
func (o *objectProperty) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl objectProperty cannot be exited")
}

// Apply sets Current to be a Property, unless it is already a Property.
func (o *objectProperty) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	// Prepare a new VocabularyProperty in the context. If one already
	// exists, skip.
	if _, ok := ctx.Current.(*rdf.VocabularyProperty); ok {
		return true, nil
	} else if !ctx.IsReset() {
		return true, fmt.Errorf("owl objectProperty applied with non-reset ParsingContext")
	}
	ctx.Current = &rdf.VocabularyProperty{}
	return true, nil
}

var _ rdf.RDFNode = &functionalProperty{}

// functionalProperty represents owl:functionalProperty
type functionalProperty struct{}

// Enter returns an error.
func (f *functionalProperty) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl functionalProperty cannot be entered")
}

// Exit returns an error.
func (f *functionalProperty) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("owl functionalProperty cannot be exited")
}

// Apply sets the Current Property's Functional to true.
//
// Returns an error if Current is not a Property.
func (f *functionalProperty) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	if ctx.Current == nil {
		return true, fmt.Errorf("owl functionalProperty given nil Current in context")
	}
	prop, ok := ctx.Current.(*rdf.VocabularyProperty)
	if !ok {
		return true, fmt.Errorf("owl functionalProperty given Current that is not *rdf.VocabularyProperty")
	}
	prop.Functional = true
	return true, nil
}
