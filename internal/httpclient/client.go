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

package httpclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"runtime"
	"strconv"
	"strings"
	"time"

	"codeberg.org/gruf/go-cache/v3"
	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-iotools"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

var (
	// ErrInvalidRequest is returned if a given HTTP request is invalid and cannot be performed.
	ErrInvalidRequest = errors.New("invalid http request")

	// ErrInvalidNetwork is returned if the request would not be performed over TCP
	ErrInvalidNetwork = errors.New("invalid network type")

	// ErrReservedAddr is returned if a dialed address resolves to an IP within a blocked or reserved net.
	ErrReservedAddr = errors.New("dial within blocked / reserved IP range")
)

// Config provides configuration details for setting up a new
// instance of httpclient.Client{}. Within are a subset of the
// configuration values passed to initialized http.Transport{}
// and http.Client{}, along with httpclient.Client{} specific.
type Config struct {

	// MaxOpenConnsPerHost limits the max
	// number of open connections to a host.
	MaxOpenConnsPerHost int

	// AllowRanges allows outgoing
	// communications to given IP nets.
	AllowRanges []netip.Prefix

	// BlockRanges blocks outgoing
	// communiciations to given IP nets.
	BlockRanges []netip.Prefix

	// TLSInsecureSkipVerify can be set to true to
	// skip validation of remote TLS certificates.
	//
	// THIS SHOULD BE USED FOR TESTING ONLY, IF YOU
	// TURN THIS ON WHILE RUNNING IN PRODUCTION YOU
	// ARE LEAVING YOUR SERVER WIDE OPEN TO ATTACKS!
	TLSInsecureSkipVerify bool

	// MaxIdleConns: see http.Transport{}.MaxIdleConns.
	MaxIdleConns int

	// ReadBufferSize: see http.Transport{}.ReadBufferSize.
	ReadBufferSize int

	// WriteBufferSize: see http.Transport{}.WriteBufferSize.
	WriteBufferSize int

	// Timeout: see http.Client{}.Timeout.
	Timeout time.Duration

	// DisableCompression: see http.Transport{}.DisableCompression.
	DisableCompression bool
}

// Client wraps an underlying http.Client{} to provide the following:
//   - setting a maximum received request body size, returning error on
//     large content lengths, and using a limited reader in all other
//     cases to protect against forged / unknown content-lengths
//   - protection from server side request forgery (SSRF) by only dialing
//     out to known public IP prefixes, configurable with allows/blocks
//   - retry-backoff logic for error temporary HTTP error responses
//   - optional request signing
//   - request logging
type Client struct {
	client   http.Client
	badHosts cache.TTLCache[string, struct{}]
	retries  uint
}

// New returns a new instance of Client initialized using configuration.
func New(cfg Config) *Client {
	var c Client
	c.retries = 5

	d := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  &net.Resolver{},
	}

	if cfg.MaxOpenConnsPerHost <= 0 {
		// By default base this value on GOMAXPROCS.
		maxprocs := runtime.GOMAXPROCS(0)
		cfg.MaxOpenConnsPerHost = maxprocs * 20
	}

	if cfg.MaxIdleConns <= 0 {
		// By default base this value on MaxOpenConns.
		cfg.MaxIdleConns = cfg.MaxOpenConnsPerHost * 10
	}

	// Protect the dialer
	// with IP range sanitizer.
	d.Control = (&Sanitizer{
		Allow: cfg.AllowRanges,
		Block: cfg.BlockRanges,
	}).Sanitize

	// Prepare client fields.
	c.client.Timeout = cfg.Timeout

	// Prepare transport TLS config.
	tlsClientConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLSInsecureSkipVerify, //nolint:gosec
	}

	if tlsClientConfig.InsecureSkipVerify {
		// Warn against playing silly buggers.
		log.Warn(nil, "http-client.tls-insecure-skip-verify was set to TRUE. "+
			"*****THIS SHOULD BE USED FOR TESTING ONLY, IF YOU TURN THIS ON WHILE "+
			"RUNNING IN PRODUCTION YOU ARE LEAVING YOUR SERVER WIDE OPEN TO ATTACKS! "+
			"IF IN DOUBT, STOP YOUR SERVER *NOW* AND ADJUST YOUR CONFIGURATION!*****",
		)
	}

	// Set underlying HTTP client roundtripper.
	c.client.Transport = &signingtransport{http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		DialContext:           d.DialContext,
		TLSClientConfig:       tlsClientConfig,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ReadBufferSize:        cfg.ReadBufferSize,
		WriteBufferSize:       cfg.WriteBufferSize,
		DisableCompression:    cfg.DisableCompression,
	}}

	// Initiate outgoing bad hosts lookup cache.
	c.badHosts = cache.NewTTL[string, struct{}](0, 512, 0)
	c.badHosts.SetTTL(time.Hour, false)
	if !c.badHosts.Start(time.Minute) {
		log.Panic(nil, "failed to start transport controller cache")
	}

	return &c
}

