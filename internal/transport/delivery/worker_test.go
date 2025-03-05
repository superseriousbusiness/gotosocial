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

package delivery_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"testing"

	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/queue"
	"github.com/superseriousbusiness/gotosocial/internal/transport/delivery"
)

func TestDeliveryWorkerPool(t *testing.T) {
	for _, i := range []int{1, 2, 4, 8, 16, 32} {
		t.Run("size="+strconv.Itoa(i), func(t *testing.T) {
			testDeliveryWorkerPool(t, i, generateInput(100*i))
		})
	}
}

func testDeliveryWorkerPool(t *testing.T, sz int, input []*testrequest) {
	wp := new(delivery.WorkerPool)
	allowLocal := []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}
	wp.Init(httpclient.New(httpclient.Config{AllowRanges: allowLocal}))
	wp.Start(sz)
	defer wp.Stop()
	test(t, &wp.Queue, input)
}

func test(
	t *testing.T,
	queue *queue.StructQueue[*delivery.Delivery],
	input []*testrequest,
) {
	expect := make(chan *testrequest)
	errors := make(chan error)

	// Prepare an HTTP test handler that ensures expected delivery is received.
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		errors <- (<-expect).Equal(r)
	})

	// Start new HTTP test server listener.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Start the HTTP server.
	//
	// specifically not using httptest.Server{} here as httptest
	// links that server with its own http.Client{}, whereas we're
	// using an httpclient.Client{} (well, delivery routine is).
	srv := new(http.Server)
	srv.Addr = "http://" + l.Addr().String()
	srv.Handler = handler
	go srv.Serve(l)
	defer srv.Close()

	// Range over test input.
	for _, test := range input {

		// Generate req for input.
		req := test.Generate(srv.Addr)
		r := httpclient.WrapRequest(req)

		// Wrap the request in delivery.
		dlv := new(delivery.Delivery)
		dlv.Request = r

		// Enqueue delivery!
		queue.Push(dlv)
		expect <- test

		// Wait for errors from handler.
		if err := <-errors; err != nil {
			t.Error(err)
		}
	}
}

type testrequest struct {
	method string
	uri    string
	body   []byte
}

// generateInput generates 'n' many testrequest cases.
func generateInput(n int) []*testrequest {
	tests := make([]*testrequest, n)
	for i := range tests {
		tests[i] = new(testrequest)
		tests[i].method = randomMethod()
		tests[i].uri = randomURI()
		tests[i].body = randomBody(tests[i].method)
	}
	return tests
}

var methods = []string{
	http.MethodConnect,
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
	http.MethodTrace,
}

// randomMethod generates a random http method.
func randomMethod() string {
	return methods[rand.Intn(len(methods))]
}

// randomURI generates a random http uri.
func randomURI() string {
	n := rand.Intn(5)
	p := make([]string, n)
	for i := range p {
		p[i] = strconv.Itoa(rand.Int())
	}
	return "/" + strings.Join(p, "/")
}

// randomBody generates a random http body DEPENDING on method.
func randomBody(method string) []byte {
	if requiresBody(method) {
		return []byte(method + " " + randomURI())
	}
	return nil
}

// requiresBody returns whether method requires body.
func requiresBody(method string) bool {
	switch method {
	case http.MethodPatch,
		http.MethodPost,
		http.MethodPut:
		return true
	default:
		return false
	}
}

// Generate will generate a real http.Request{} from test data.
func (t *testrequest) Generate(addr string) *http.Request {
	var body io.ReadCloser
	if t.body != nil {
		var b bytes.Reader
		b.Reset(t.body)
		body = io.NopCloser(&b)
	}
	req, err := http.NewRequest(t.method, addr+t.uri, body)
	if err != nil {
		panic(err)
	}
	return req
}

// Equal checks if request matches receiving test request.
func (t *testrequest) Equal(r *http.Request) error {
	// Ensure methods match.
	if t.method != r.Method {
		return fmt.Errorf("differing request methods: t=%q r=%q", t.method, r.Method)
	}

	// Ensure request URIs match.
	if t.uri != r.URL.RequestURI() {
		return fmt.Errorf("differing request urls: t=%q r=%q", t.uri, r.URL.RequestURI())
	}

	// Ensure body cases match.
	if requiresBody(t.method) {

		// Read request into memory.
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading request body: %v", err)
		}

		// Compare the request bodies.
		st := strings.TrimSpace(string(t.body))
		sr := strings.TrimSpace(string(b))
		if st != sr {
			return fmt.Errorf("differing request bodies: t=%q r=%q", st, sr)
		}
	}

	return nil
}
