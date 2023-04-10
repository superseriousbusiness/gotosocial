package httpclient

import "net/http"

// SignFunc is a function signature that provides request signing.
type SignFunc func(r *http.Request) error

type SigningClient interface {
	Do(r *http.Request) (*http.Response, error)
	DoSigned(r *http.Request, sign SignFunc) (*http.Response, error)
}
