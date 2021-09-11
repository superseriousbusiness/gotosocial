Variable size bufferpool that supports storing buffers up to 512kb in size

See documentation for more information: https://godocs.io/git.iim.gay/grufwub/go-bufpool

Please note, the test here is a worst-case scenario for allocations (the size
requests always increase so a new slice is always required)