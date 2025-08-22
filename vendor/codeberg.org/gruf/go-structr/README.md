# go-structr

A library with a series of performant data types with automated struct value indexing. Indexing is supported via arbitrary combinations of fields, and in the case of the cache type, negative results (errors!) are also supported.

Under the hood, go-structr maintains a hashmap per index, where each hashmap is keyed by serialized input key. This is handled by the incredibly performant serialization library [go-mangler/v2](https://codeberg.org/gruf/go-mangler), which at this point in time supports all concrete types, so feel free to index by by *almost* anything!

See the [docs](https://pkg.go.dev/codeberg.org/gruf/go-structr) for more API information.

## Notes

This is a core underpinning of [GoToSocial](https://github.com/superseriousbusiness/gotosocial)'s performance.