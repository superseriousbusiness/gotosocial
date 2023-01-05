/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

// Note:
// There is no test for the .MaxOpenConns implementation
// in the httpclient.Client{}, due to the difficult to test
// this. The block is only held for the actual dial out to
// the connection, so the usual test of blocking and holding
// open this queue slot to check we can't open another isn't
// an easy test here.

func TestHTTPClientSmallBody(t *testing.T) {
	for _, body := range bodies {
		_TestHTTPClientWithBody(t, []byte(body), int(^uint16(0)))
	}
}

func TestHTTPClientExactBody(t *testing.T) {
	for _, body := range bodies {
		_TestHTTPClientWithBody(t, []byte(body), len(body))
	}
}

func TestHTTPClientLargeBody(t *testing.T) {
	for _, body := range bodies {
		_TestHTTPClientWithBody(t, []byte(body), len(body)-1)
	}
}

func _TestHTTPClientWithBody(t *testing.T, body []byte, max int) {
	var (
		handler http.HandlerFunc

		expect []byte

		expectErr error
	)

	// If this is a larger body, reslice and
	// set error so we know what to expect
	expect = body
	if max < len(body) {
		expect = expect[:max]
		expectErr = httpclient.ErrBodyTooLarge
	}

	// Create new HTTP client with maximum body size
	client := httpclient.New(httpclient.Config{
		MaxBodySize:        int64(max),
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
	if !errors.Is(err, expectErr) {
		t.Fatalf("error performing client request: %v", err)
	} else if err != nil {
		return // expected error
	}
	defer rsp.Body.Close()

	// Read response body into memory
	check, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}

	// Check actual response body matches expected
	if !bytes.Equal(expect, check) {
		t.Errorf("response body did not match expected: expect=%q actual=%q", string(expect), string(check))
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
