package rdf

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
)

// ParsedVocabulary is the internal data structure produced after parsing the
// definition of an ActivityStream vocabulary. It is the intermediate
// understanding of the specification in the context of certain ontologies.
//
// At the end of parsing, the ParsedVocabulary is not guaranteed to be
// semantically valid, just that the parser resolved all important ontological
// details.
//
// Note that the Order field contains the order in which parsed specs were
// understood and resolved. Kinds added as references (such as XML, Schema.org,
// or rdfs types) are not included in Order. It is expected that the last
// element of Order must be the vocabulary in Vocab.
type ParsedVocabulary struct {
	Vocab      Vocabulary
	References map[string]*Vocabulary
	Order      []string
}

// Size returns the number of types, properties, and values in the parsed
// vocabulary.
func (p ParsedVocabulary) Size() int {
	s := p.Vocab.Size()
	for _, v := range p.References {
		s += v.Size()
	}
	return s
}

// Clone creates a copy of this ParsedVocabulary. Note that the cloned
// vocabulary does not copy References, so the original and clone both have
// pointers to the same referenced vocabularies.
func (p ParsedVocabulary) Clone() *ParsedVocabulary {
	clone := &ParsedVocabulary{
		Vocab:      p.Vocab,
		References: make(map[string]*Vocabulary, len(p.References)),
		Order:      make([]string, len(p.Order)),
	}
	for k, v := range p.References {
		clone.References[k] = v
	}
	copy(clone.Order, p.Order)
	return clone
}

// GetReference looks up a reference based on its URI.
func (p *ParsedVocabulary) GetReference(uri string) (*Vocabulary, error) {
	httpSpec, httpsSpec, err := ToHttpAndHttps(uri)
	if err != nil {
		return nil, err
	}
	if p.References == nil {
		p.References = make(map[string]*Vocabulary)
	}
	if v, ok := p.References[httpSpec]; ok {
		return v, nil
	} else if v, ok := p.References[httpsSpec]; ok {
		return v, nil
	} else {
		p.References[uri] = &Vocabulary{}
	}
	return p.References[uri], nil
}

// String returns a printable version of this ParsedVocabulary for debugging.
func (p ParsedVocabulary) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("Vocab:\n%s", p.Vocab))
	for k, v := range p.References {
		b.WriteString(fmt.Sprintf("Reference %s:\n\t%s\n", k, v))
	}
	return b.String()
}

// Vocabulary contains the type, property, and value definitions for a single
// ActivityStreams or extension vocabulary.
type Vocabulary struct {
	Name           string
	WellKnownAlias string // Hack.
	URI            *url.URL
	Types          map[string]VocabularyType
	Properties     map[string]VocabularyProperty
	Values         map[string]VocabularyValue
	Registry       *RDFRegistry
}

// Size returns the number of types, properties, and values in this vocabulary.
func (v Vocabulary) Size() int {
	return len(v.Types) + len(v.Properties) + len(v.Values)
}

// GetName returns the vocabulary's name.
func (v Vocabulary) GetName() string {
	return v.Name
}

// GetWellKnownAlias returns the vocabulary's name.
func (v Vocabulary) GetWellKnownAlias() string {
	return v.WellKnownAlias
}

// SetName sets the vocabulary's name.
func (v *Vocabulary) SetName(s string) {
	v.Name = s
}

// SetURI sets the value's URI.
func (v *Vocabulary) SetURI(s string) error {
	var e error
	v.URI, e = url.Parse(s)
	return e
}

// String returns a printable version of this Vocabulary for debugging.
func (v Vocabulary) String() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("Vocabulary %q\n", v.Name))
	for k, v := range v.Types {
		b.WriteString(fmt.Sprintf("Type %s:\n\t%s\n", k, v))
	}
	for k, v := range v.Properties {
		b.WriteString(fmt.Sprintf("Property %s:\n\t%s\n", k, v))
	}
	for k, v := range v.Values {
		b.WriteString(fmt.Sprintf("Value %s:\n\t%s\n", k, v))
	}
	return b.String()
}

// SetType sets a type keyed by its name. Returns an error if a type is already
// set for that name.
func (v *Vocabulary) SetType(name string, a *VocabularyType) error {
	if v.Types == nil {
		v.Types = make(map[string]VocabularyType, 1)
	}
	if _, has := v.Types[name]; has {
		return fmt.Errorf("name %q already exists for vocabulary Types", name)
	}
	v.Types[name] = *a
	return nil
}

