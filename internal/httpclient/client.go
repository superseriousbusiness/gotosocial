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

package httpclient

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"runtime"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-cache/v2"
)

// ErrInvalidRequest is returned if a given HTTP request is invalid and cannot be performed.
var ErrInvalidRequest = errors.New("invalid http request")

// ErrReservedAddr is returned if a dialed address resolves to an IP within a blocked or reserved net.
var ErrReservedAddr = errors.New("dial within blocked / reserved IP range")

// ErrBodyTooLarge is returned when a received response body is above predefined limit (default 40MB).
var ErrBodyTooLarge = errors.New("body size too large")

// Config provides configuration details for setting up a new
// instance of httpclient.Client{}. Within are a subset of the
// configuration values passed to initialized http.Transport{}
// and http.Client{}, along with httpclient.Client{} specific.
type Config struct {
	// MaxOpenConnsPerHost limits the max number of open connections to a host.
	MaxOpenConnsPerHost int

	// MaxIdleConns: see http.Transport{}.MaxIdleConns.
	MaxIdleConns int

	// ReadBufferSize: see http.Transport{}.ReadBufferSize.
	ReadBufferSize int

	// WriteBufferSize: see http.Transport{}.WriteBufferSize.
	WriteBufferSize int

	// MaxBodySize determines the maximum fetchable body size.
	MaxBodySize int64

	// Timeout: see http.Client{}.Timeout.
	Timeout time.Duration

	// DisableCompression: see http.Transport{}.DisableCompression.
	DisableCompression bool

	// AllowRanges allows outgoing communications to given IP nets.
	AllowRanges []netip.Prefix

	// BlockRanges blocks outgoing communiciations to given IP nets.
	BlockRanges []netip.Prefix
}

// Client wraps an underlying http.Client{} to provide the following:
//   - setting a maximum received request body size, returning error on
//     large content lengths, and using a limited reader in all other
//     cases to protect against forged / unknown content-lengths
//   - protection from server side request forgery (SSRF) by only dialing
//     out to known public IP prefixes, configurable with allows/blocks
//   - limit number of concurrent requests, else blocking until a slot
//     is available (context channels still respected)
type Client struct {
	client http.Client
	queue  cache.Cache[string, chan struct{}]
	bmax   int64 // max response body size
	cmax   int   // max open conns per host
}

// New returns a new instance of Client initialized using configuration.
func New(cfg Config) *Client {
	var c Client

	d := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  &net.Resolver{},
	}

	if cfg.MaxOpenConnsPerHost <= 0 {
		// By default base this value on GOMAXPROCS.
		maxprocs := runtime.GOMAXPROCS(0)
		cfg.MaxOpenConnsPerHost = maxprocs * 20
	}

	if cfg.MaxIdleConns <= 0 {
		// By default base this value on MaxOpenConns
		cfg.MaxIdleConns = cfg.MaxOpenConnsPerHost * 10
	}

	if cfg.MaxBodySize <= 0 {
		// By default set this to a reasonable 40MB
		cfg.MaxBodySize = 40 * bytesize.MiB
	}

	// Protect dialer with IP range sanitizer
	d.Control = (&sanitizer{
		allow: cfg.AllowRanges,
		block: cfg.BlockRanges,
	}).Sanitize

	// Prepare client fields
	c.bmax = cfg.MaxBodySize
	c.queue = cache.New[string, chan struct{}]()
	c.client.Timeout = cfg.Timeout

	// Start cache sweep routines
	c.queue.SetTTL(time.Minute*10, true)
	_ = c.queue.Start(time.Second * 30)

	// Set underlying HTTP client roundtripper
	c.client.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		DialContext:           d.DialContext,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ReadBufferSize:        cfg.ReadBufferSize,
		WriteBufferSize:       cfg.WriteBufferSize,
		DisableCompression:    cfg.DisableCompression,
	}

	return &c
}

// Do will perform given request when an available slot in the queue is available,
// and block until this time. For returned values, this follows the same semantics
// as the standard http.Client{}.Do() implementation except that response body will
// be wrapped by an io.LimitReader() to limit response body sizes.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Get host's wait queue
	wait := c.wait(req.Host)

	// ... and wait our turn
	select {
	case <-req.Context().Done():
		// the request was canceled before we
		// got to our turn: no need to release
		return nil, req.Context().Err()
	case wait <- struct{}{}:
		// it's our turn!

		// NOTE:
		// Ideally here we would set the slot release to happen either
		// on error return, or via callback from the response body closer.
		// However when implementing this, there appear deadlocks between
		// the channel queue here and the media manager worker pool. So
		// currently we only place a limit on connections dialing out, but
		// there may still be more connections open than len(c.queue) given
		// that connections may not be closed until response body is closed.
		// The current implementation will reduce the viability of denial of
		// service attacks, but if there are future issues heed this advice :]
		defer func() { <-wait }()
	}

	// Firstly, ensure this is a valid request
	if err := ValidateRequest(req); err != nil {
		return nil, err
	}

	// Perform the HTTP request
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Check response body not too large
	if rsp.ContentLength > c.bmax {
		return nil, ErrBodyTooLarge
	}

	// Seperate the body implementers
	rbody := (io.Reader)(rsp.Body)
	cbody := (io.Closer)(rsp.Body)

	var limit int64

	if limit = rsp.ContentLength; limit < 0 {
		// If unknown, use max as reader limit
		limit = c.bmax
	}

	// Don't trust them, limit body reads
	rbody = io.LimitReader(rbody, limit)

	// Wrap body with limit
	rsp.Body = &struct {
		io.Reader
		io.Closer
	}{rbody, cbody}

	return rsp, nil
}

// wait acquires the 'wait' queue for the given host string, or allocates new.
func (c *Client) wait(host string) chan struct{} {
	// Look for an existing queue for host
	queue, ok := c.queue.Get(host)
	if ok {
		return queue
	}

	// Allocate a new host queue
	queue = make(chan struct{}, c.cmax)

	// Attempt to cache this queue by host
	if !c.queue.Put(host, queue) {
		// Someone beat us to the punch...
		queue, _ = c.queue.Get(host)
	}

	return queue
}
