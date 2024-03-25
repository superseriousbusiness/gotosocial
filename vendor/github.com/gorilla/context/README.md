# gorilla/context

[![License](https://img.shields.io/github/license/gorilla/.github)](https://img.shields.io/github/license/gorilla/.github)
![testing](https://github.com/gorilla/context/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/github/gorilla/context/branch/main/graph/badge.svg)](https://codecov.io/github/gorilla/context)
[![godoc](https://godoc.org/github.com/gorilla/context?status.svg)](https://godoc.org/github.com/gorilla/context)
[![sourcegraph](https://sourcegraph.com/github.com/gorilla/context/-/badge.svg)](https://sourcegraph.com/github.com/gorilla/context?badge)
[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/7656/badge)](https://bestpractices.coreinfrastructure.org/projects/7656)

![Gorilla Logo](https://github.com/gorilla/.github/assets/53367916/d92caabf-98e0-473e-bfbf-ab554ba435e5)

> ⚠⚠⚠ **Note** ⚠⚠⚠ gorilla/context, having been born well before `context.Context` existed, does not play well
> with the shallow copying of the request that [`http.Request.WithContext`](https://golang.org/pkg/net/http/#Request.WithContext) (added to net/http Go 1.7 onwards) performs.
>
> Using gorilla/context may lead to memory leaks under those conditions, as the pointers to each `http.Request` become "islanded" and will not be cleaned up when the response is sent.
>
> You should use the `http.Request.Context()` feature in Go 1.7.

gorilla/context is a general purpose registry for global request variables.

* It stores a `map[*http.Request]map[interface{}]interface{}` as a global singleton, and thus tracks variables by their HTTP request.


### License

See the LICENSE file for details.
