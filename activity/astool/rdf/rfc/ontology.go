// Package RFC contains ontology values that are defined in RFCs, BCPs, and
// other miscellaneous standards.
package rfc

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/superseriousbusiness/activity/astool/codegen"
	"github.com/superseriousbusiness/activity/astool/rdf"
)

const (
	rfcName   = "RFC"
	rfcSpec   = "https://tools.ietf.org/html/"
	bcp47Spec = "bcp47"
	mimeSpec  = "rfc2045" // See also: rfc2046 and rfc6838
	relSpec   = "rfc5988"
)

// RFCOntology represents standards and values that originate from RFC
// specifications.
type RFCOntology struct {
	Package string
}

// SpecURI returns the RFC specifications URI.
func (o *RFCOntology) SpecURI() string {
	return rfcSpec
}

// Load without an alias.
func (o *RFCOntology) Load() ([]rdf.RDFNode, error) {
	return o.LoadAsAlias("")
}

// LoadAsAlias loads with the given alias.
func (o *RFCOntology) LoadAsAlias(s string) ([]rdf.RDFNode, error) {
	return []rdf.RDFNode{
		&rdf.AliasedDelegate{
			Spec:     rfcSpec,
			Alias:    s,
			Name:     bcp47Spec,
			Delegate: &bcp47{pkg: o.Package},
		},
		&rdf.AliasedDelegate{
			Spec:     rfcSpec,
			Alias:    s,
			Name:     mimeSpec,
			Delegate: &mime{pkg: o.Package},
		},
		&rdf.AliasedDelegate{
			Spec:     rfcSpec,
			Alias:    s,
			Name:     relSpec,
			Delegate: &rel{pkg: o.Package},
		},
	}, nil
}

// LoadSpecificAsAlias loads a specific item with a given alias.
func (o *RFCOntology) LoadSpecificAsAlias(alias, name string) ([]rdf.RDFNode, error) {
	switch name {
	case bcp47Spec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &bcp47{pkg: o.Package},
			},
		}, nil
	case mimeSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &mime{pkg: o.Package},
			},
		}, nil
	case relSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &rel{pkg: o.Package},
			},
		}, nil
	}
	return nil, fmt.Errorf("rfc ontology cannot find %q to alias to %q", name, alias)
}

// LoadElement does nothing.
func (o *RFCOntology) LoadElement(name string, payload map[string]interface{}) ([]rdf.RDFNode, error) {
	return nil, nil
}

// GetByName obtains a bare node by name.
func (o *RFCOntology) GetByName(name string) (rdf.RDFNode, error) {
	name = strings.TrimPrefix(name, o.SpecURI())
	switch name {
	case bcp47Spec:
		return &bcp47{pkg: o.Package}, nil
	case mimeSpec:
		return &mime{pkg: o.Package}, nil
	case relSpec:
		return &rel{pkg: o.Package}, nil
	}
	return nil, fmt.Errorf("rfc ontology could not find node for name %s", name)
}

var _ rdf.RDFNode = &bcp47{}

// BCP47 represents a BCP47 value.
//
// No validation is done on deserialized values.
type bcp47 struct {
	pkg string
}

// Enter does nothing.
func (b *bcp47) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("bcp47 langaugetag cannot be entered")
}

// Exit does nothing.
func (b *bcp47) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("bcp47 languagetag cannot be exited")
}

// Apply adds BCP47 as a value Kind.
func (b *bcp47) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	v, err := ctx.GetResultReferenceWithDefaults(rfcSpec, rfcName)
	if err != nil {
		return true, err
	}
	if len(v.Values[bcp47Spec].Name) == 0 {
		u, err := url.Parse(rfcSpec + bcp47Spec)
		if err != nil {
			return true, err
		}
		val := &rdf.VocabularyValue{
			Name:           bcp47Spec,
			URI:            u,
			DefinitionType: jen.String(),
			Zero:           "\"\"",
			IsNilable:      false,
			SerializeFn: rdf.SerializeValueFunction(
				b.pkg,
				bcp47Spec,
				jen.String(),
				[]jen.Code{
					jen.Return(
						jen.Id(codegen.This()),
						jen.Nil(),
					),
				}),
			DeserializeFn: rdf.DeserializeValueFunction(
				b.pkg,
				bcp47Spec,
				jen.String(),
				[]jen.Code{
					jen.If(
						jen.List(
							jen.Id("s"),
							jen.Id("ok"),
						).Op(":=").Id(codegen.This()).Assert(jen.String()),
						jen.Id("ok"),
					).Block(
						jen.Return(
							jen.Id("s"),
							jen.Nil(),
						),
					).Else().Block(
						jen.Return(
							jen.Lit(""),
							jen.Qual("fmt", "Errorf").Call(
								jen.Lit("%v cannot be interpreted as a string for bcp47 languagetag"),
								jen.Id(codegen.This()),
							),
						),
					),
				}),
			LessFn: rdf.LessFunction(
				b.pkg,
				bcp47Spec,
				jen.String(),
				[]jen.Code{
					jen.Return(
						jen.Id("lhs").Op("<").Id("rhs"),
					),
				}),
		}
		if err = v.SetValue(bcp47Spec, val); err != nil {
			return true, err
		}
	}
	return true, nil
}