// SetProperty sets a property keyed by its name. Returns an error if a property
// is already set for that name.
func (v *Vocabulary) SetProperty(name string, a *VocabularyProperty) error {
	if v.Properties == nil {
		v.Properties = make(map[string]VocabularyProperty, 1)
	}
	if _, has := v.Properties[name]; has {
		return fmt.Errorf("name already exists for vocabulary Properties")
	}
	v.Properties[name] = *a
	return nil
}

// SetValue sets a value keyed by its name. Returns an error if the value is
// already set for that name.
func (v *Vocabulary) SetValue(name string, a *VocabularyValue) error {
	if v.Values == nil {
		v.Values = make(map[string]VocabularyValue, 1)
	}
	if _, has := v.Values[name]; has {
		return fmt.Errorf("name already exists for vocabulary Values")
	}
	v.Values[name] = *a
	return nil
}

var (
	_ NameSetter = &Vocabulary{}
	_ NameGetter = &Vocabulary{}
	_ URISetter  = &Vocabulary{}
)

// VocabularyValue represents a value type that properties can take on.
type VocabularyValue struct {
	Name           string
	URI            *url.URL
	DefinitionType *jen.Statement
	Zero           string
	IsNilable      bool
	IsURI          bool
	SerializeFn    *codegen.Function
	DeserializeFn  *codegen.Function
	LessFn         *codegen.Function
}

// String returns a printable version of this value for debugging.
func (v VocabularyValue) String() string {
	return fmt.Sprintf("Value=%s,%s,%s,%s", v.Name, v.URI, v.DefinitionType, v.Zero)
}

// SetName sets the value's name.
func (v *VocabularyValue) SetName(s string) {
	v.Name = s
}

// GetName returns the value's name.
func (v VocabularyValue) GetName() string {
	return v.Name
}

// SetURI sets the value's URI.
func (v *VocabularyValue) SetURI(s string) error {
	var e error
	v.URI, e = url.Parse(s)
	return e
}

var (
	_ NameSetter = &VocabularyValue{}
	_ NameGetter = &VocabularyValue{}
	_ URISetter  = &VocabularyValue{}
)

// VocabularyType represents a single ActivityStream type in a vocabulary.
type VocabularyType struct {
	Name              string
	Typeless          bool // Hack
	URI               *url.URL
	Notes             string
	DisjointWith      []VocabularyReference
	Extends           []VocabularyReference
	Examples          []VocabularyExample
	Properties        []VocabularyReference
	WithoutProperties []VocabularyReference
}

// String returns a printable version of this type, for debugging.
func (v VocabularyType) String() string {
	return fmt.Sprintf("Type=%s,%s,%s\n\tDJW=%s\n\tExt=%s\n\tEx=%s", v.Name, v.URI, v.Notes, v.DisjointWith, v.Extends, v.Examples)
}

// SetName sets the name of this type.
func (v *VocabularyType) SetName(s string) {
	v.Name = s
}

// SetName returns the name of this type.
func (v VocabularyType) GetName() string {
	return v.Name
}

// TypeName returns the name of this type.
//
// Used to satisfy an interface.
func (v VocabularyType) TypeName() string {
	return v.Name
}

// SetURI sets the URI of this type, returning an error if it cannot parse the
// URI.
func (v *VocabularyType) SetURI(s string) error {
	var e error
	v.URI, e = url.Parse(s)
	return e
}

// SetNotes sets the notes on this type.
func (v *VocabularyType) SetNotes(s string) {
	v.Notes = s
}

// AddExample adds an example on this type.
func (v *VocabularyType) AddExample(e *VocabularyExample) {
	v.Examples = append(v.Examples, *e)
}

// IsTypeless determines if this type is, in fact, typeless
func (v *VocabularyType) IsTypeless() bool {
	return v.Typeless
}

var (
	_ NameSetter   = &VocabularyType{}
	_ NameGetter   = &VocabularyType{}
	_ URISetter    = &VocabularyType{}
	_ NotesSetter  = &VocabularyType{}
	_ ExampleAdder = &VocabularyType{}
)

