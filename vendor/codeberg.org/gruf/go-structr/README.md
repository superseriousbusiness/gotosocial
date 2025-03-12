# go-structr

A library with a series of performant data types with automated struct value indexing. Indexing is supported via arbitrary combinations of fields, and in the case of the cache type, negative results (errors!) are also supported.

Under the hood, go-structr maintains a hashmap per index, where each hashmap is a hashmap keyed by serialized input key type. This is handled by the incredibly performant serialization library [go-mangler](https://codeberg.org/gruf/go-mangler), which at this point in time supports *most* arbitrary types (other than maps, channels, functions), so feel free to index by by almost *anything*!

See the [docs](https://pkg.go.dev/codeberg.org/gruf/go-structr) for more API information.

## Notes

This is a core underpinning of [GoToSocial](https://github.com/superseriousbusiness/gotosocial)'s performance.