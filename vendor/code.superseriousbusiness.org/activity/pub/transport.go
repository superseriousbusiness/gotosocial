package pub

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"code.superseriousbusiness.org/httpsig"
)

const (
	// acceptHeaderValue is the Accept header value indicating that the
	// response should contain an ActivityStreams object.
	acceptHeaderValue = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
)

// isSuccess returns true if the HTTP status code is either OK, Created, or
// Accepted.
func isSuccess(code int) bool {
	return code == http.StatusOK ||
		code == http.StatusCreated ||
		code == http.StatusAccepted
}

// Transport makes ActivityStreams calls to other servers in order to send or
// receive ActivityStreams data.
//
// It is responsible for setting the appropriate request headers, signing the
// requests if needed, and facilitating the traffic between this server and
// another.
//
// The transport is exclusively used to issue requests on behalf of an actor,
// and is never sending requests on behalf of the server in general.
//
// It may be reused multiple times, but never concurrently.
type Transport interface {
	// Dereference fetches the ActivityStreams object located at this IRI with
	// a GET request. Note that Response will only be returned on status = OK.
	Dereference(c context.Context, iri *url.URL) (*http.Response, error)

	// Deliver sends an ActivityStreams object.
	Deliver(c context.Context, obj map[string]interface{}, to *url.URL) error

	// BatchDeliver sends an ActivityStreams object to multiple recipients.
	BatchDeliver(c context.Context, obj map[string]interface{}, recipients []*url.URL) error
}

// Transport must be implemented by HttpSigTransport.
var _ Transport = &HttpSigTransport{}

// HttpSigTransport makes a dereference call using HTTP signatures to
// authenticate the request on behalf of a particular actor.
//
// No rate limiting is applied.
//
// Only one request is tried per call.
type HttpSigTransport struct {
	client       HttpClient
	appAgent     string
	gofedAgent   string
	clock        Clock
	getSigner    httpsig.Signer
	getSignerMu  *sync.Mutex
	postSigner   httpsig.Signer
	postSignerMu *sync.Mutex
	pubKeyId     string
	privKey      crypto.PrivateKey
}

// NewHttpSigTransport returns a new Transport.
//
// It sends requests specifically on behalf of a specific actor on this server.
// The actor's credentials are used to add an HTTP Signature to requests, which
// requires an actor's private key, a unique identifier for their public key,
// and an HTTP Signature signing algorithm.
//
// The client lets users issue requests through any HTTP client, including the
// standard library's HTTP client.
//
// The appAgent uniquely identifies the calling application's requests, so peers
// may aid debugging the requests incoming from this server. Note that the
// agent string will also include one for go-fed, so at minimum peer servers can
// reach out to the go-fed library to aid in notifying implementors of malformed
// or unsupported requests.
func NewHttpSigTransport(
	client HttpClient,
	appAgent string,
	clock Clock,
	getSigner, postSigner httpsig.Signer,
	pubKeyId string,
	privKey crypto.PrivateKey) *HttpSigTransport {
	return &HttpSigTransport{
		client:       client,
		appAgent:     appAgent,
		gofedAgent:   goFedUserAgent(),
		clock:        clock,
		getSigner:    getSigner,
		getSignerMu:  &sync.Mutex{},
		postSigner:   postSigner,
		postSignerMu: &sync.Mutex{},
		pubKeyId:     pubKeyId,
		privKey:      privKey,
	}
}

// Dereference sends a GET request signed with an HTTP Signature to obtain an ActivityStreams value.
func (h HttpSigTransport) Dereference(c context.Context, iri *url.URL) (*http.Response, error) {
	req, err := http.NewRequest("GET", iri.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(c)
	req.Header.Add(acceptHeader, acceptHeaderValue)
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("Date", h.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Add("User-Agent", fmt.Sprintf("%s %s", h.appAgent, h.gofedAgent))
	req.Header.Set("Host", iri.Host)
	h.getSignerMu.Lock()
	err = h.getSigner.SignRequest(h.privKey, h.pubKeyId, req, nil)
	h.getSignerMu.Unlock()
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("GET request to %s failed (%d): %s", iri.String(), resp.StatusCode, resp.Status)
	}
	return resp, nil
}

// Deliver sends a POST request with an HTTP Signature.
func (h HttpSigTransport) Deliver(c context.Context, data map[string]interface{}, to *url.URL) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return h.deliver(c, b, to)
}

// BatchDeliver sends concurrent POST requests. Returns an error if any of the requests had an error.
func (h HttpSigTransport) BatchDeliver(c context.Context, data map[string]interface{}, recipients []*url.URL) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(recipients))
	for _, recipient := range recipients {
		wg.Add(1)
		go func(r *url.URL) {
			defer wg.Done()
			if err := h.deliver(c, b, r); err != nil {
				errCh <- err
			}
		}(recipient)
	}
	wg.Wait()
	errs := make([]string, 0, len(recipients))
outer:
	for {
		select {
		case e := <-errCh:
			errs = append(errs, e.Error())
		default:
			break outer
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("batch deliver had at least one failure: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (h HttpSigTransport) deliver(c context.Context, b []byte, to *url.URL) error {
	req, err := http.NewRequest("POST", to.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req = req.WithContext(c)
	req.Header.Add(contentTypeHeader, contentTypeHeaderValue)
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("Date", h.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Add("User-Agent", fmt.Sprintf("%s %s", h.appAgent, h.gofedAgent))
	req.Header.Set("Host", to.Host)
	h.postSignerMu.Lock()
	err = h.postSigner.SignRequest(h.privKey, h.pubKeyId, req, b)
	h.postSignerMu.Unlock()
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !isSuccess(resp.StatusCode) {
		return fmt.Errorf("POST request to %s failed (%d): %s", to.String(), resp.StatusCode, resp.Status)
	}
	return nil
}

// HttpClient sends http requests, and is an abstraction only needed by the
// HttpSigTransport. The standard library's Client satisfies this interface.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HttpClient must be implemented by http.Client.
var _ HttpClient = &http.Client{}
