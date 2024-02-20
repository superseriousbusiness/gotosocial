package rdf

import (
	"fmt"
	"net/url"
)

const (
	JSON_LD_CONTEXT = "@context"
	JSON_LD_TYPE    = "@type"
	JSON_LD_TYPE_AS = "type"
)

// JSONLD is an alias for the generic map of keys to interfaces, presumably
// parsed from a JSON-encoded context definition file.
type JSONLD map[string]interface{}

// ParsingContext contains the results of the parsing as well as scratch space
// required for RDFNodes to be able to statefully apply changes.
type ParsingContext struct {
	// Result contains the final ParsedVocabulary from a file.
	Result *ParsedVocabulary
	// Current item to operate upon. A call to Push or Pop will overwrite
	// this field.
	Current interface{}
	// Name of the Current item. A call to Push or Pop will modify this
	// field.
	Name string
	// The Stack of Types, Properties, References, Examples, and other
	// items being analyzed. A call to Push or Pop will modify this field.
	//
	// Do not use directly, instead use Push and Pop.
	Stack []interface{}
	// Applies the node only for the next level of processing.
	//
	// Do not touch, instead use the accessor methods.
	OnlyApplyThisNodeNextLevel RDFNode
	// OnlyApplied keeps track if OnlyApplyThisNodeNextLevel has applied
	// once.
	OnlyApplied bool
	// Applies the node once, for the rest of the data. This skips the
	// recursive parsing, and the node's Apply is given an empty string
	// for a key.
	//
	// Do not touch, instead use the accessor methods.
	OnlyApplyThisNode RDFNode
}

// GetResultReferenceWithDefaults will fetch the spec and set the Vocabulary
// Name and URI values as well. Helper function when getting a reference in
// order to populate known value types.
func (p *ParsingContext) GetResultReferenceWithDefaults(spec, name string) (*Vocabulary, error) {
	v, err := p.Result.GetReference(spec)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(spec)
	if err != nil {
		return nil, err
	}
	v.Name = name
	v.URI = u
	return v, nil
}

// SetOnlyApplyThisNode sets the provided node to be the only one applied until
// ResetOnlyApplyThisNode is called.
func (p *ParsingContext) SetOnlyApplyThisNode(n RDFNode) {
	p.OnlyApplyThisNode = n
}

// ResetOnlyApplyThisNode clears the only node to apply, if set.
func (p *ParsingContext) ResetOnlyApplyThisNode() {
	p.OnlyApplyThisNode = nil
}

// SetOnlyApplyThisNodeNExtLevel will apply the next node only for the next
// level.
func (p *ParsingContext) SetOnlyApplyThisNodeNextLevel(n RDFNode) {
	p.OnlyApplyThisNodeNextLevel = n
	p.OnlyApplied = false
}

// GetNextNodes is given the list of nodes a parent process believes should be
// applied, and returns the list of nodes that actually should be used.
//
// If there is node that should only apply or should only apply at the next
// level (and hasn't yet), then the passed in list will not match the resulting
// list.
func (p *ParsingContext) GetNextNodes(n []RDFNode) (r []RDFNode, clearFn func()) {
	if p.OnlyApplyThisNodeNextLevel == nil {
		return n, func() {}
	} else if p.OnlyApplied {
		return n, func() {}
	} else {
		p.OnlyApplied = true
		return []RDFNode{p.OnlyApplyThisNodeNextLevel}, func() {
			p.OnlyApplied = false
		}
	}
}

// ResetOnlyAppliedThisNodeNextLevel clears the node that should have been
// applied for the next level of depth only.
func (p *ParsingContext) ResetOnlyAppliedThisNodeNextLevel() {
	p.OnlyApplyThisNodeNextLevel = nil
	p.OnlyApplied = false
}

// Push puts the Current onto the Stack.
func (p *ParsingContext) Push() {
	p.Stack = append([]interface{}{p.Current}, p.Stack...)
	p.Reset()
}

// Pop puts the top item on the Stack into Current, and sets Name as
// appropriate.
func (p *ParsingContext) Pop() {
	p.Current = p.Stack[0]
	p.Stack = p.Stack[1:]
	if ng, ok := p.Current.(NameGetter); ok {
		p.Name = ng.GetName()
	}
}

