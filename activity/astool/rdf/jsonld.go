package rdf

import (
	"fmt"
)

const (
	typeSpec                = "@type"
	typeActivityStreamsSpec = "type"
	IdSpec                  = "@id"
	IdActivityStreamsSpec   = "id"
	ContainerSpec           = "@container"
	IndexSpec               = "@index"
	// ActivityStreams specifically disallows the 'object' property on
	// certain IntransitiveActivity and subtypes. There is no RDF mechanism
	// to describe this. So this is a stupid hack, based on the assumption
	// that no one -- W3C or otherwise -- will name a reserved word with a
	// "@wtf_" prefix due to the reserved '@', the use of the unprofessional
	// 'wtf', and a style-breaking underscore.
	withoutPropertySpec = "@wtf_without_property"
	typelessSpec        = "@wtf_typeless"
	// TODO: Support WellKnownAlias
)

// jsonLDNodes contains the well-known set of nodes as defined by the JSON-LD
// specification.
func jsonLDNodes(r *RDFRegistry) []RDFNode {
	// Order matters -- we want to be able to distinguish the types of
	// things without other nodes hijacking the flow.
	return []RDFNode{
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     typeSpec,
			Delegate: &typeLD{r: r},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     typeActivityStreamsSpec,
			Delegate: &typeLD{r: r},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     IdSpec,
			Delegate: &idLD{},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     IdActivityStreamsSpec,
			Delegate: &idLD{},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     ContainerSpec,
			Delegate: &ContainerLD{},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     IndexSpec,
			Delegate: &IndexLD{},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     withoutPropertySpec,
			Delegate: &withoutProperty{},
		},
		&AliasedDelegate{
			Spec:     "",
			Alias:    "",
			Name:     typelessSpec,
			Delegate: &typeless{},
		},
	}
}

var _ RDFNode = &idLD{}

// idLD is an RDFNode for the 'id' property.
type idLD struct{}

// Enter returns an error.
func (i *idLD) Enter(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("id cannot be entered")
}

// Exit returns an error.
func (i *idLD) Exit(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("id cannot be exited")
}

// Apply sets the URI for the context's Current item.
func (i *idLD) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	if ctx.Current == nil {
		return true, nil
	} else if ider, ok := ctx.Current.(URISetter); !ok {
		return true, fmt.Errorf("id apply called with non-URISetter")
	} else if str, ok := value.(string); !ok {
		return true, fmt.Errorf("id apply called with non-string value")
	} else {
		return true, ider.SetURI(str)
	}
}

var _ RDFNode = &typeLD{}

// typeLD is an RDFNode for the 'type' property.
type typeLD struct {
	r *RDFRegistry
}

// Enter does nothing.
func (t *typeLD) Enter(key string, ctx *ParsingContext) (bool, error) {
	return true, nil
}

// Exit does nothing.
func (t *typeLD) Exit(key string, ctx *ParsingContext) (bool, error) {
	return true, nil
}

// Apply attempts to get the RDFNode for the type and apply it.
func (t *typeLD) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	vs, ok := value.(string)
	if !ok {
		return true, fmt.Errorf("@type is not string")
	}
	n, e := t.r.getNode(vs)
	if e != nil {
		return true, e
	}
	return n.Apply(vs, nil, ctx)
}

var _ RDFNode = &ContainerLD{}

// ContainerLD is an RDFNode that delegates to an RDFNode but only at this
// next level.
type ContainerLD struct {
	ContainsNode RDFNode
}

// Enter sets OnlyApplyThisNodeNextLevel on the ParsingContext.
//
// Returns an error if this is the second time Enter is called in a row.
func (c *ContainerLD) Enter(key string, ctx *ParsingContext) (bool, error) {
	if ctx.OnlyApplyThisNodeNextLevel != nil {
		return true, fmt.Errorf("@container parsing context exit already has non-nil node")
	}
	ctx.SetOnlyApplyThisNodeNextLevel(c.ContainsNode)
	return true, nil
}

// Exit clears OnlyApplyThisNodeNextLevel on the ParsingContext.
//
// Returns an error if this is the second time Exit is called in a row.
func (c *ContainerLD) Exit(key string, ctx *ParsingContext) (bool, error) {
	if ctx.OnlyApplyThisNodeNextLevel == nil {
		return true, fmt.Errorf("@container parsing context exit already has nil node")
	}
	ctx.ResetOnlyAppliedThisNodeNextLevel()
	return true, nil
}

// Apply does nothing.
func (c *ContainerLD) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	return true, nil
}

var _ RDFNode = &IndexLD{}

// IndexLD does nothing.
//
// It could try to manage human-defined indices, but the machine doesn't care.
type IndexLD struct{}

// Enter does nothing.
func (i *IndexLD) Enter(key string, ctx *ParsingContext) (bool, error) {
	return true, nil
}

// Exit does nothing.
func (i *IndexLD) Exit(key string, ctx *ParsingContext) (bool, error) {
	return true, nil
}

// Apply does nothing.
func (i *IndexLD) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	return true, nil
}

var _ RDFNode = &withoutProperty{}

// withoutProperty is a hacky-as-hell way to manage ActivityStream's concept of
// "WithoutProperty". It isn't a defined RDF relationship, so this is
// non-standard but required of the ActivityStreams Core or Extended Types spec.
type withoutProperty struct{}

// Enter pushes a VocabularyReference. It is expected further nodes will
// populate it with information before dxiting this node.
func (w *withoutProperty) Enter(key string, ctx *ParsingContext) (bool, error) {
	ctx.Push()
	ctx.Current = &VocabularyReference{}
	return true, nil
}

// Exit pops a VocabularyReferences and sets DoesNotApplyTo on the parent
// VocabularyProperty on the stack.
func (w *withoutProperty) Exit(key string, ctx *ParsingContext) (bool, error) {
	i := ctx.Current
	ctx.Pop()
	vr, ok := i.(*VocabularyReference)
	if !ok {
		return true, fmt.Errorf("hacky withoutProperty exit did not get *rdf.VocabularyReference")
	}
	vp, ok := ctx.Current.(*VocabularyProperty)
	if !ok {
		return true, fmt.Errorf("hacky withoutProperty exit Current is not *rdf.VocabularyProperty")
	}
	vp.DoesNotApplyTo = append(vp.DoesNotApplyTo, *vr)
	return true, nil
}

// Apply returns an error.
func (w *withoutProperty) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("hacky withoutProperty cannot be applied")
}

var _ RDFNode = &typeless{}

// typeless is a hacky-as-hell way to rectify the fact that certain ontologies
// have classes that do not correspond to the JSON-LD idea of an @type.
// I didn't even bother looking for an existing RDF concept and instead would
// rather force myself to suffer in order to prove how awful this is. Waah.
type typeless struct{}

// Enter returns an error.
func (t *typeless) Enter(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("hacky typeless cannot be entered")
}

// Exit returns an error.
func (t *typeless) Exit(key string, ctx *ParsingContext) (bool, error) {
	return true, fmt.Errorf("hacky typeless cannot be exited")
}

// Apply sets whether this type is actually typeless.
func (t *typeless) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	val, ok := value.(bool)
	if !ok {
		return true, fmt.Errorf("hacky typeless value is not a bool")
	}
	vt, ok := ctx.Current.(*VocabularyType)
	if !ok {
		return true, fmt.Errorf("hacky typeless Current is not *rdf.VocabularyType")
	}
	vt.Typeless = val
	return true, nil
}
