# streams

ActivityStreams vocabularies automatically code-generated with `astool`.

## Reference & Tutorial

The [go-fed website](https://go-fed.org/) contains tutorials and reference
materials, in addition to the rest of this README.

## How To Use

```
go get github.com/go-fed/activity
```

All generated types and properties are interfaces in
`github.com/go-fed/streams/vocab`, but note that the constructors and supporting
functions live in `github.com/go-fed/streams`.

To create a type and set properties:

```golang
var actorURL *url.URL = // ...

// A new "Create" Activity.
create := streams.NewActivityStreamsCreate()
// A new "actor" property.
actor := streams.NewActivityStreamsActorProperty()
actor.AppendIRI(actorURL)
// Set the "actor" property on the "Create" Activity.
create.SetActivityStreamsActor(actor)
```

To process properties on a type:

```golang
// Returns true if the "Update" has at least one "object" with an IRI value.
func hasObjectWithIRIValue(update vocab.ActivityStreamsUpdate) bool {
  objectProperty := update.GetActivityStreamsObject()
  // Any property may be nil if it was either empty in the original JSON or
  // never set on the golang type.
  if objectProperty == nil {
    return false
  }
  // The "object" property is non-functional: it could have multiple values. The
  // generated code has slightly different methods for a functional property
  // versus a non-functional one.
  //
  // While it may be easy to ignore multiple values in other languages
  // (accidentally or purposefully), go-fed is designed to make it hard to do
  // so.
  for iter := objectProperty.Begin(); iter != objectProperty.End(); iter = iter.Next() {
    // If this particular value is an IRI, return true.
    if iter.IsIRI() {
      return true
    }
  }
  // All values are literal embedded values and not IRIs.
  return false
}
```

The ActivityStreams type hierarchy of "extends" and "disjoint" is not the same
as the Object Oriented definition of inheritance. It is also not the same as
golang's interface duck-typing. Helper functions are provided to guarantee that
an application's logic can correctly apply the type hierarchy.

```golang
thing := // Pick a type from streams.NewActivityStreams<Type>()
if streams.ActivityStreamsObjectIsDisjointWith(thing) {
  fmt.Printf("The \"Object\" type is Disjoint with the %T type.\n", thing)
}
if streams.ActivityStreamsLinkIsExtendedBy(thing) {
  fmt.Printf("The %T type Extends from the \"Link\" type.\n", thing)
}
if streams.ActivityStreamsActivityExtends(thing) {
  fmt.Printf("The \"Activity\" type extends from the %T type.\n", thing)
}
```

When given a generic JSON payload, it can be resolved to a concrete type by
creating a `streams.JSONResolver` and giving it a callback function that accepts
the interesting concrete type:

```golang
// Callbacks must be in the form:
//   func(context.Context, <TypeInterface>) error
createCallback := func(c context.Context, create vocab.ActivityStreamsCreate) error {
  // Do something with 'create'
  fmt.Printf("createCallback called: %T\n", create)
  return nil
}
updateCallback := func(c context.Context, update vocab.ActivityStreamsUpdate) error {
  // Do something with 'update'
  fmt.Printf("updateCallback called: %T\n", update)
  return nil
}
jsonResolver, err := streams.NewJSONResolver(createCallback, updateCallback)
if err != nil {
  // Something in the setup was wrong. For example, a callback has an
  // unsupported signature and would never be called
  panic(err)
}
// Create a context, which allows you to pass data opaquely through the
// JSONResolver.
c := context.Background()
// Example 15 of the ActivityStreams specification.
b := []byte(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally created a note",
  "type": "Create",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Note",
    "name": "A Simple Note",
    "content": "This is a simple note"
  }
}`)
var jsonMap map[string]interface{}
if err = json.Unmarshal(b, &jsonMap); err != nil {
  panic(err)
}
// The createCallback function will be called.
err = jsonResolver.Resolve(c, jsonMap)
if err != nil && !streams.IsUnmatchedErr(err) {
  // Something went wrong
  panic(err)
} else if streams.IsUnmatchedErr(err) {
  // Everything went right but the callback didn't match or the ActivityStreams
  // type is one that wasn't code generated.
  fmt.Println("No match: ", err)
}
```

A `streams.TypeResolver` is similar but uses the golang types instead. It
accepts the generic `vocab.Type`. This is the abstraction when needing to handle
any ActivityStreams type. The function `ToType` can convert a JSON-decoded-map
into this kind of value if needed.

A `streams.PredicatedTypeResolver` lets you apply a boolean predicate function
that acts as a check whether a callback is allowed to be invoked.

## FAQ

### Why Are Empty Properties Nil And Not Zero-Valued?

Due to implementation design decisions, it would require a lot of plumbing to
ensure this would work properly. It would also require allocation of a
non-trivial amount of memory.
