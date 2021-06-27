package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (t *transport) DereferenceInstance(c context.Context, iri *url.URL) (*gtsmodel.Instance, error) {
	l := t.log.WithField("func", "DereferenceInstance")

	var i *gtsmodel.Instance
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

func dereferenceByAPIV1Instance(t *transport, c context.Context, iri *url.URL) (*gtsmodel.Instance, error) {
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

	if len(b) == 0 {
		return nil, errors.New("response bytes was len 0")
	}

	// try to parse the returned bytes directly into an Instance model
	apiResp := &apimodel.Instance{}
	if err := json.Unmarshal(b, apiResp); err != nil {
		return nil, err
	}

	var contactUsername string
	if apiResp.ContactAccount != nil {
		contactUsername = apiResp.ContactAccount.Username
	}

	ulid, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	i := &gtsmodel.Instance{
		ID:                     ulid,
		Domain:                 iri.Host,
		Title:                  apiResp.Title,
		URI:                    fmt.Sprintf("%s://%s", iri.Scheme, iri.Host),
		ShortDescription:       apiResp.ShortDescription,
		Description:            apiResp.Description,
		ContactEmail:           apiResp.Email,
		ContactAccountUsername: contactUsername,
		Version:                apiResp.Version,
	}

	return i, nil
}

func dereferenceByNodeInfo(t *transport, c context.Context, iri *url.URL) (*gtsmodel.Instance, error) {
	l := t.log.WithField("func", "dereferenceByNodeInfo")

	cleanIRI := &url.URL{
		Scheme: iri.Scheme,
		Host:   iri.Host,
		Path:   ".well-known/nodeinfo",
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

	if len(b) == 0 {
		return nil, errors.New("dereferenceByNodeInfo: response bytes was len 0")
	}

	wellKnownResp := &apimodel.WellKnownResponse{}
	if err := json.Unmarshal(b, wellKnownResp); err != nil {
		return nil, fmt.Errorf("dereferenceByNodeInfo: could not unmarshal server response as WellKnownResponse: %s", err)
	}

	// look through the links for the first one that matches the nodeinfo schema, this is what we need
	var nodeinfoHref *url.URL
	for _, l := range wellKnownResp.Links {
		if l.Href == "" || !strings.HasPrefix(l.Rel, "http://nodeinfo.diaspora.software/ns/schema") {
			continue
		}
		nodeinfoHref, err = url.Parse(l.Href)
		if err != nil {
			return nil, fmt.Errorf("dereferenceByNodeInfo: couldn't parse url %s: %s", l.Href, err)
		}
	}
	if nodeinfoHref == nil {
		return nil, errors.New("could not find nodeinfo rel in well known response")
	}
    // TODO: do the second query
	return nil, errors.New("not yet implemented")
}
