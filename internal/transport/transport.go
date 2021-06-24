package transport

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

// Transport wraps the pub.Transport interface with some additional
// functionality for fetching remote media.
type Transport interface {
	pub.Transport
	// DereferenceMedia fetches the bytes of the given media attachment IRI, with the expectedContentType.
	DereferenceMedia(c context.Context, iri *url.URL, expectedContentType string) ([]byte, error)
	// DereferenceInstance dereferences remote instance information, first by checking /api/v1/instance, and then by checking /.well-known/nodeinfo.
	DereferenceInstance(c context.Context, iri *url.URL) (*apimodel.Instance, error)
	// Finger performs a webfinger request with the given username and domain, and returns the bytes from the response body.
	Finger(c context.Context, targetUsername string, targetDomains string) ([]byte, error)
}

// transport implements the Transport interface
type transport struct {
	client       pub.HttpClient
	appAgent     string
	gofedAgent   string
	clock        pub.Clock
	pubKeyID     string
	privkey      crypto.PrivateKey
	sigTransport *pub.HttpSigTransport
	getSigner    httpsig.Signer
	getSignerMu  *sync.Mutex
	log          *logrus.Logger
}

func (t *transport) BatchDeliver(c context.Context, b []byte, recipients []*url.URL) error {
	return t.sigTransport.BatchDeliver(c, b, recipients)
}

func (t *transport) Deliver(c context.Context, b []byte, to *url.URL) error {
	l := t.log.WithField("func", "Deliver")
	l.Debugf("performing POST to %s", to.String())
	return t.sigTransport.Deliver(c, b, to)
}

func (t *transport) Dereference(c context.Context, iri *url.URL) ([]byte, error) {
	l := t.log.WithField("func", "Dereference")
	l.Debugf("performing GET to %s", iri.String())
	return t.sigTransport.Dereference(c, iri)
}

func (t *transport) DereferenceMedia(c context.Context, iri *url.URL, expectedContentType string) ([]byte, error) {
	l := t.log.WithField("func", "DereferenceMedia")
	l.Debugf("performing GET to %s", iri.String())
	req, err := http.NewRequest("GET", iri.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(c)
	if expectedContentType == "" {
		req.Header.Add("Accept", "*/*")
	} else {
		req.Header.Add("Accept", expectedContentType)
	}
	req.Header.Add("Date", t.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Add("User-Agent", fmt.Sprintf("%s %s", t.appAgent, t.gofedAgent))
	req.Header.Set("Host", iri.Host)
	t.getSignerMu.Lock()
	err = t.getSigner.SignRequest(t.privkey, t.pubKeyID, req, nil)
	t.getSignerMu.Unlock()
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request to %s failed (%d): %s", iri.String(), resp.StatusCode, resp.Status)
	}
	return ioutil.ReadAll(resp.Body)
}

func (t *transport) Finger(c context.Context, targetUsername string, targetDomain string) ([]byte, error) {
	l := t.log.WithField("func", "Finger")
	urlString := fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s@%s", targetDomain, targetUsername, targetDomain)
	l.Debugf("performing GET to %s", urlString)

	iri, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("Finger: error parsing url %s: %s", urlString, err)
	}

	l.Debugf("performing GET to %s", iri.String())

	req, err := http.NewRequest("GET", iri.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(c)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept", "application/jrd+json")
	req.Header.Add("Date", t.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Add("User-Agent", fmt.Sprintf("%s %s", t.appAgent, t.gofedAgent))
	req.Header.Set("Host", iri.Host)
	t.getSignerMu.Lock()
	err = t.getSigner.SignRequest(t.privkey, t.pubKeyID, req, nil)
	t.getSignerMu.Unlock()
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request to %s failed (%d): %s", iri.String(), resp.StatusCode, resp.Status)
	}
	return ioutil.ReadAll(resp.Body)
}

func (t *transport) DereferenceInstance(c context.Context, iri *url.URL) (*apimodel.Instance, error) {
	l := t.log.WithField("func", "DereferenceInstance")

	var i *apimodel.Instance
	var err error

	// First try to dereference using /api/v1/instance.
	// This will provide the most complete picture of an instance, and avoid unnecessary api calls.
	//
	// This will only work with Mastodon-api compatible instances: Mastodon, some Pleroma instances, GoToSocial.
	l.Debugf("trying to dereference instance %s by /api/v1/instance", iri.Host)
	i, err = dereferenceByAPIV1Instance(t, c, iri)
	if err == nil {
		l.Debugf("successfully dereferenced instance using /api/v1/instance")
		return i, nil
	}
	l.Debugf("couldn't dereference instance using /api/v1/instance: %s", err)

	// If that doesn't work, try to dereference using /.well-known/nodeinfo.
	// This will involve two API calls and return less info overall, but should be more widely compatible.
	l.Debugf("trying to dereference instance %s by /.well-known/nodeinfo", iri.Host)
	i, err = dereferenceByNodeInfo(t, c, iri)
	if err == nil {
		l.Debugf("successfully dereferenced instance using /.well-known/nodeinfo")
		return i, nil
	}
	l.Debugf("couldn't dereference instance using /.well-known/nodeinfo: %s", err)

	return nil, fmt.Errorf("couldn't dereference instance %s using either /api/v1/instance or /.well-known/nodeinfo", iri.Host)
}

func dereferenceByAPIV1Instance(t *transport, c context.Context, iri *url.URL) (*apimodel.Instance, error) {
	l := t.log.WithField("func", "dereferenceByAPIV1Instance")

	cleanIRI := &url.URL{
		Scheme: iri.Scheme,
		Host:   iri.Host,
		Path:   "api/v1/instance",
	}

	l.Debugf("performing GET to %s", cleanIRI.String())
	req, err := http.NewRequest("GET", cleanIRI.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(c)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Date", t.clock.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Add("User-Agent", fmt.Sprintf("%s %s", t.appAgent, t.gofedAgent))
	req.Header.Set("Host", cleanIRI.Host)
	t.getSignerMu.Lock()
	err = t.getSigner.SignRequest(t.privkey, t.pubKeyID, req, nil)
	t.getSignerMu.Unlock()
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request to %s failed (%d): %s", cleanIRI.String(), resp.StatusCode, resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// try to parse the returned bytes directly into an Instance model
	i := &apimodel.Instance{}
	if err := json.Unmarshal(b, i); err != nil {
		return nil, err
	}

	return i, nil
}

func dereferenceByNodeInfo(t *transport, c context.Context, iri *url.URL) (*apimodel.Instance, error) {
	return nil, errors.New("not yet implemented")
}
