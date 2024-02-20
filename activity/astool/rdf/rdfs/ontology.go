package rdfs

import (
	"fmt"
	"strings"

	"github.com/superseriousbusiness/activity/astool/rdf"
)

const (
	rdfsSpecURI       = "http://www.w3.org/2000/01/rdf-schema#"
	commentSpec       = "comment"
	domainSpec        = "domain"
	isDefinedBySpec   = "isDefinedBy"
	rangeSpec         = "range"
	subClassOfSpec    = "subClassOf"
	subPropertyOfSpec = "subPropertyOf"
)

// RDFSchemaOntology is the Ontology for rdfs.
type RDFSchemaOntology struct{}

// SpecURI returns the URI for the RDFS spec.
func (o *RDFSchemaOntology) SpecURI() string {
	return rdfsSpecURI
}

// Load this Ontology without an alias.
func (o *RDFSchemaOntology) Load() ([]rdf.RDFNode, error) {
	return o.LoadAsAlias("")
}

// LoadAsAlias loads this ontology with an alias.
func (o *RDFSchemaOntology) LoadAsAlias(s string) ([]rdf.RDFNode, error) {
	return []rdf.RDFNode{
		&rdf.AliasedDelegate{
			Spec:     rdfsSpecURI,
			Alias:    s,
			Name:     commentSpec,
			Delegate: &comment{},
		},
		&rdf.AliasedDelegate{
			Spec:     rdfsSpecURI,
			Alias:    s,
			Name:     domainSpec,
			Delegate: &domain{},
		},
		&rdf.AliasedDelegate{
			Spec:     rdfsSpecURI,
			Alias:    s,
			Name:     isDefinedBySpec,
			Delegate: &isDefinedBy{},
		},
		&rdf.AliasedDelegate{
			Spec:     rdfsSpecURI,
			Alias:    s,
			Name:     rangeSpec,
			Delegate: &ranges{},
		},
		&rdf.AliasedDelegate{
			Spec:     rdfsSpecURI,
			Alias:    s,
			Name:     subClassOfSpec,
			Delegate: &subClassOf{},
		},
		&rdf.AliasedDelegate{
			Spec:     rdfsSpecURI,
			Alias:    s,
			Name:     subPropertyOfSpec,
			Delegate: &subPropertyOf{},
		},
	}, nil
}

// LoadSpecificAsAlias loads a specific node as an alias.
func (o *RDFSchemaOntology) LoadSpecificAsAlias(alias, name string) ([]rdf.RDFNode, error) {
	switch name {
	case commentSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &comment{},
			},
		}, nil
	case domainSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &domain{},
			},
		}, nil
	case isDefinedBySpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &isDefinedBy{},
			},
		}, nil
	case rangeSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &ranges{},
			},
		}, nil
	case subClassOfSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &subClassOf{},
			},
		}, nil
	case subPropertyOfSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &subPropertyOf{},
			},
		}, nil
	}
	return nil, fmt.Errorf("rdfs ontology cannot find %q to alias to %q", name, alias)
}

// LoadElement does nothing.
func (o *RDFSchemaOntology) LoadElement(name string, payload map[string]interface{}) ([]rdf.RDFNode, error) {
	return nil, nil
}

// GetByName returns a bare node by name.
func (o *RDFSchemaOntology) GetByName(name string) (rdf.RDFNode, error) {
	name = strings.TrimPrefix(name, o.SpecURI())
	switch name {
	case commentSpec:
		return &comment{}, nil
	case domainSpec:
		return &domain{}, nil
	case isDefinedBySpec:
		return &isDefinedBy{}, nil
	case rangeSpec:
		return &ranges{}, nil
	case subClassOfSpec:
		return &subClassOf{}, nil
	case subPropertyOfSpec:
		return &subPropertyOf{}, nil
	}
	return nil, fmt.Errorf("rdfs ontology could not find node for name %s", name)
}

var _ rdf.RDFNode = &comment{}

// comment sets Notes on vocabulary items.
type comment struct{}

// Enter returns an error.
func (n *comment) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs comment cannot be entered")
}

// Exit returns an error.
func (n *comment) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs comment cannot be exited")
}

// Apply sets the string value on Current's Notes.
//
// Returns an error if value isn't a string or Current can't set Notes.
func (n *comment) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	note, ok := value.(string)
	if !ok {
		return true, fmt.Errorf("rdf comment not given string value")
	}
	if ctx.Current == nil {
		return true, fmt.Errorf("rdf comment given nil Current")
	}
	noteSetter, ok := ctx.Current.(rdf.NotesSetter)
	if !ok {
		return true, fmt.Errorf("rdf comment not given NotesSetter")
	}
	noteSetter.SetNotes(note)
	return true, nil
}

var _ rdf.RDFNode = &domain{}

// domain is rdfs:domain.
type domain struct{}

