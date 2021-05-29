package transport

import (
	"context"
	"crypto"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
)

// Transport wraps the pub.Transport interface with some additional
// functionality for fetching remote media.
type Transport interface {
	pub.Transport
	DereferenceMedia(c context.Context, iri *url.URL, expectedContentType string) ([]byte, error)
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
