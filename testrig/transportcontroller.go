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
	"io/ioutil"
	"net/http"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
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

// NewMockHTTPClient returns a client that conforms to the pub.HttpClient interface,
// but will always just execute the given `do` function, allowing responses to be mocked.
//
// If 'do' is nil, then a no-op function will be used instead, that just returns status 200.
//
// Note that you should never ever make ACTUAL http calls with this thing.
func NewMockHTTPClient(do func(req *http.Request) (*http.Response, error)) pub.HttpClient {
	if do == nil {
		return &mockHTTPClient{
			do: func(req *http.Request) (*http.Response, error) {
				r := ioutil.NopCloser(bytes.NewReader([]byte{}))
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			},
		}
	}
	return &mockHTTPClient{
		do: do,
	}
}

type mockHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}
