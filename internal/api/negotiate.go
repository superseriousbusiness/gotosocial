package api

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Offer string

const (
	OfferAppJSON         Offer = `application/json`
	OfferAppActivityJson Offer = `application/activity+json`
	OfferAppLDJson       Offer = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
)

// ActivityPubAcceptHeaders represents the Accept headers mentioned here:
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubAcceptHeaders = []Offer{
	OfferAppActivityJson,
	OfferAppLDJson,
}

func NegotiateAccept(c *gin.Context, offers []Offer) (string, error) {
	if len(offers) == 0 {
		return "", errors.New("no format offered")
	}

	accepts := c.Request.Header.Values("Accept")
	if len(accepts) == 0 {
		return "", errors.New("no Accept header(s) set on incoming request")
	}

	strings := []string{}
	for _, o := range offers {
		strings = append(strings, string(o))
	}

	format := c.NegotiateFormat(strings...)
	if format == "" {
		return "", fmt.Errorf("no format can be offered for Accept header(s) %s; we offered %s", accepts, offers)
	}

	return format, nil
}