// IsReset determines if the Context's Current is nil and Name is empty. Note
func (p *ParsingContext) IsReset() bool {
	return p.Current == nil &&
		p.Name == ""
}

// Reset sets Current to nil and Name to empty string.
func (p *ParsingContext) Reset() {
	p.Current = nil
	p.Name = ""
}

// NameSetter is a utility interface for the rdf Vocabulary types.
type NameSetter interface {
	SetName(string)
}

// NameGetter is a utility interface for the rdf Vocabulary types.
type NameGetter interface {
	GetName() string
}

// URISetter is a utility interface for the rdf Vocabulary types.
type URISetter interface {
	SetURI(string) error
}

// NotesSetter is a utility interface for the rdf Vocabulary types.
type NotesSetter interface {
	SetNotes(string)
}

// ExampleAdder is a utility interface for the rdf Vocabulary types.
type ExampleAdder interface {
	AddExample(*VocabularyExample)
}

// RDFNode is able to operate on a specific key if it applies towards its
// ontology (determined at creation time). It applies the value in its own
// specific implementation on the context.
type RDFNode interface {
	// Enter is called when the RDFNode is a label for an array of values or
	// a key within a JSON object, and the parser is about to examine its
	// value(s). Exit is guaranteed to be called afterwards.
	Enter(key string, ctx *ParsingContext) (bool, error)
	// Exit is called after the parser examines the node's value(s).
	Exit(key string, ctx *ParsingContext) (bool, error)
	// Apply is called by the parser on nodes when they appear as values.
	Apply(key string, value interface{}, ctx *ParsingContext) (bool, error)
}

// ParseVocabularies parses the provided inputs in order as an ActivityStreams
// context that specifies one or more extension vocabularies.
func ParseVocabularies(registry *RDFRegistry, inputs []JSONLD) (vocabulary *ParsedVocabulary, err error) {
	vocabulary = &ParsedVocabulary{
		References: make(map[string]*Vocabulary, len(inputs)-1),
	}
	currentRegistry := registry.clone()
	for i, input := range inputs {
		var v *ParsedVocabulary
		v, err = parseVocabulary(currentRegistry, input, vocabulary.References)
		if err != nil {
			return
		}
		for k, ref := range v.References {
			if ref.Registry != nil {
				err = ref.Registry.AddOntology(&ReferenceOntology{v.Vocab})
				if err != nil {
					return
				}
			}
			vocabulary.References[k] = ref
		}
		if i < len(inputs)-1 {
			currentRegistry = v.Vocab.Registry.clone()
			err = currentRegistry.AddOntology(&ReferenceOntology{v.Vocab})
			if err != nil {
				return
			}
			vocabulary.References[v.Vocab.URI.String()] = &v.Vocab
		} else {
			vocabulary.Vocab = v.Vocab
		}
		vocabulary.Order = append(vocabulary.Order, v.Vocab.URI.String())
	}
	return
}

// parseVocabulary parses the specified input as an ActivityStreams context that
// specifies a Core, Extended, or Extension vocabulary.
func parseVocabulary(registry *RDFRegistry, input JSONLD, references map[string]*Vocabulary) (vocabulary *ParsedVocabulary, err error) {
	var nodes []RDFNode
	nodes, err = parseJSONLDContext(registry, input)
	if err != nil {
		return
	}
	vocabulary = &ParsedVocabulary{References: make(map[string]*Vocabulary, len(references))}
	for k, v := range references {
		vocabulary.References[k] = v
	}
	ctx := &ParsingContext{
		Result: vocabulary,
	}
	// Prepend well-known JSON LD parsing nodes. Order matters, so that the
	// parser can understand things like types so that other nodes do not
	// hijack processing.
	nodes = append(jsonLDNodes(registry), nodes...)
	// Step 1: Parse all core data, excluding:
	//   - Value types
	//   - Referenced types
	//   - VocabularyType's 'Properties' and 'WithoutProperties' fields
	//
	// This is all horrible code but it works, so....
	err = apply(nodes, input, ctx)
	if err != nil {
		return
	}
	ctx.Reset()
	// Step 2: Populate value and referenced types.
	err = resolveReferences(registry, ctx)
	if err != nil {
		return
	}
	// Step 3: Populate VocabularyType's 'Properties' and
	// 'WithoutProperties' fields
	err = populatePropertiesOnTypes(registry, ctx)
	vocabulary.Vocab.Registry = registry
	return
}

