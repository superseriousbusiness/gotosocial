package streams

import (
	"github.com/go-fed/activity/streams/vocab"
)

const (
	// jsonLDContext is the key for the JSON-LD specification's context
	// value. It contains the definitions of the types contained within the
	// rest of the payload. Important for linked-data representations, but
	// only applicable to go-fed at code-generation time.
	jsonLDContext = "@context"
)

// Serialize adds the context vocabularies contained within the type
// into the JSON-LD @context field, and aliases them appropriately.
func Serialize(a vocab.Type) (m map[string]interface{}, e error) {
	m, e = a.Serialize()
	if e != nil {
		return
	}
	v := a.JSONLDContext()
	// Transform the map of vocabulary-to-aliases into a context payload,
	// but do so in a way that at least keeps it readable for other humans.
	var contextValue interface{}
	if len(v) == 1 {
		for vocab, alias := range v {
			if len(alias) == 0 {
				contextValue = vocab
			} else {
				contextValue = map[string]string{
					alias: vocab,
				}
			}
		}
	} else {
		var arr []interface{}
		aliases := make(map[string]string)
		for vocab, alias := range v {
			if len(alias) == 0 {
				arr = append(arr, vocab)
			} else {
				aliases[alias] = vocab
			}
		}
		if len(aliases) > 0 {
			arr = append(arr, aliases)
		}
		contextValue = arr
	}
	// TODO: Update the context instead if it already exists
	m[jsonLDContext] = contextValue
	// TODO: Sort the context based on arbitrary order.
	// Delete any existing `@context` in child maps.
	var cleanFnRecur func(map[string]interface{})
	cleanFnRecur = func(r map[string]interface{}) {
		for _, v := range r {
			if n, ok := v.(map[string]interface{}); ok {
				delete(n, jsonLDContext)
				cleanFnRecur(n)
			}
		}
	}
	cleanFnRecur(m)
	return
}
