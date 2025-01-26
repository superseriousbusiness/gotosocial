// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package testrig

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

const (
	applicationJSON         = "application/json"
	applicationActivityJSON = "application/activity+json"
	textCSV                 = "text/csv"
	textPlain               = "text/plain"
)

// NewTestTransportController returns a test transport controller with the given http client.
//
// Obviously for testing purposes you should not be making actual http calls to other servers.
// To obviate this, use the function NewMockHTTPClient in this package to return a mock http
// client that doesn't make any remote calls but just returns whatever you tell it to.
//
// Unlike the other test interfaces provided in this package, you'll probably want to call this function
// PER TEST rather than per suite, so that the do function can be set on a test by test (or even more granular)
// basis.
func NewTestTransportController(state *state.State, client pub.HttpClient) transport.Controller {
	return transport.NewController(state, NewTestFederatingDB(state), &federation.Clock{}, client)
}

type MockHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)

	TestRemoteStatuses    map[string]vocab.ActivityStreamsNote
	TestRemotePeople      map[string]vocab.ActivityStreamsPerson
	TestRemoteGroups      map[string]vocab.ActivityStreamsGroup
	TestRemoteServices    map[string]vocab.ActivityStreamsService
	TestRemoteAttachments map[string]RemoteAttachmentFile
	TestRemoteEmojis      map[string]vocab.TootEmoji
	TestTombstones        map[string]*gtsmodel.Tombstone

	SentMessages sync.Map
}