// Do will essentially perform http.Client{}.Do() with retry-backoff functionality.
func (c *Client) Do(r *http.Request) (rsp *http.Response, err error) {

	// First validate incoming request.
	if err := ValidateRequest(r); err != nil {
		return nil, err
	}

	// Wrap in our own request
	// type for retry-backoff.
	req := WrapRequest(r)

	if gtscontext.IsFastfail(r.Context()) {
		// If the fast-fail flag was set, just
		// attempt a single iteration instead of
		// following the below retry-backoff loop.
		rsp, _, err = c.DoOnce(req)
		if err != nil {
			return nil, fmt.Errorf("%w (fast fail)", err)
		}
		return rsp, nil
	}

	for {
		var retry bool

		// Perform the http request.
		rsp, retry, err = c.DoOnce(req)
		if err == nil {
			return rsp, nil
		}

		if !retry {
			// reached max retries, don't further backoff
			return nil, fmt.Errorf("%w (max retries)", err)
		}

		// Start new backoff sleep timer.
		backoff := time.NewTimer(req.BackOff())

		select {
		// Request ctx cancelled.
		case <-r.Context().Done():
			backoff.Stop()

			// Return context error.
			err = r.Context().Err()
			return nil, err

		// Backoff for time.
		case <-backoff.C:
		}
	}
}

// DoOnce wraps an underlying http.Client{}.Do() to perform our wrapped request type:
// rewinding response body to permit reuse, signing request data when SignFunc provided,
// marking erroring hosts, updating retry attempt counts and setting backoff from header.
func (c *Client) DoOnce(r *Request) (rsp *http.Response, retry bool, err error) {
	if r.attempts > c.retries {
		// Ensure request hasn't reached max number of attempts.
		err = fmt.Errorf("httpclient: reached max retries (%d)", c.retries)
		return
	}

	// Update no.
	// attempts.
	r.attempts++

	// Reset backoff.
	r.backoff = 0

	// Perform main routine.
	rsp, retry, err = c.do(r)

	if rsp != nil {
		// Log successful rsp.
		r.Entry.Info(rsp.Status)
		return
	}

	// Log any errors.
	r.Entry.Error(err)

	switch {
	case !retry:
		// If they were told not to
		// retry, also set number of
		// attempts to prevent retry.
		r.attempts = c.retries + 1

	case r.attempts > c.retries:
		// On max retries, mark this as
		// a "badhost", i.e. is erroring.
		c.badHosts.Set(r.Host, struct{}{})

		// Ensure retry flag is unset
		// when reached max attempts.
		retry = false

	case c.badHosts.Has(r.Host):
		// When retry is still permitted,
		// check host hasn't been marked
		// as a "badhost", i.e. erroring.
		r.attempts = c.retries + 1
		retry = false
	}

	return
}

// do performs the "meat" of DoOnce(), but it's separated out to allow
// easier wrapping of the response, retry, error returns with further logic.
func (c *Client) do(r *Request) (rsp *http.Response, retry bool, err error) {
	// Perform the HTTP request.
	rsp, err = c.client.Do(r.Request)
	if err != nil {

		if errorsv2.IsV2(err,
			context.DeadlineExceeded,
			context.Canceled,
			ErrReservedAddr,
		) {
			// Non-retryable errors.
			return nil, false, err
		}

		if errstr := err.Error(); //
		strings.Contains(errstr, "stopped after 10 redirects") ||
			strings.Contains(errstr, "tls: ") ||
			strings.Contains(errstr, "x509: ") {
			// These error types aren't wrapped
			// so we have to check the error string.
			// All are unrecoverable!
			return nil, false, err
		}

		if dnserr := errorsv2.AsV2[*net.DNSError](err); //
		dnserr != nil && dnserr.IsNotFound {
			// DNS lookup failure, this domain does not exist
			return nil, false, gtserror.SetNotFound(err)
		}

		// A retryable error.
		return nil, true, err

	} else if rsp.StatusCode >= 500 ||
		rsp.StatusCode == http.StatusTooManyRequests {

		// Codes over 500 (and 429: too many requests)
		// are generally temporary errors. For these
		// we replace the response with a loggable error.
		err = fmt.Errorf(`http response: %s`, rsp.Status)

		// Search for a provided "Retry-After" header value.
		if after := rsp.Header.Get("Retry-After"); after != "" {

			// Get cur time.
			now := time.Now()

			if u, _ := strconv.ParseUint(after, 10, 32); u != 0 {
				// An integer no. of backoff seconds was provided.
				r.backoff = time.Duration(u) * time.Second // #nosec G115 -- We clamp backoff below.
			} else if at, _ := http.ParseTime(after); !at.Before(now) {
				// An HTTP formatted future date-time was provided.
				r.backoff = at.Sub(now)
			}

			// Don't let their provided backoff exceed our max.
			if max := baseBackoff * time.Duration(c.retries); // #nosec G115 -- We control c.retries.
			r.backoff > max {
				r.backoff = max
			}
		}

		// Unset + close rsp.
		_ = rsp.Body.Close()
		return nil, true, err
	}

	// Seperate the body implementers.
	rbody := (io.Reader)(rsp.Body)
	cbody := (io.Closer)(rsp.Body)

	// Wrap closer to ensure body drained BEFORE close.
	cbody = iotools.CloserAfterCallback(cbody, func() {
		_, _ = discard.ReadFrom(rbody)
	})

	// Set the wrapped response body.
	rsp.Body = &iotools.ReadCloserType{
		Reader: rbody,
		Closer: cbody,
	}

	return rsp, true, nil
}

// cast discard writer to full interface it supports.
var discard = io.Discard.(interface { //nolint
	io.Writer
	io.StringWriter
	io.ReaderFrom
})
