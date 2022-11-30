/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package testrig

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

const (
	applicationJSON         = "application/json"
	applicationActivityJSON = "application/activity+json"
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
func NewTestTransportController(client pub.HttpClient, db db.DB, fedWorker *concurrency.WorkerPool[messages.FromFederator]) transport.Controller {
	return transport.NewController(db, NewTestFederatingDB(db, fedWorker), &federation.Clock{}, client)
}

type MockHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)

	testRemoteStatuses    map[string]vocab.ActivityStreamsNote
	testRemotePeople      map[string]vocab.ActivityStreamsPerson
	testRemoteGroups      map[string]vocab.ActivityStreamsGroup
	testRemoteServices    map[string]vocab.ActivityStreamsService
	testRemoteAttachments map[string]RemoteAttachmentFile
	testRemoteEmojis      map[string]vocab.TootEmoji
	testTombstones        map[string]*gtsmodel.Tombstone

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
func NewMockHTTPClient(do func(req *http.Request) (*http.Response, error), relativeMediaPath string) *MockHTTPClient {
	mockHTTPClient := &MockHTTPClient{}

	if do != nil {
		mockHTTPClient.do = do
		return mockHTTPClient
	}

	mockHTTPClient.testRemoteStatuses = NewTestFediStatuses()
	mockHTTPClient.testRemotePeople = NewTestFediPeople()
	mockHTTPClient.testRemoteGroups = NewTestFediGroups()
	mockHTTPClient.testRemoteServices = NewTestFediServices()
	mockHTTPClient.testRemoteAttachments = NewTestFediAttachments(relativeMediaPath)
	mockHTTPClient.testRemoteEmojis = NewTestFediEmojis()
	mockHTTPClient.testTombstones = NewTestTombstones()

	mockHTTPClient.do = func(req *http.Request) (*http.Response, error) {
		responseCode := http.StatusNotFound
		responseBytes := []byte(`{"error":"404 not found"}`)
		responseContentType := applicationJSON
		responseContentLength := len(responseBytes)

		if req.Method == http.MethodPost {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}

			if sI, loaded := mockHTTPClient.SentMessages.LoadOrStore(req.URL.String(), [][]byte{b}); loaded {
				s, ok := sI.([][]byte)
				if !ok {
					panic("SentMessages entry wasn't [][]byte")
				}
				s = append(s, b)
				mockHTTPClient.SentMessages.Store(req.URL.String(), s)
			}

			responseCode = http.StatusOK
			responseBytes = []byte(`{"ok":"accepted"}`)
			responseContentType = applicationJSON
			responseContentLength = len(responseBytes)
		} else if strings.Contains(req.URL.String(), ".well-known/webfinger") {
			responseCode, responseBytes, responseContentType, responseContentLength = WebfingerResponse(req)
		} else if note, ok := mockHTTPClient.testRemoteStatuses[req.URL.String()]; ok {
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
		} else if person, ok := mockHTTPClient.testRemotePeople[req.URL.String()]; ok {
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
		} else if group, ok := mockHTTPClient.testRemoteGroups[req.URL.String()]; ok {
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
		} else if service, ok := mockHTTPClient.testRemoteServices[req.URL.String()]; ok {
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
		} else if emoji, ok := mockHTTPClient.testRemoteEmojis[req.URL.String()]; ok {
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
		} else if attachment, ok := mockHTTPClient.testRemoteAttachments[req.URL.String()]; ok {
			responseCode = http.StatusOK
			responseBytes = attachment.Data
			responseContentType = attachment.ContentType
			responseContentLength = len(attachment.Data)
		} else if _, ok := mockHTTPClient.testTombstones[req.URL.String()]; ok {
			responseCode = http.StatusGone
			responseBytes = []byte{}
			responseContentType = "text/html"
			responseContentLength = 0
		}

		log.Debugf("returning response %s", string(responseBytes))
		reader := bytes.NewReader(responseBytes)
		readCloser := io.NopCloser(reader)
		return &http.Response{
			StatusCode:    responseCode,
			Body:          readCloser,
			ContentLength: int64(responseContentLength),
			Header: http.Header{
				"content-type": {responseContentType},
			},
		}, nil
	}

	return mockHTTPClient
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

func WebfingerResponse(req *http.Request) (responseCode int, responseBytes []byte, responseContentType string, responseContentLength int) {
	var wfr *apimodel.WellKnownResponse

	switch req.URL.String() {
	case "https://unknown-instance.com/.well-known/webfinger?resource=acct:some_group@unknown-instance.com":
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
	case "https://owncast.example.org/.well-known/webfinger?resource=acct:rgh@owncast.example.org":
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
	case "https://unknown-instance.com/.well-known/webfinger?resource=acct:brand_new_person@unknown-instance.com":
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
	case "https://turnip.farm/.well-known/webfinger?resource=acct:turniplover6969@turnip.farm":
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
	case "https://fossbros-anonymous.io/.well-known/webfinger?resource=acct:foss_satan@fossbros-anonymous.io":
		wfr = &apimodel.WellKnownResponse{
			Subject: "acct:foss_satan@fossbros-anonymous.io",
			Links: []apimodel.Link{
				{
					Rel:  "self",
					Type: applicationActivityJSON,
					Href: "https://fossbros-anonymous.io/users/foss_satan",
				},
			},
		}
	case "https://example.org/.well-known/webfinger?resource=acct:Some_User@example.org":
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
	}

	if wfr == nil {
		log.Debugf("webfinger response not available for %s", req.URL)
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