// VocabularyProperty represents a single ActivityStream property type in a
// vocabulary.
type VocabularyProperty struct {
	Name           string
	URI            *url.URL
	Notes          string
	Domain         []VocabularyReference
	Range          []VocabularyReference
	DoesNotApplyTo []VocabularyReference
	Examples       []VocabularyExample
	// SubpropertyOf is ignorable as long as data is set up correctly
	SubpropertyOf      VocabularyReference // Must be a VocabularyProperty
	Functional         bool
	NaturalLanguageMap bool
}

// String returns a printable version of this property for debugging.
func (v VocabularyProperty) String() string {
	return fmt.Sprintf("Property=%s,%s,%s\n\tD=%s\n\tR=%s\n\tEx=%s\n\tSub=%s\n\tDNApply=%s\n\tfunc=%t,natLangMap=%t", v.Name, v.URI, v.Notes, v.Domain, v.Range, v.Examples, v.SubpropertyOf, v.DoesNotApplyTo, v.Functional, v.NaturalLanguageMap)
}

// SetName sets the name on this property.
func (v *VocabularyProperty) SetName(s string) {
	v.Name = s
}

// GetName returns the name of this property.
func (v VocabularyProperty) GetName() string {
	return v.Name
}

// PropertyName returns the name of this property.
//
// Used to satisfy an interface.
func (v VocabularyProperty) PropertyName() string {
	return v.Name
}

// SetURI sets the URI for this property, returning an error if it cannot be
// parsed.
func (v *VocabularyProperty) SetURI(s string) error {
	var e error
	v.URI, e = url.Parse(s)
	return e
}

// SetNotes sets notes on this Property.
func (v *VocabularyProperty) SetNotes(s string) {
	v.Notes = s
}

// AddExample adds an example for this property.
func (v *VocabularyProperty) AddExample(e *VocabularyExample) {
	v.Examples = append(v.Examples, *e)
}

var (
	_ NameSetter   = &VocabularyProperty{}
	_ NameGetter   = &VocabularyProperty{}
	_ URISetter    = &VocabularyProperty{}
	_ NotesSetter  = &VocabularyProperty{}
	_ ExampleAdder = &VocabularyProperty{}
)

// VocabularyExample documents an Example for an ActivityStream type or property
// in the vocabulary.
type VocabularyExample struct {
	Name    string
	URI     *url.URL
	Example interface{}
}

// String returns a printable string used for debugging.
func (v VocabularyExample) String() string {
	return fmt.Sprintf("VocabularyExample: %s,%s,%s", v.Name, v.URI, v.Example)
}

// SetName sets the name on this example.
func (v *VocabularyExample) SetName(s string) {
	v.Name = s
}

// GetName returns the name of this example.
func (v VocabularyExample) GetName() string {
	return v.Name
}

// SetURI sets the URI of this example, returning an error if it cannot be
// parsed.
func (v *VocabularyExample) SetURI(s string) error {
	var e error
	v.URI, e = url.Parse(s)
	return e
}

var (
	_ NameSetter = &VocabularyExample{}
	_ NameGetter = &VocabularyExample{}
	_ URISetter  = &VocabularyExample{}
)

// VocabularyReference refers to another Vocabulary reference, either a
// VocabularyType, VocabularyValue, or a VocabularyProperty. It may refer to
// another Vocabulary's type or property entirely.
type VocabularyReference struct {
	Name  string
	URI   *url.URL
	Vocab string // If present, must match key in ParsedVocabulary.References
}

// String returns a printable string for this reference, used for debugging.
func (v VocabularyReference) String() string {
	return fmt.Sprintf("VocabularyReference: %s,%s,%s", v.Name, v.URI, v.Vocab)
}

// SetName sets the name of this reference.
func (v *VocabularyReference) SetName(s string) {
	v.Name = s
}

// GetName returns the name of this reference.
func (v VocabularyReference) GetName() string {
	return v.Name
}

// SetURI sets the URI for this reference. Returns an error if the URI cannot
// be parsed.
func (v *VocabularyReference) SetURI(s string) error {
	var e error
	v.URI, e = url.Parse(s)
	return e
}

var (
	_ NameSetter = &VocabularyReference{}
	_ NameGetter = &VocabularyReference{}
	_ URISetter  = &VocabularyReference{}
)
