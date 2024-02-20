package schema

import (
	"fmt"
	neturl "net/url"
	"strings"

	"github.com/superseriousbusiness/activity/astool/rdf"
)

const (
	schemaSpec       = "http://schema.org/"
	exampleSpec      = "workExample"
	mainEntitySpec   = "mainEntity"
	urlSpec          = "URL"
	nameSpec         = "name"
	creativeWorkSpec = "CreativeWork"
)

// SchemaOntology represents Ontologies from schema.org.
type SchemaOntology struct{}

// SpecURI returns the Schema.org URI.
func (o *SchemaOntology) SpecURI() string {
	return schemaSpec
}

// Load without an alias.
func (o *SchemaOntology) Load() ([]rdf.RDFNode, error) {
	return o.LoadAsAlias("")
}

// LoadAsAlias loads with an alias.
func (o *SchemaOntology) LoadAsAlias(s string) ([]rdf.RDFNode, error) {
	return []rdf.RDFNode{
		&rdf.AliasedDelegate{
			Spec:     schemaSpec,
			Alias:    s,
			Name:     exampleSpec,
			Delegate: &example{},
		},
		&rdf.AliasedDelegate{
			Spec:     schemaSpec,
			Alias:    s,
			Name:     mainEntitySpec,
			Delegate: &mainEntity{},
		},
		&rdf.AliasedDelegate{
			Spec:     schemaSpec,
			Alias:    s,
			Name:     urlSpec,
			Delegate: &url{},
		},
		&rdf.AliasedDelegate{
			Spec:     schemaSpec,
			Alias:    s,
			Name:     nameSpec,
			Delegate: &name{},
		},
		&rdf.AliasedDelegate{
			Spec:     schemaSpec,
			Alias:    s,
			Name:     creativeWorkSpec,
			Delegate: &creativeWork{},
		},
	}, nil
}

// LoadSpecificAsAlias loads a specific node and aliases it.
func (o *SchemaOntology) LoadSpecificAsAlias(alias, n string) ([]rdf.RDFNode, error) {
	switch n {
	case exampleSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &example{},
			},
		}, nil
	case mainEntitySpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &mainEntity{},
			},
		}, nil
	case urlSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &url{},
			},
		}, nil
	case nameSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &name{},
			},
		}, nil
	case creativeWorkSpec:
		return []rdf.RDFNode{
			&rdf.AliasedDelegate{
				Spec:     "",
				Alias:    "",
				Name:     alias,
				Delegate: &creativeWork{},
			},
		}, nil
	}
	return nil, fmt.Errorf("schema ontology cannot find %q to alias to %q", n, alias)
}

// LoadElement does nothing.
func (o *SchemaOntology) LoadElement(name string, payload map[string]interface{}) ([]rdf.RDFNode, error) {
	return nil, nil
}

// GetByName returns a bare node by name.
func (o *SchemaOntology) GetByName(n string) (rdf.RDFNode, error) {
	n = strings.TrimPrefix(n, o.SpecURI())
	switch n {
	case exampleSpec:
		return &example{}, nil
	case mainEntitySpec:
		return &mainEntity{}, nil
	case urlSpec:
		return &url{}, nil
	case nameSpec:
		return &name{}, nil
	case creativeWorkSpec:
		return &creativeWork{}, nil
	}
	return nil, fmt.Errorf("schema ontology could not find node for name %s", n)
}

var _ rdf.RDFNode = &example{}

// example is best understood by giving an example, such as this.
type example struct{}

// Enter Pushes an Example as Current.
func (e *example) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Push()
	ctx.Current = &rdf.VocabularyExample{}
	return true, nil
}

// Exit Pops an Example and sets it on the parent item.
//
// Exit returns an error if the popped item is not an Example, or if after
// popping the Current item cannot have an Example added to it.
func (e *example) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	ei := ctx.Current
	ctx.Pop()
	if ve, ok := ei.(*rdf.VocabularyExample); !ok {
		return true, fmt.Errorf("schema example did not pop a *VocabularyExample")
	} else if ea, ok := ctx.Current.(rdf.ExampleAdder); !ok {
		return true, fmt.Errorf("schema example not given an ExampleAdder")
	} else {
		ea.AddExample(ve)
	}
	return true, nil
}

// Apply returns an error.
func (e *example) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema example cannot be applied")
}

var _ rdf.RDFNode = &mainEntity{}

