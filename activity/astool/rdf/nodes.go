package rdf

var _ RDFNode = &AliasedDelegate{}

// AliasedDelegate will call the delegated RDFNode if the passed in keys
// conform to either the spec or alias.
type AliasedDelegate struct {
	Spec     string
	Alias    string
	Name     string
	Delegate RDFNode
}

// Enter calls the Delegate's Enter if the key conforms to the Spec or Alias.
func (a *AliasedDelegate) Enter(key string, ctx *ParsingContext) (bool, error) {
	if IsKeyApplicable(key, a.Spec, a.Alias, a.Name) {
		return a.Delegate.Enter(key, ctx)
	}
	return false, nil
}

// Exit calls the Delegate's Exit if the key conforms to the Spec or Alias.
func (a *AliasedDelegate) Exit(key string, ctx *ParsingContext) (bool, error) {
	if IsKeyApplicable(key, a.Spec, a.Alias, a.Name) {
		return a.Delegate.Exit(key, ctx)
	}
	return false, nil
}

// Apply calls the Delegate's Apply if the key conforms to the Spec or Alias.
func (a *AliasedDelegate) Apply(key string, value interface{}, ctx *ParsingContext) (bool, error) {
	if IsKeyApplicable(key, a.Spec, a.Alias, a.Name) {
		return a.Delegate.Apply(key, value, ctx)
	}
	return false, nil
}
