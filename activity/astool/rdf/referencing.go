package rdf

import (
	"fmt"
)

var (
	_ Ontology = &ReferenceOntology{}
)

// ReferenceOntology wraps a previously-parsed spec so it can be made known to
// the registry.
type ReferenceOntology struct {
	v Vocabulary
}

// SpecURI returns the URI for this specification
func (r *ReferenceOntology) SpecURI() string {
	return r.v.URI.String()
}

// Load loads the ontology without an alias.
func (r *ReferenceOntology) Load() ([]RDFNode, error) {
	return r.LoadAsAlias("")
}

// LoadAsAlias loads the vocabulary ontology with an alias.
//
// Values cannot be loaded because their serialization and deserialization types
// are not known at runtime if not embedded in the go-fed tool. If the error is
// generated when running the tool, then file a bug so that the tool can
// properly "know" about this particular value and how to serialize and
// deserialize it properly.
func (r *ReferenceOntology) LoadAsAlias(s string) ([]RDFNode, error) {
	var nodes []RDFNode
	for name, t := range r.v.Types {
		nodes = append(nodes, &AliasedDelegate{
			Spec:     r.v.URI.String(),
			Alias:    s,
			Name:     name,
			Delegate: &typeReference{t: t, vocabName: r.SpecURI()},
		})
	}
	for name, p := range r.v.Properties {
		nodes = append(nodes, &AliasedDelegate{
			Spec:     r.v.URI.String(),
			Alias:    s,
			Name:     name,
			Delegate: &propertyReference{p: p, vocabName: r.SpecURI()},
		})
	}
	// Note: Values cannot be added this way as there's no way to detect
	// at runtime what the correct serialization and deserialization scheme
	// are for particular vocabulary values. Therefore, we omit them here
	// and will emit an error.
	//
	// If this error is emitted, it means a code change to the tool is
	// required. A new ontology implementation for this vocabulary needs to
	// be added, and a hardcoded implementation of the value's serialization
	// and deserialization functions must be created. This will then let the
	// rest of the generated code properly serialize and deserialize these
	// values.
	if len(r.v.Values) > 0 {
		return nil, fmt.Errorf("known limitation: value type definitions in a new vocabulary must be embedded in the go-fed tool to ensure that the value is properly serialized and deserialized. This tool is not intelligent enough to automatically somehow deduce what encoding is necessary for new values.")
	}
	return nodes, nil
}

// LoadSpecificAsAlias loads a specific RDFNode with the given alias.
//
// Values cannot be loaded because their serialization and deserialization types
// are not known at runtime if not embedded in the go-fed tool. If the error is
// generated when running the tool, then file a bug so that the tool can
// properly "know" about this particular value and how to serialize and
// deserialize it properly.
func (r *ReferenceOntology) LoadSpecificAsAlias(alias, name string) ([]RDFNode, error) {
	if t, ok := r.v.Types[name]; ok {
		return []RDFNode{
			&AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &typeReference{t: t, vocabName: r.SpecURI()},
			},
		}, nil
	}
	if p, ok := r.v.Properties[name]; ok {
		return []RDFNode{
			&AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &propertyReference{p: p, vocabName: r.SpecURI()},
			},
		}, nil
	}
	if _, ok := r.v.Values[name]; ok {
		// Note: Values cannot be added this way as there's no way to detect
		// at runtime what the correct serialization and deserialization scheme
		// are for particular vocabulary values. Therefore, we omit them here
		// and will emit an error.
		//
		// If this error is emitted, it means a code change to the tool is
		// required. A new ontology implementation for this vocabulary needs to
		// be added, and a hardcoded implementation of the value's serialization
		// and deserialization functions must be created. This will then let the
		// rest of the generated code properly serialize and deserialize these
		// values.
		return nil, fmt.Errorf("known limitation: value type definitions in a new vocabulary must be embedded in the go-fed tool to ensure that the value is properly serialized and deserialized. This tool is not intelligent enough to automatically somehow deduce what encoding is necessary for new values.")
	}
	return nil, fmt.Errorf("ontology (%s) cannot find %q to make alias %q", r.SpecURI(), name, alias)
}

// LoadElement does nothing.
func (r *ReferenceOntology) LoadElement(name string, payload map[string]interface{}) ([]RDFNode, error) {
	return nil, nil
}

// GetByName returns a raw, unguarded node by name.
//
// Values cannot be loaded because their serialization and deserialization types
// are not known at runtime if not embedded in the go-fed tool. If the error is
// generated when running the tool, then file a bug so that the tool can
// properly "know" about this particular value and how to serialize and
// deserialize it properly.
func (r *ReferenceOntology) GetByName(name string) (RDFNode, error) {
	if t, ok := r.v.Types[name]; ok {
		return &typeReference{t: t, vocabName: r.SpecURI()}, nil
	}
	if p, ok := r.v.Properties[name]; ok {
		return &propertyReference{p: p, vocabName: r.SpecURI()}, nil
	}
	if _, ok := r.v.Values[name]; ok {
		// Note: Values cannot be added this way as there's no way to detect
		// at runtime what the correct serialization and deserialization scheme
		// are for particular vocabulary values. Therefore, we omit them here
		// and will emit an error.
		//
		// If this error is emitted, it means a code change to the tool is
		// required. A new ontology implementation for this vocabulary needs to
		// be added, and a hardcoded implementation of the value's serialization
		// and deserialization functions must be created. This will then let the
		// rest of the generated code properly serialize and deserialize these
		// values.
		return nil, fmt.Errorf("known limitation: value type definitions in a new vocabulary must be embedded in the go-fed tool to ensure that the value is properly serialized and deserialized. This tool is not intelligent enough to automatically somehow deduce what encoding is necessary for new values.")
	}
	return nil, fmt.Errorf("ontology (%s) cannot find node for name %s", r.SpecURI(), name)
}

var _ RDFNode = &typeReference{}

// typeReference adds a VocabularyReference for a VocabularyType in another
// vocabulary.
type typeReference struct {
	t         VocabularyType
	vocabName string
}

// Enter returns an error.
func (*typeReference) Enter(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("typeReference cannot be entered")
}

// Exit returns an error.
func (*typeReference) Exit(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("typeReference cannot be exited")
}

// Apply sets a reference in the context.
func (t *typeReference) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	ref, ok := ctx.Current.(*VocabularyReference)
	if !ok {
		// May be during resolve reference phase -- nothing to do.
		return true, nil
	}
	ref.Name = t.t.GetName()
	ref.URI = t.t.URI
	ref.Vocab = t.vocabName
	return true, nil
}

var _ RDFNode = &propertyReference{}

// typeReference adds a VocabularyReference for a VocabularyProperty in another
// vocabulary.
type propertyReference struct {
	p         VocabularyProperty
	vocabName string
}

// Enter returns an error.
func (*propertyReference) Enter(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("propertyReference cannot be entered")
}

// Exit returns an error.
func (*propertyReference) Exit(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("propertyReference cannot be exited")
}

// Apply sets a reference in the context.
func (p *propertyReference) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	ref, ok := ctx.Current.(*VocabularyReference)
	if !ok {
		// May be during resolve reference phase -- nothing to do.
		return true, nil
	}
	ref.Name = p.p.GetName()
	ref.URI = p.p.URI
	ref.Vocab = p.vocabName
	return true, nil
}