var _ rdf.RDFNode = &mime{}

// mime represents MIME values.
type mime struct {
	pkg string
}

// Enter does nothing.
func (*mime) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("MIME media type cannot be entered")
}

// Exit does nothing.
func (*mime) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("MIME media type cannot be exited")
}

// Apply adds MIME as a value Kind.
func (m *mime) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	v, err := ctx.GetResultReferenceWithDefaults(rfcSpec, rfcName)
	if err != nil {
		return true, err
	}
	if len(v.Values[mimeSpec].Name) == 0 {
		u, err := url.Parse(rfcSpec + mimeSpec)
		if err != nil {
			return true, err
		}
		val := &rdf.VocabularyValue{
			Name:           mimeSpec,
			URI:            u,
			DefinitionType: jen.String(),
			Zero:           "\"\"",
			IsNilable:      false,
			SerializeFn: rdf.SerializeValueFunction(
				m.pkg,
				mimeSpec,
				jen.String(),
				[]jen.Code{
					jen.Return(
						jen.Id(codegen.This()),
						jen.Nil(),
					),
				}),
			DeserializeFn: rdf.DeserializeValueFunction(
				m.pkg,
				mimeSpec,
				jen.String(),
				[]jen.Code{
					jen.If(
						jen.List(
							jen.Id("s"),
							jen.Id("ok"),
						).Op(":=").Id(codegen.This()).Assert(jen.String()),
						jen.Id("ok"),
					).Block(
						jen.Return(
							jen.Id("s"),
							jen.Nil(),
						),
					).Else().Block(
						jen.Return(
							jen.Lit(""),
							jen.Qual("fmt", "Errorf").Call(
								jen.Lit("%v cannot be interpreted as a string for MIME media type"),
								jen.Id(codegen.This()),
							),
						),
					),
				}),
			LessFn: rdf.LessFunction(
				m.pkg,
				mimeSpec,
				jen.String(),
				[]jen.Code{
					jen.Return(
						jen.Id("lhs").Op("<").Id("rhs"),
					),
				}),
		}
		if err = v.SetValue(mimeSpec, val); err != nil {
			return true, err
		}
	}
	return true, nil
}

var _ rdf.RDFNode = &rel{}

// rel is a Link Relation.
type rel struct {
	pkg string
}

// Enter does nothing.
func (*rel) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rel cannot be entered")
}

// Exit does nothing.
func (*rel) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("rel cannot be exited")
}

// Apply adds rel as a supported value Kind.
func (r *rel) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	v, err := ctx.GetResultReferenceWithDefaults(rfcSpec, rfcName)
	if err != nil {
		return true, err
	}
	if len(v.Values[relSpec].Name) == 0 {
		u, err := url.Parse(rfcSpec + relSpec)
		if err != nil {
			return true, err
		}
		val := &rdf.VocabularyValue{
			Name:           relSpec,
			URI:            u,
			DefinitionType: jen.String(),
			Zero:           "\"\"",
			IsNilable:      false,
			SerializeFn: rdf.SerializeValueFunction(
				r.pkg,
				relSpec,
				jen.String(),
				[]jen.Code{
					jen.Return(
						jen.Id(codegen.This()),
						jen.Nil(),
					),
				}),
			DeserializeFn: rdf.DeserializeValueFunction(
				r.pkg,
				relSpec,
				jen.String(),
				[]jen.Code{
					jen.If(
						jen.List(
							jen.Id("s"),
							jen.Id("ok"),
						).Op(":=").Id(codegen.This()).Assert(jen.String()),
						jen.Id("ok"),
					).Block(
						jen.Return(
							jen.Id("s"),
							jen.Nil(),
						),
					).Else().Block(
						jen.Return(
							jen.Lit(""),
							jen.Qual("fmt", "Errorf").Call(
								jen.Lit("%v cannot be interpreted as a string for rel"),
								jen.Id(codegen.This()),
							),
						),
					),
				}),
			LessFn: rdf.LessFunction(
				r.pkg,
				relSpec,
				jen.String(),
				[]jen.Code{
					jen.Return(
						jen.Id("lhs").Op("<").Id("rhs"),
					),
				}),
		}
		if err = v.SetValue(relSpec, val); err != nil {
			return true, err
		}
	}
	return true, nil
}