// NewMockHTTPClient returns a client that conforms to the pub.HttpClient interface.
//
// If do is nil, then a standard response set will be mocked out, which includes models stored in the
// testrig, and webfinger responses as well.
//
// If do is not nil, then the given do function will always be used, which allows callers
// to customize how the client is mocked.
//
// Note that you should never ever make ACTUAL http calls with this thing.
func NewMockHTTPClient(do func(req *http.Request) (*http.Response, error), relativeMediaPath string, extraPeople ...ap.Accountable) *MockHTTPClient {
	mockHTTPClient := &MockHTTPClient{}

	if do != nil {
		mockHTTPClient.do = do
		return mockHTTPClient
	}

	mockHTTPClient.TestRemoteStatuses = NewTestFediStatuses()
	mockHTTPClient.TestRemotePeople = NewTestFediPeople()
	mockHTTPClient.TestRemoteGroups = NewTestFediGroups()
	mockHTTPClient.TestRemoteServices = NewTestFediServices()
	mockHTTPClient.TestRemoteAttachments = NewTestFediAttachments(relativeMediaPath)
	mockHTTPClient.TestRemoteEmojis = NewTestFediEmojis()
	mockHTTPClient.TestTombstones = NewTestTombstones()

	mockHTTPClient.do = func(req *http.Request) (*http.Response, error) {
		var (
			responseCode          = http.StatusNotFound
			responseBytes         = []byte(`{"error":"404 not found"}`)
			responseContentType   = applicationJSON
			responseContentLength = len(responseBytes)
			extraHeaders          = make(map[string]string, 0)
			reqURLString          = req.URL.String()
		)

		if req.Method == http.MethodPost {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}

			if sI, loaded := mockHTTPClient.SentMessages.LoadOrStore(reqURLString, [][]byte{b}); loaded {
				s, ok := sI.([][]byte)
				if !ok {
					panic("SentMessages entry wasn't [][]byte")
				}
				s = append(s, b)
				mockHTTPClient.SentMessages.Store(reqURLString, s)
			}

			responseCode = http.StatusOK
			responseBytes = []byte(`{"ok":"accepted"}`)
			responseContentType = applicationJSON
			responseContentLength = len(responseBytes)
		} else if strings.Contains(reqURLString, ".well-known/webfinger") {
			responseCode, responseBytes, responseContentType, responseContentLength, extraHeaders = WebfingerResponse(req)
		} else if strings.Contains(reqURLString, ".weird-webfinger-location/webfinger") {
			responseCode, responseBytes, responseContentType, responseContentLength, extraHeaders = WebfingerResponse(req)
		} else if strings.Contains(reqURLString, ".well-known/host-meta") {
			responseCode, responseBytes, responseContentType, responseContentLength, extraHeaders = HostMetaResponse(req)
		} else if strings.Contains(reqURLString, "lists.example.org") {
			responseCode, responseBytes, responseContentType, responseContentLength, extraHeaders = DomainPermissionSubscriptionResponse(req)
		} else if note, ok := mockHTTPClient.TestRemoteStatuses[reqURLString]; ok {
			// the request is for a note that we have stored
			noteI, err := streams.Serialize(note)
			if err != nil {
				panic(err)
			}
			noteJSON, err := json.Marshal(noteI)
			if err != nil {
				panic(err)
			}
			responseCode = http.StatusOK
			responseBytes = noteJSON
			responseContentType = applicationActivityJSON
			responseContentLength = len(noteJSON)
		} else if person, ok := mockHTTPClient.TestRemotePeople[reqURLString]; ok {
			// the request is for a person that we have stored
			personI, err := streams.Serialize(person)
			if err != nil {
				panic(err)
			}
			personJSON, err := json.Marshal(personI)
			if err != nil {
				panic(err)
			}
			responseCode = http.StatusOK
			responseBytes = personJSON
			responseContentType = applicationActivityJSON
			responseContentLength = len(personJSON)
		} else if group, ok := mockHTTPClient.TestRemoteGroups[reqURLString]; ok {
			// the request is for a person that we have stored
			groupI, err := streams.Serialize(group)
			if err != nil {
				panic(err)
			}
			groupJSON, err := json.Marshal(groupI)
			if err != nil {
				panic(err)
			}
			responseCode = http.StatusOK
			responseBytes = groupJSON
			responseContentType = applicationActivityJSON
			responseContentLength = len(groupJSON)
		} else if service, ok := mockHTTPClient.TestRemoteServices[reqURLString]; ok {
			serviceI, err := streams.Serialize(service)
			if err != nil {
				panic(err)
			}
			serviceJSON, err := json.Marshal(serviceI)
			if err != nil {
				panic(err)
			}
			responseCode = http.StatusOK
			responseBytes = serviceJSON
			responseContentType = applicationActivityJSON
			responseContentLength = len(serviceJSON)
		} else if emoji, ok := mockHTTPClient.TestRemoteEmojis[reqURLString]; ok {
			emojiI, err := streams.Serialize(emoji)
			if err != nil {
				panic(err)
			}
			emojiJSON, err := json.Marshal(emojiI)
			if err != nil {
				panic(err)
			}
			responseCode = http.StatusOK
			responseBytes = emojiJSON
			responseContentType = applicationActivityJSON
			responseContentLength = len(emojiJSON)
		} else if attachment, ok := mockHTTPClient.TestRemoteAttachments[reqURLString]; ok {
			responseCode = http.StatusOK
			responseBytes = attachment.Data
			responseContentType = attachment.ContentType
			responseContentLength = len(attachment.Data)
		} else if _, ok := mockHTTPClient.TestTombstones[reqURLString]; ok {
			responseCode = http.StatusGone
			responseBytes = []byte{}
			responseContentType = "text/html"
			responseContentLength = 0
		} else {
			for _, person := range extraPeople {
				// For any extra people, check if the
				// request matches one of:
				//
				//   - Public key URI
				//   - ActivityPub URI/id
				//   - Web URL.
				//
				// Since this is a test environment,
				// just assume all these values have
				// been properly set.
				if reqURLString == person.GetW3IDSecurityV1PublicKey().At(0).Get().GetJSONLDId().GetIRI().String() ||
					reqURLString == person.GetJSONLDId().GetIRI().String() ||
					reqURLString == person.GetActivityStreamsUrl().At(0).GetIRI().String() {
					personI, err := streams.Serialize(person)
					if err != nil {
						panic(err)
					}
					personJSON, err := json.Marshal(personI)
					if err != nil {
						panic(err)
					}
					responseCode = http.StatusOK
					responseBytes = personJSON
					responseContentType = applicationActivityJSON
					responseContentLength = len(personJSON)
				}
			}
		}

		log.Debugf(nil, "returning response %s", string(responseBytes))

		reader := bytes.NewReader(responseBytes)
		readCloser := io.NopCloser(reader)

		header := http.Header{
			"Content-Type": {responseContentType},
		}
		for k, v := range extraHeaders {
			header.Add(k, v)
		}

		return &http.Response{
			Request:       req,
			StatusCode:    responseCode,
			Body:          readCloser,
			ContentLength: int64(responseContentLength),
			Header:        header,
		}, nil
	}

	return mockHTTPClient
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