// mainEntity reapplies itself in all sublevels and simply saves the value onto
// Current. This saves the JSON example in raw form.
type mainEntity struct{}

// Enter Pushes the Current item and tells the context to only apply itself for
// all sublevels.
func (m *mainEntity) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Push()
	ctx.SetOnlyApplyThisNode(m)
	return true, nil
}

// Exit saves the current raw JSON example onto a parent Example.
//
// Exit reutrns an error if Current after popping is not an Example.
func (m *mainEntity) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	// Save the example
	example := ctx.Current
	// Undo the Enter operations
	ctx.ResetOnlyApplyThisNode()
	ctx.Pop()
	// Set the example data
	if vEx, ok := ctx.Current.(*rdf.VocabularyExample); !ok {
		return true, fmt.Errorf("mainEntity exit not given a *VocabularyExample")
	} else {
		vEx.Example = example
	}
	return true, nil
}

// Apply simply saves the value onto Current.
func (m *mainEntity) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	ctx.Current = value
	return true, nil
}

var _ rdf.RDFNode = &url{}

// url sets the URI on an item.
type url struct{}

// Enter does nothing.
func (u *url) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema url cannot be entered")
}

// Exit does nothing.
func (u *url) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema url cannot be exited")
}

// Apply sets the value as a URI onto an item.
//
// Returns an error if the value is not a string, or it cannot set the URI on
// the Current item.
func (u *url) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	if urlString, ok := value.(string); !ok {
		return true, fmt.Errorf("schema url not given a string")
	} else if uriSetter, ok := ctx.Current.(rdf.URISetter); !ok {
		return true, fmt.Errorf("schema url not given a URISetter in context")
	} else {
		return true, uriSetter.SetURI(urlString)
	}
}

var _ rdf.RDFNode = &name{}

// name sets the Name on an item.
type name struct{}

// Enter does nothing.
func (n *name) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema name cannot be entered")
}

// Exit does nothing.
func (n *name) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema name cannot be exited")
}

// Apply sets the value as a name on the Current item.
//
// Returns an error if the value is not a string, or if the Current item cannot
// have its name set.
func (n *name) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	if s, ok := value.(string); !ok {
		return true, fmt.Errorf("schema name not given string")
	} else if ns, ok := ctx.Current.(rdf.NameSetter); !ok {
		return true, fmt.Errorf("schema name not given NameSetter in context")
	} else {
		var vocab string
		// Parse will interpret "ActivityStreams" as a valid URL without
		// a scheme. It will also interpret "as:Object" as a valid URL
		// with a scheme of "as".
		if u, err := neturl.Parse(s); err == nil && len(u.Scheme) > 0 && len(u.Host) > 0 {
			// If the name is a URL, use heuristics to determine the
			// name versus vocabulary part.
			//
			// The vocabulary is usually the URI without the
			// fragment or final path entry. The name is usually the
			// fragment or final path entry.
			if len(u.Fragment) > 0 {
				// Attempt to parse the fragment
				s = u.Fragment
				u.Fragment = ""
				vocab = u.String()
			} else {
				// Use the final path component
				comp := strings.Split(s, "/")
				s = comp[len(comp)-1]
				vocab = strings.Join(comp[:len(comp)-1], "/")
			}
		} else if sp := rdf.SplitAlias(s); len(sp) == 2 {
			// The name may be aliased.
			vocab = sp[0]
			s = sp[1]
		} // Else the name has no vocabulary reference.
		if len(vocab) > 0 {
			if ref, ok := ctx.Current.(*rdf.VocabularyReference); !ok {
				return true, fmt.Errorf("schema name not given *rdf.VocabularyReference in context")
			} else {
				ref.Vocab = vocab
			}
		}
		ns.SetName(s)
		ctx.Name = s
		return true, nil
	}
}

var _ rdf.RDFNode = &creativeWork{}

// creativeWork does nothing.
type creativeWork struct{}

// Enter returns an error.
func (c *creativeWork) Enter(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema creative work cannot be entered")
}

// Exit does nothing.
func (c *creativeWork) Exit(key string, ctx *rdf.ParsingContext) (bool, error) {
	return true, fmt.Errorf("schema creative work cannot be exited")
}

// Apply does nothing.
func (c *creativeWork) Apply(key string, value interface{}, ctx *rdf.ParsingContext) (bool, error) {
	// Do nothing -- should already be an example.
	return true, nil
}
