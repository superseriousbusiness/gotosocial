package streams

import (
	"maps"
	"slices"

	"code.superseriousbusiness.org/activity/streams/vocab"
)

const (
	// jsonLDContext is the key for the JSON-LD specification's context
	// value. It contains the definitions of the types contained within the
	// rest of the payload. Important for linked-data representations, but
	// only applicable to go-fed at code-generation time.
	jsonLDContext = "@context"

	asNS     = "https://www.w3.org/ns/activitystreams"
	tootNS   = "http://joinmastodon.org/ns"
	schemaNS = "http://schema.org"
)

// Map of inlines @context entries that may need to be added
// when vocabs include "https://www.w3.org/ns/activitystreams".
var asInlines = map[string]any{
	"Hashtag":                   "as:Hashtag",
	"alsoKnownAs":               "as:alsoKnownAs",
	"manuallyApprovesFollowers": "as:manuallyApprovesFollowers",
	"sensitive":                 "as:sensitive",

	"movedTo": map[string]string{
		"@id":   "as:movedTo",
		"@type": "@id",
	},
}

// Map of inlines @context entries that may need to be
// added when vocabs include "http://joinmastodon.org/ns".
var tootInlines = map[string]any{
	"Emoji":        "toot:Emoji",
	"blurhash":     "toot:blurhash",
	"discoverable": "toot:discoverable",
	"indexable":    "toot:indexable",
	"memorial":     "toot:memorial",
	"suspended":    "toot:suspended",
	"votersCount":  "toot:votersCount",

	"featured": map[string]string{
		"@id":   "toot:featured",
		"@type": "@id",
	},

	"featuredTags": map[string]string{
		"@id":   "toot:featuredTags",
		"@type": "@id",
	},

	"focalPoint": map[string]string{
		"@container": "@list",
		"@id":        "toot:focalPoint",
	},
}

// Map of inlines @context entries that may need to
// be added when vocabs include "http://schema.org".
var schemaInlines = map[string]any{
	"PropertyValue": "schema:PropertyValue",
	"value":         "schema:value",
}

// getLookup returns a lookup map of all interesting field names
// + type names on the given "in" map that may need to be inlined.
func getLookup(in map[string]any) map[string]struct{} {
	out := make(map[string]struct{})

	for k, v := range in {
		// Pull out keys from any nested maps.
		if nested, ok := v.(map[string]any); ok {
			maps.Copy(out, getLookup(nested))
			continue
		}

		// Pull out keys from any
		// arrays of nested maps.
		if nestedIs, ok := v.([]any); ok {
			for _, nestedI := range nestedIs {
				if nested, ok := nestedI.(map[string]any); ok {
					maps.Copy(out, getLookup(nested))
					continue
				}
			}
		}

		// For types, we actually care about
		// the *value*, ie., the name of the
		// type, not the type key itself.
		if k == "type" {
			out[v.(string)] = struct{}{}
			continue
		}

		out[k] = struct{}{}
	}

	return out
}

func copyInlines(
	src map[string]any,
	dst map[string]any,
	lookup map[string]struct{},
) {
	for k, v := range src {
		_, ok := lookup[k]
		if ok {
			dst[k] = v
		}
	}
}

// Serialize adds the context vocabularies contained within the type
// into the JSON-LD @context field, and aliases them appropriately.
func Serialize(a vocab.Type) (m map[string]any, e error) {
	m, e = a.Serialize()
	if e != nil {
		return
	}

	var (
		// Slice of vocab URIs
		// used in this vocab.Type.
		vocabs = a.JSONLDContext()

		// Slice of vocab URIs to add
		// to the base @context slice.
		includeVocabs []string

		// Object to inline as an extra
		// entry in the @context slice.
		inlinedContext = make(map[string]any)
	)

	// Get a lookup of all field and
	// type names we need to care about.
	lookup := getLookup(m)

	// Go through each used vocab and see
	// if we need to special case it.
	for vocab := range vocabs {

		switch vocab {

		case asNS:
			// ActivityStreams vocab.
			//
			// The namespace URI already points to
			// a proper @context document but we
			// need to add some extra inlines.
			includeVocabs = append(includeVocabs, asNS)
			copyInlines(asInlines, inlinedContext, lookup)

		case schemaNS:
			// Schema vocab.
			//
			// The URI doesn't point to a @context
			// document so we need to inline everything.
			inlinedContext["schema"] = schemaNS + "#"
			copyInlines(schemaInlines, inlinedContext, lookup)

		case tootNS:
			// Toot/Mastodon vocab.
			//
			// The URI doesn't point to a @context
			// document so we need to inline everything.
			inlinedContext["toot"] = tootNS + "#"
			copyInlines(tootInlines, inlinedContext, lookup)

		default:
			// No special case.
			includeVocabs = append(includeVocabs, vocab)
		}
	}

	// Sort used vocab entries alphabetically
	// to make their ordering predictable.
	slices.Sort(includeVocabs)

	// Create final slice of @context
	// entries we'll need to include.
	contextEntries := make([]any, 0, len(includeVocabs)+1)

	// Append each included vocab to the slice.
	for _, vocab := range includeVocabs {
		contextEntries = append(contextEntries, vocab)
	}

	// Append any inlinedContext to the slice.
	if len(inlinedContext) != 0 {
		contextEntries = append(contextEntries, inlinedContext)
	}

	// Include @context on the final output,
	// using an array if there's more than
	// one entry, just a property otherwise.
	if len(contextEntries) != 1 {
		m[jsonLDContext] = contextEntries
	} else {
		m[jsonLDContext] = contextEntries[0]
	}

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