func (m *MockHTTPClient) DoSigned(req *http.Request, sign httpclient.SignFunc) (*http.Response, error) {
	return m.do(req)
}

func HostMetaResponse(req *http.Request) (
	responseCode int,
	responseBytes []byte,
	responseContentType string,
	responseContentLength int,
	extraHeaders map[string]string,
) {
	var hm *apimodel.HostMeta

	if req.URL.String() == "https://misconfigured-instance.com/.well-known/host-meta" {
		hm = &apimodel.HostMeta{
			XMLNS: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
			Link: []apimodel.Link{
				{
					Rel:      "lrdd",
					Type:     "application/xrd+xml",
					Template: "https://misconfigured-instance.com/.weird-webfinger-location/webfinger?resource={uri}",
				},
			},
		}
	}

	if hm == nil {
		log.Debugf(nil, "hostmeta response not available for %s", req.URL)
		responseCode = http.StatusNotFound
		responseBytes = []byte(``)
		responseContentType = "application/xml"
		responseContentLength = len(responseBytes)
		return
	}

	hmXML, err := xml.Marshal(hm)
	if err != nil {
		panic(err)
	}
	responseCode = http.StatusOK
	responseBytes = hmXML
	responseContentType = "application/xml"
	responseContentLength = len(hmXML)
	return
}

func WebfingerResponse(req *http.Request) (
	responseCode int,
	responseBytes []byte,
	responseContentType string,
	responseContentLength int,
	extraHeaders map[string]string,
) {
	var wfr *apimodel.WellKnownResponse

	switch req.URL.String() {
	case "https://unknown-instance.com/.well-known/webfinger?resource=acct%3Asome_group%40unknown-instance.com":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:some_group@unknown-instance.com",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://unknown-instance.com/groups/some_group",
				},
			},
		}
	case "https://owncast.example.org/.well-known/webfinger?resource=acct%3Argh%40owncast.example.org":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:rgh@example.org",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://owncast.example.org/federation/user/rgh",
				},
			},
		}
	case "https://unknown-instance.com/.well-known/webfinger?resource=acct%3Abrand_new_person%40unknown-instance.com":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:brand_new_person@unknown-instance.com",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://unknown-instance.com/users/brand_new_person",
				},
			},
		}
	case "https://xn--pnycde-zxa8b.example.org/.well-known/webfinger?resource=acct%3Abrand_new_person%40xn--pnycde-zxa8b.example.org":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:brand_new_person@unknown-instance.com",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://unknown-instance.com/users/brand_new_person",
				},
			},
		}
	case "https://turnip.farm/.well-known/webfinger?resource=acct%3Aturniplover6969%40turnip.farm":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:turniplover6969@turnip.farm",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://turnip.farm/users/turniplover6969",
				},
			},
		}
	case "https://fossbros-anonymous.io/.well-known/webfinger?resource=acct%3Afoss_satan%40fossbros-anonymous.io":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:foss_satan@fossbros-anonymous.io",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "http://fossbros-anonymous.io/users/foss_satan",
				},
			},
		}
	case "https://example.org/.well-known/webfinger?resource=acct%3ASome_User%40example.org":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:Some_User@example.org",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://example.org/users/Some_User",
				},
			},
		}
	case "https://misconfigured-instance.com/.weird-webfinger-location/webfinger?resource=acct%3Asomeone%40misconfigured-instance.com":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:someone@misconfigured-instance.com",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://misconfigured-instance.com/users/someone",
				},
			},
		}
	}

	if wfr == nil {
		log.Debugf(nil, "webfinger response not available for %s", req.URL)
		responseCode = http.StatusNotFound
		responseBytes = []byte(`{"error":"not found"}`)
		responseContentType = applicationJSON
		responseContentLength = len(responseBytes)
		return
	}

	wfrJSON, err := json.Marshal(wfr)
	if err != nil {
		panic(err)
	}
	responseCode = http.StatusOK
	responseBytes = wfrJSON
	responseContentType = applicationJSON
	responseContentLength = len(wfrJSON)
	return
}

