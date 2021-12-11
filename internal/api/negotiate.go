package api

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type offer string

const (
	AppJSON           offer = `application/json`                                                     // AppJSON is the mime type for 'application/json'.
	AppActivityJSON   offer = `application/activity+json`                                            // AppActivityJSON is the mime type for 'application/activity+json'.
	AppActivityLDJSON offer = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"` // AppActivityLDJSON is the mime type for 'application/ld+json; profile="https://www.w3.org/ns/activitystreams"'
	TextHTML          offer = `text/html`                                                            // TextHTML is the mime type for 'text/html'.
)

// ActivityPubAcceptHeaders represents the Accept headers mentioned here:
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubAcceptHeaders = []offer{
	AppActivityJSON,
	AppActivityLDJSON,
}

// JSONAcceptHeaders is a slice of offers that just contains application/json types.
var JSONAcceptHeaders = []offer{
	AppJSON,
}

// HTMLAcceptHeaders is a slice of offers that just contains text/html types.
var HTMLAcceptHeaders = []offer{
	TextHTML,
}

// NegotiateAccept takes the *gin.Context from an incoming request, and a
// slice of Offers, and performs content negotiation for the given request
// with the given content-type offers. It will return a string representation
// of the first suitable content-type, or an error if something goes wrong or
// a suiteable content-type cannot be matched.
//
// For example, if the request in the *gin.Context has Accept headers of value
// [application/json, text/html], and the provided offers are of value
// [application/json, application/xml], then the returned string will be
// 'application/json', which indicates the content-type that should be returned.
//
// If there are no Accept headers in the request, or the length of offers is 0,
// then an error will be returned, so this function should only be called in places
// where format negotiation is actually needed and headers are expected to be present
// on incoming requests.
//
// Callers can use the offer slices exported in this package as shortcuts for
// often-used Accept types.
//
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation#server-driven_content_negotiation
func NegotiateAccept(c *gin.Context, offers ...offer) (string, error) {
	if len(offers) == 0 {
		return "", errors.New("no format offered")
	}

	accepts := c.Request.Header.Values("Accept")
	if len(accepts) == 0 {
		return "", fmt.Errorf("no Accept header(s) set on incoming request; this endpoint offers %s", offers)
	}

	strings := []string{}
	for _, o := range offers {
		strings = append(strings, string(o))
	}

	format := c.NegotiateFormat(strings...)
	if format == "" {
		return "", fmt.Errorf("no format can be offered for requested Accept header(s) %s; this endpoint offers %s", accepts, offers)
	}

	return format, nil
}
