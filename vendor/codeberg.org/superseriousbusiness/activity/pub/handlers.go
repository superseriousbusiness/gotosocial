package pub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"codeberg.org/superseriousbusiness/activity/streams"
)

var ErrNotFound = errors.New("go-fed/activity: ActivityStreams data not found")

// HandlerFunc determines whether an incoming HTTP request is an ActivityStreams
// GET request, and if so attempts to serve ActivityStreams data.
//
// If an error is returned, then the calling function is responsible for writing
// to the ResponseWriter as part of error handling.
//
// If 'isASRequest' is false and there is no error, then the calling function
// may continue processing the request, and the HandlerFunc will not have
// written anything to the ResponseWriter. For example, a webpage may be served
// instead.
//
// If 'isASRequest' is true and there is no error, then the HandlerFunc
// successfully served the request and wrote to the ResponseWriter.
//
// Callers are responsible for authorized access to this resource.
type HandlerFunc func(c context.Context, w http.ResponseWriter, r *http.Request) (isASRequest bool, err error)

// NewActivityStreamsHandler creates a HandlerFunc to serve ActivityStreams
// requests which are coming from other clients or servers that wish to obtain
// an ActivityStreams representation of data.
//
// Strips retrieved ActivityStreams values of sensitive fields ('bto' and 'bcc')
// before responding with them. Sets the appropriate HTTP status code for
// Tombstone Activities as well.
//
// Defaults to supporting content to be retrieved by HTTPS only.
func NewActivityStreamsHandler(db Database, clock Clock) HandlerFunc {
	return NewActivityStreamsHandlerScheme(db, clock, "https")
}

// NewActivityStreamsHandlerScheme creates a HandlerFunc to serve
// ActivityStreams requests which are coming from other clients or servers that
// wish to obtain an ActivityStreams representation of data provided by the
// specified protocol scheme.
//
// Strips retrieved ActivityStreams values of sensitive fields ('bto' and 'bcc')
// before responding with them. Sets the appropriate HTTP status code for
// Tombstone Activities as well.
//
// Specifying the "scheme" allows for retrieving ActivityStreams content with
// identifiers such as HTTP, HTTPS, or other protocol schemes.
//
// Returns ErrNotFound when the database does not retrieve any data and no
// errors occurred during retrieval.
func NewActivityStreamsHandlerScheme(db Database, clock Clock, scheme string) HandlerFunc {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) (isASRequest bool, err error) {
		// Do nothing if it is not an ActivityPub GET request
		if !isActivityPubGet(r) {
			return
		}
		isASRequest = true
		id := requestId(r, scheme)

		var unlock func()

		// Lock and obtain a copy of the requested ActivityStreams value
		unlock, err = db.Lock(c, id)
		if err != nil {
			return
		}
		// WARNING: Unlock not deferred
		t, err := db.Get(c, id)
		unlock() // unlock even on error
		if err != nil {
			return
		}
		// Unlock must have been called by this point and in every
		// branch above
		if t == nil {
			err = ErrNotFound
			return
		}
		// Remove sensitive fields.
		clearSensitiveFields(t)
		// Serialize the fetched value.
		m, err := streams.Serialize(t)
		if err != nil {
			return
		}
		raw, err := json.Marshal(m)
		if err != nil {
			return
		}
		// Construct the response.
		addResponseHeaders(w.Header(), clock, raw)
		// Write the response.
		if streams.IsOrExtendsActivityStreamsTombstone(t) {
			w.WriteHeader(http.StatusGone)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		n, err := w.Write(raw)
		if err != nil {
			return
		} else if n != len(raw) {
			err = fmt.Errorf("only wrote %d of %d bytes", n, len(raw))
			return
		}
		return
	}
}