func DomainPermissionSubscriptionResponse(req *http.Request) (
	responseCode int,
	responseBytes []byte,
	responseContentType string,
	responseContentLength int,
	extraHeaders map[string]string,
) {

	const (
		lastModified        = "Sat, 21 Sep 2024 22:00:00 GMT"
		futureLastModified  = "Mon, 15 Jan 2300 22:00:00 GMT"
		garbageLastModified = "I LIKE BIG BUTTS AND I CANNOT LIE"

		csvResp = `#domain,#severity,#reject_media,#reject_reports,#public_comment,#obfuscate
bumfaces.net,suspend,false,false,big jerks,false
peepee.poopoo,suspend,false,false,harassment,false
nothanks.com,suspend,false,false,,false`
		csvRespETag = "\"bigbums6969\""

		textResp = `bumfaces.net
peepee.poopoo
nothanks.com`
		textRespETag = "\"this is a legit etag i swear\""
		jsonResp     = `[
  {
    "domain": "bumfaces.net",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "big jerks"
  },
    {
    "domain": "peepee.poopoo",
    "suspended_at": "2020-05-13T13:29:12.000Z",
    "public_comment": "harassment"
  },
    {
    "domain": "nothanks.com",
    "suspended_at": "2020-05-13T13:29:12.000Z"
  }
]`
		jsonRespETag = "\"don't modify me daddy\""
	)

	switch req.URL.String() {
	case "https://lists.example.org/baddies.csv":
		extraHeaders = map[string]string{
			"Last-Modified": lastModified,
			"ETag":          csvRespETag,
		}
		if req.Header.Get("If-None-Match") == csvRespETag {
			// Cached.
			responseCode = http.StatusNotModified
		} else {
			responseBytes = []byte(csvResp)
			responseContentType = textCSV
			responseCode = http.StatusOK
		}
		responseContentLength = len(responseBytes)

	case "https://lists.example.org/baddies.txt":
		extraHeaders = map[string]string{
			"Last-Modified": lastModified,
			"ETag":          textRespETag,
		}
		if req.Header.Get("If-None-Match") == textRespETag {
			// Cached.
			responseCode = http.StatusNotModified
		} else {
			responseBytes = []byte(textResp)
			responseContentType = textPlain
			responseCode = http.StatusOK
		}
		responseContentLength = len(responseBytes)

	case "https://lists.example.org/baddies.json":
		extraHeaders = map[string]string{
			"Last-Modified": lastModified,
			"ETag":          jsonRespETag,
		}
		if req.Header.Get("If-None-Match") == jsonRespETag {
			// Cached.
			responseCode = http.StatusNotModified
		} else {
			responseBytes = []byte(jsonResp)
			responseContentType = applicationJSON
			responseCode = http.StatusOK
		}
		responseContentLength = len(responseBytes)

	case "https://lists.example.org/baddies.csv?future=true":
		extraHeaders = map[string]string{
			// Provide the future last modified value.
			"Last-Modified": futureLastModified,
			"ETag":          csvRespETag,
		}
		if req.Header.Get("If-None-Match") == csvRespETag {
			// Cached.
			responseCode = http.StatusNotModified
		} else {
			responseBytes = []byte(csvResp)
			responseContentType = textCSV
			responseCode = http.StatusOK
		}
		responseContentLength = len(responseBytes)

	case "https://lists.example.org/baddies.csv?garbage=true":
		extraHeaders = map[string]string{
			// Provide the garbage last modified value.
			"Last-Modified": garbageLastModified,
			"ETag":          csvRespETag,
		}
		if req.Header.Get("If-None-Match") == csvRespETag {
			// Cached.
			responseCode = http.StatusNotModified
		} else {
			responseBytes = []byte(csvResp)
			responseContentType = textCSV
			responseCode = http.StatusOK
		}
		responseContentLength = len(responseBytes)

	default:
		responseCode = http.StatusNotFound
		responseBytes = []byte(`{"error":"not found"}`)
		responseContentType = applicationJSON
		responseContentLength = len(responseBytes)
	}

	return
}
