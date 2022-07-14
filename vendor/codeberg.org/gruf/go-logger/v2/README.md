# go-logger

Logger aims to be a highly performant, levelled logging package with a simplistic API at first glance, but with the tools for much greater customization if necessary.

This was born after my own increasing frustrations with overly complex logging packages when often I just want the simple `Print()`, `Info()`, `Error()`. That's not to say they are bad packages, but the complexity of their APIs increases time required to adjust or later move to another. Not to mention being less enjoyable to use!

If you have highly specific logging needs, e.g. in a production environment with log parsing utilities, then using an existing solution e.g. `zerolog` is a good idea.

If you're looking for levelled logging that doesn't dump a kitchen sink in your lap, but keeps the parts around should you want it, then give this a try :)

Pass `-tags=kvformat` to your build to use the custom formatter in `go-kv` when printing fields (see package for more details).

Import with `"codeberg.org/gruf/go-logger/v2"`. Only v2 is supported going forward.