// populatePropertiesOnTypes populates the 'Properties' and 'WithoutProperties'
// entries on a VocabularyType.
func populatePropertiesOnTypes(registry *RDFRegistry, ctx *ParsingContext) error {
	for _, p := range ctx.Result.Vocab.Properties {
		if err := populatePropertyOnTypes(registry, p, ctx.Result.Vocab.URI.String(), ctx); err != nil {
			return err
		}
	}
	return nil
}

// populatePropertyOnTypes populates the VocabularyType's 'Properties' and
// 'WithoutProperties' fields based on the 'Domain' and 'DoesNotApplyTo'.
func populatePropertyOnTypes(registry *RDFRegistry, p VocabularyProperty, vocabName string, ctx *ParsingContext) error {
	ref := VocabularyReference{
		Name: p.Name,
		URI:  p.URI,
		// Vocab will only be populated on types outside of its own
		// vocabulary.
	}
	for _, d := range p.Domain {
		if len(d.Vocab) == 0 {
			t, ok := ctx.Result.Vocab.Types[d.Name]
			if !ok {
				return fmt.Errorf("cannot populate property on type %q for desired vocab", d.Name)
			}
			t.Properties = append(t.Properties, ref)
			ctx.Result.Vocab.Types[d.Name] = t
		} else {
			vocab := d.Vocab
			if u, err := registry.ResolveAlias(d.Vocab); err == nil {
				vocab = u
			}
			v, err := ctx.Result.GetReference(vocab)
			if err != nil {
				return err
			}
			t, ok := v.Types[d.Name]
			if !ok {
				return fmt.Errorf("cannot populate property on type %q for vocab %q", d.Name, vocab)
			}
			// Since the type is outside this property's vocabulary,
			// populate the Vocab field.
			refCopy := ref
			refCopy.Vocab = vocabName
			t.Properties = append(t.Properties, refCopy)
			v.Types[d.Name] = t
		}
	}
	for _, dna := range p.DoesNotApplyTo {
		if len(dna.Vocab) == 0 {
			t, ok := ctx.Result.Vocab.Types[dna.Name]
			if !ok {
				return fmt.Errorf("cannot populate withoutproperty on type %q for desired vocab", dna.Name)
			}
			t.WithoutProperties = append(t.WithoutProperties, ref)
			ctx.Result.Vocab.Types[dna.Name] = t
		} else {
			vocab := dna.Vocab
			if u, err := registry.ResolveAlias(dna.Vocab); err == nil {
				vocab = u
			}
			v, err := ctx.Result.GetReference(vocab)
			if err != nil {
				return err
			}
			t, ok := v.Types[dna.Name]
			if !ok {
				return fmt.Errorf("cannot populate withoutproperty on type %q for vocab %q", dna.Name, vocab)
			}
			// Since the type is outside this property's vocabulary,
			// populate the Vocab field.
			refCopy := ref
			refCopy.Vocab = vocabName
			t.WithoutProperties = append(t.WithoutProperties, refCopy)
			v.Types[dna.Name] = t
		}
	}
	return nil
}

