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

package httpclient_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
)

var privateIPs = []string{
	"http://127.0.0.1:80",
	"http://0.0.0.0:80",
	"http://192.168.0.1:80",
	"http://192.168.1.0:80",
	"http://10.0.0.0:80",
	"http://172.16.0.0:80",
	"http://10.255.255.255:80",
	"http://172.31.255.255:80",
	"http://255.255.255.255:80",
}

var bodies = []string{
	"hello world!",
	"{}",
	`{"key": "value", "some": "kinda bullshit"}`,
	"body with\r\nnewlines",
}

func TestHTTPClientBody(t *testing.T) {
	for _, body := range bodies {
		testHTTPClientWithBody(t, []byte(body))
	}
}

func testHTTPClientWithBody(t *testing.T, body []byte) {
	var (
		handler http.HandlerFunc
	)

	// Create new HTTP client with maximum body size
	client := httpclient.New(httpclient.Config{
		DisableCompression: true,
		AllowRanges: []netip.Prefix{
			// Loopback (used by server)
			netip.MustParsePrefix("127.0.0.1/8"),
		},
	})

	// Set simple body-writing test handler
	handler = func(rw http.ResponseWriter, r *http.Request) {
		_, _ = rw.Write(body)
	}

	// Start the test server
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Wrap body to provide reader iface
	rbody := bytes.NewReader(body)

	// Create the test HTTP request
	req, _ := http.NewRequest("POST", srv.URL, rbody)

	// Perform the test request
	rsp, err := client.Do(req)
	if err != nil {
		t.Fatalf("error performing client request: %v", err)
	}
	defer rsp.Body.Close()

	// Read response body into memory
	check, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}

	// Check actual response body matches expected
	if !bytes.Equal(body, check) {
		t.Errorf("response body did not match expected: expect=%q actual=%q", string(body), string(check))
	}
}

func TestHTTPClientPrivateIP(t *testing.T) {
	client := httpclient.New(httpclient.Config{})

	for _, addr := range privateIPs {
		// Prepare request to private IP
		req, _ := http.NewRequest("GET", addr, nil)

		// Perform the HTTP request
		_, err := client.Do(req)
		if !errors.Is(err, httpclient.ErrReservedAddr) {
			t.Errorf("dialing private address did not return expected error: %v", err)
		}
	}
}