// Enter Pushes a Reference as Current.
func (d *domain) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Push()
	ctx.Current = make([]rdf.VocabularyReference, 0)
	return true, nil
}

// Exit Pops a slice of References and sets it on the parent Property.
//
// Returns an error if the popped item is not a slice of References, or if the
// Current after Popping is not a Property.
func (d *domain) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	i := ctx.Current
	ctx.Pop()
	vr, ok := i.([]rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("rdfs domain exit did not get []rdf.VocabularyReference")
	}
	vp, ok := ctx.Current.(*rdf.VocabularyProperty)
	if !ok {
		return true, fmt.Errorf("rdf domain exit Current is not *rdf.VocabularyProperty")
	}
	vp.Domain = append(vp.Domain, vr...)
	return true, nil
}

// Apply returns an error.
func (d *domain) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs domain cannot be applied")
}

var _ rdf.RDFNode = &isDefinedBy{}

// isDefinedBy is rdfs:isDefinedBy.
type isDefinedBy struct{}

// Enter returns an error.
func (i *isDefinedBy) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs isDefinedBy cannot be entered")
}

// Exit returns an error.
func (i *isDefinedBy) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs isDefinedBy cannot be exited")
}

// Apply sets the string value as Current's URI.
//
// Returns an error if value is not a string or if Current cannot have a URI
// set.
func (i *isDefinedBy) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	s, ok := value.(string)
	if !ok {
		return true, fmt.Errorf("rdfs isDefinedBy given non-string: %T", value)
	}
	u, ok := ctx.Current.(rdf.URISetter)
	if !ok {
		return true, fmt.Errorf("rdfs isDefinedBy Current is not rdf.URISetter: %T", ctx.Current)
	}
	return true, u.SetURI(s)
}

var _ rdf.RDFNode = &ranges{}

// ranges is rdfs:ranges.
type ranges struct{}

// Enter Pushes as a slice of References as Current.
func (r *ranges) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Push()
	ctx.Current = make([]rdf.VocabularyReference, 0)
	return true, nil
}

// Exit Pops a slice of References and sets it on the parent Property.
//
// Returns an error if the popped item is not a slice of references, or if the
// Current item after popping is not a Property.
func (r *ranges) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	i := ctx.Current
	ctx.Pop()
	vr, ok := i.([]rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("rdfs ranges exit did not get []rdf.VocabularyReference")
	}
	vp, ok := ctx.Current.(*rdf.VocabularyProperty)
	if !ok {
		return true, fmt.Errorf("rdf ranges exit Current is not *rdf.VocabularyProperty")
	}
	vp.Range = append(vp.Range, vr...)
	return true, nil
}

// Apply returns an error.
func (r *ranges) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs ranges cannot be applied")
}

var _ rdf.RDFNode = &subClassOf{}

// subClassOf implements rdfs:subClassOf.
type subClassOf struct{}

// Enter Pushes a Reference as Current.
func (s *subClassOf) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Push()
	ctx.Current = &rdf.VocabularyReference{}
	return true, nil
}

// Exit Pops a Reference and appends it to the parent Type's Extends.
//
// Returns an error if the popped item is not a reference, or if the Current
// item after popping is not a Type.
func (s *subClassOf) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	i := ctx.Current
	ctx.Pop()
	vr, ok := i.(*rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("rdfs subclassof exit did not get *rdf.VocabularyReference")
	}
	vt, ok := ctx.Current.(*rdf.VocabularyType)
	if !ok {
		return true, fmt.Errorf("rdf subclassof exit Current is not *rdf.VocabularyType")
	}
	vt.Extends = append(vt.Extends, *vr)
	return true, nil
}

// Apply returns an error.
func (s *subClassOf) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs subclassof cannot be applied")
}

var _ rdf.RDFNode = &subPropertyOf{}

// subPropertyOf is rdfs:subPropertyOf
type subPropertyOf struct{}

// Enter Pushes a Reference as Current.
func (s *subPropertyOf) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Push()
	ctx.Current = &rdf.VocabularyReference{}
	return true, nil
}

// Exit Pops a Reference and sets it as the parent property's SubpropertyOf.
//
// Returns an error if the popped item is not a Reference, or if after popping
// Current is not a Property.
func (s *subPropertyOf) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	i := ctx.Current
	ctx.Pop()
	vr, ok := i.(*rdf.VocabularyReference)
	if !ok {
		return true, fmt.Errorf("rdfs subpropertyof exit did not get *rdf.VocabularyReference")
	}
	vp, ok := ctx.Current.(*rdf.VocabularyProperty)
	if !ok {
		return true, fmt.Errorf("rdf subpropertyof exit Current is not *rdf.VocabularyProperty")
	}
	vp.SubpropertyOf = *vr
	return true, nil
}

// Apply returns an error.
func (s *subPropertyOf) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rdfs subpropertyof cannot be applied")
}