// resolveReferences ensures that all references mentioned have been
// successfully parsed, and if not attempts to search the ontologies for any
// values, types, and properties that need to be referenced.
//
// Currently, this is the only way that values are added to the
// ParsedVocabulary.
func resolveReferences(registry *RDFRegistry, ctx *ParsingContext) error {
	vocabulary := ctx.Result
	for _, t := range vocabulary.Vocab.Types {
		for _, ref := range t.DisjointWith {
			if err := resolveReference(ref, registry, ctx); err != nil {
				return err
			}
		}
		for _, ref := range t.Extends {
			if err := resolveReference(ref, registry, ctx); err != nil {
				return err
			}
		}
	}
	for _, p := range vocabulary.Vocab.Properties {
		for _, ref := range p.Domain {
			if err := resolveReference(ref, registry, ctx); err != nil {
				return err
			}
		}
		for _, ref := range p.Range {
			if err := resolveReference(ref, registry, ctx); err != nil {
				return err
			}
		}
		for _, ref := range p.DoesNotApplyTo {
			if err := resolveReference(ref, registry, ctx); err != nil {
				return err
			}
		}
		if len(p.SubpropertyOf.Name) > 0 {
			if err := resolveReference(p.SubpropertyOf, registry, ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveReference will attempt to resolve the reference by either finding it
// in the known References of the vocabulary, or load it from the registry. Will
// fail if a reference is not found.
func resolveReference(reference VocabularyReference, registry *RDFRegistry, ctx *ParsingContext) error {
	name := reference.Name
	vocab := &ctx.Result.Vocab
	if len(reference.Vocab) > 0 {
		name = joinAlias(reference.Vocab, reference.Name)
		url, e := registry.ResolveAlias(reference.Vocab)
		if e != nil {
			return e
		}
		vocab, e = ctx.Result.GetReference(url)
		if e != nil {
			return e
		}
	}
	if _, ok := vocab.Types[reference.Name]; ok {
		return nil
	} else if _, ok := vocab.Properties[reference.Name]; ok {
		return nil
	} else if _, ok := vocab.Values[reference.Name]; ok {
		return nil
	} else if n, e := registry.getNode(name); e != nil {
		return e
	} else {
		applicable, e := n.Apply("", nil, ctx)
		if !applicable {
			return fmt.Errorf("cannot resolve reference with unapplicable node for %s", reference)
		} else if e != nil {
			return e
		}
		return nil
	}
}

// apply takes a specification input to populate the ParsingContext, based on
// the capabilities of the RDFNodes created from ontologies.
//
// This function will populate all non-value data in the Vocabulary. It does not
// populate the 'Properties' nor the 'WithoutProperties' fields on any
// VocabularyType.
func apply(nodes []RDFNode, input JSONLD, ctx *ParsingContext) error {
	// Hijacked processing: Process the rest of the data in this single
	// node.
	if ctx.OnlyApplyThisNode != nil {
		if applied, err := ctx.OnlyApplyThisNode.Apply("", input, ctx); !applied {
			return fmt.Errorf("applying requested node failed")
		} else {
			return err
		}
	}
	// Special processing: '@type' or 'type' if they are present
	if v, ok := input[JSON_LD_TYPE]; ok {
		if err := doApply(nodes, JSON_LD_TYPE, v, ctx); err != nil {
			return err
		}
	} else if v, ok := input[JSON_LD_TYPE_AS]; ok {
		if err := doApply(nodes, JSON_LD_TYPE_AS, v, ctx); err != nil {
			return err
		}
	}
	// Normal recursive processing
	for k, v := range input {
		// Skip things we have already processed: context and type
		if k == JSON_LD_CONTEXT {
			continue
		} else if k == JSON_LD_TYPE {
			continue
		} else if k == JSON_LD_TYPE_AS {
			continue
		}
		if err := doApply(nodes, k, v, ctx); err != nil {
			return err
		}
	}
	return nil
}

// doApply actually does the application logic for the apply function.
func doApply(nodes []RDFNode,
	k string, v interface{},
	ctx *ParsingContext) error {
	// Hijacked processing: Only use the ParsingContext's node to
	// handle all elements.
	recurNodes := nodes
	enterApplyExitNodes, clearFn := ctx.GetNextNodes(nodes)
	defer clearFn()
	// Normal recursive processing
	if mapValue, ok := v.(map[string]interface{}); ok {
		if err := enterFirstNode(enterApplyExitNodes, k, ctx); err != nil {
			return err
		} else if err = apply(recurNodes, mapValue, ctx); err != nil {
			return err
		} else if err = exitFirstNode(enterApplyExitNodes, k, ctx); err != nil {
			return err
		}
	} else if arrValue, ok := v.([]interface{}); ok {
		for _, val := range arrValue {
			// First, enter for this key
			if err := enterFirstNode(enterApplyExitNodes, k, ctx); err != nil {
				return err
			}
			// Recur or handle the value as necessary.
			if mapValue, ok := val.(map[string]interface{}); ok {
				if err := apply(recurNodes, mapValue, ctx); err != nil {
					return err
				}
			} else if err := applyFirstNode(enterApplyExitNodes, k, val, ctx); err != nil {
				return err
			}
			// Finally, exit for this key
			if err := exitFirstNode(enterApplyExitNodes, k, ctx); err != nil {
				return err
			}
		}
	} else if err := applyFirstNode(enterApplyExitNodes, k, v, ctx); err != nil {
		return err
	}
	return nil
}

// enterFirstNode will Enter the first RDFNode that returns true or an error.
func enterFirstNode(nodes []RDFNode, key string, ctx *ParsingContext) error {
	for _, node := range nodes {
		if applied, err := node.Enter(key, ctx); applied {
			return err
		} else if err != nil {
			return err
		}
	}
	return fmt.Errorf("no RDFNode applicable for entering %q", key)
}

// exitFirstNode will Exit the first RDFNode that returns true or an error.
func exitFirstNode(nodes []RDFNode, key string, ctx *ParsingContext) error {
	for _, node := range nodes {
		if applied, err := node.Exit(key, ctx); applied {
			return err
		} else if err != nil {
			return err
		}
	}
	return fmt.Errorf("no RDFNode applicable for exiting %q", key)
}

// applyFirstNode will Apply the first RDFNode that returns true or an error.
func applyFirstNode(nodes []RDFNode, key string, value interface{}, ctx *ParsingContext) error {
	for _, node := range nodes {
		if applied, err := node.Apply(key, value, ctx); applied {
			return err
		} else if err != nil {
			return err
		}
	}
	return fmt.Errorf("no RDFNode applicable for applying %q with value %v", key, value)
}

// parseJSONLDContext implements a super basic JSON-LD @context parsing
// algorithm in order to build a set of nodes which will be able to parse the
// rest of the document.
func parseJSONLDContext(registry *RDFRegistry, input JSONLD) (nodes []RDFNode, err error) {
	i, ok := input[JSON_LD_CONTEXT]
	if !ok {
		err = fmt.Errorf("no @context in input")
		return
	}
	if inArray, ok := i.([]interface{}); ok {
		// @context is an array
		for _, iVal := range inArray {
			if valMap, ok := iVal.(map[string]interface{}); ok {
				// Element is a JSON Object (dictionary)
				for alias, val := range valMap {
					if s, ok := val.(string); ok {
						var n []RDFNode
						n, err = registry.getAliased(alias, s)
						if err != nil {
							return
						}
						nodes = append(nodes, n...)
					} else if aliasedMap, ok := val.(map[string]interface{}); ok {
						var n []RDFNode
						n, err = registry.getAliasedObject(alias, aliasedMap)
						if err != nil {
							return
						}
						nodes = append(nodes, n...)
					} else {
						err = fmt.Errorf("@context value in dict in array is neither a dict nor a string")
						return
					}
				}
			} else if s, ok := iVal.(string); ok {
				// Element is a single value
				var n []RDFNode
				n, err = registry.getFor(s)
				if err != nil {
					return
				}
				nodes = append(nodes, n...)
			} else {
				err = fmt.Errorf("@context value in array is neither a dict nor a string")
				return
			}
		}
	} else if inMap, ok := i.(map[string]interface{}); ok {
		// @context is a JSON object (dictionary)
		for alias, iVal := range inMap {
			if s, ok := iVal.(string); ok {
				var n []RDFNode
				n, err = registry.getAliased(alias, s)
				if err != nil {
					return
				}
				nodes = append(nodes, n...)
			} else if aliasedMap, ok := iVal.(map[string]interface{}); ok {
				var n []RDFNode
				n, err = registry.getAliasedObject(alias, aliasedMap)
				if err != nil {
					return
				}
				nodes = append(nodes, n...)
			} else {
				err = fmt.Errorf("@context value in dict is neither a dict nor a string")
				return
			}
		}
	} else {
		// @context is a single value
		s, ok := i.(string)
		if !ok {
			err = fmt.Errorf("single @context value is not a string")
			return
		}
		return registry.getFor(s)
	}
	return
}
