package httpclient

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"runtime"
	"time"
)

// ErrIsInternalAddr is returned if a dialed address resolves to an internal IP address.
var ErrIsInternalAddr = errors.New("ip is private")

// ErrBodyTooLarge is returned when a received response body is above predefined limit (default 40MB).
var ErrBodyTooLarge = errors.New("body size too large")

// dialer is the net.Dialer used by all http.Transport{}'s.
var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
	Resolver:  &net.Resolver{Dial: nil},
}

// dialcontext wraps dialer.DialContext() to check for private addresses being dialed out to.
func dialcontext(ctx context.Context, network string, address string) (net.Conn, error) {
	// Attempt to dial out to requested address
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	// Cast remote addr so we can get IP
	addr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok /* this should never happen as HTTP is over tcp... */ {
		_ = conn.Close() // immediately close
		panic("unhandled network type")
	} else if addr.IP.IsPrivate() {
		_ = conn.Close() // immediately close
		return nil, ErrIsInternalAddr
	}

	return conn, nil
}

// Config ...
type Config struct {
	// MaxOpenConns limits the max number of concurrent open connections.
	MaxOpenConns int

	// MaxIdleConns: see http.Transport{}.MaxIdleConns
	MaxIdleConns int

	// ReadBufferSize: see http.Transport{}.ReadBufferSize
	ReadBufferSize int

	// WriteBufferSize: see http.Transport{}.WriteBufferSize
	WriteBufferSize int

	// MaxBodySize determines the maximum fetchable body size.
	MaxBodySize int64

	// AllowPrivateIPs allows dialing out to hosts in private address spaces.
	AllowPrivateIPs bool
}

// Client ...
type Client struct {
	client http.Client
	queue  chan struct{}
	bmax   int64
}

// New returns a new instance of Client initialized using configuration.
func New(cfg Config) *Client {
	var c Client

	if cfg.MaxOpenConns <= 0 {
		// By default base this value on GOMAXPROCS.
		maxprocs := runtime.GOMAXPROCS(0)
		cfg.MaxOpenConns = maxprocs * 10
	}

	if cfg.MaxIdleConns <= 0 {
		// By default base this value on MaxOpenConns
		cfg.MaxIdleConns = cfg.MaxOpenConns * 10
	}

	if cfg.MaxBodySize <= 0 {
		// By default set this to a reasonable 40MB
		cfg.MaxBodySize = 40 * 1024 * 1024
	}

	var dial func(context.Context, string, string) (net.Conn, error)

	if cfg.AllowPrivateIPs {
		// Bypase our wrapper that protects against private IP dialing
		dial = dialer.DialContext
	} else {
		// By default use wrapper to block private IPs
		dial = dialcontext
	}

	// Prepare client fields
	c.bmax = cfg.MaxBodySize
	c.queue = make(chan struct{}, cfg.MaxOpenConns)

	// Set underlying HTTP client roundtripper
	c.client.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		DialContext:           dial,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ReadBufferSize:        cfg.ReadBufferSize,
		WriteBufferSize:       cfg.WriteBufferSize,
	}

	return &c
}

// Do will perform given request when an available slot in the queue is available,
// and block until this time. For returned values, this follows the same semantics
// as the standard http.Client{}.Do() implementation, except that when response is
// returned the response body will be wrapped to release queue slot on close. i.e.
// YOU ABSOLUTELY MUST CLOSE THE RESPONSE BODY ON SUCCESSFUL REQUEST.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	select {
	// Request context cancelled
	case <-req.Context().Done():
		return nil, req.Context().Err()

	// Slot in queue acquired
	case c.queue <- struct{}{}:
		defer func() { <-c.queue }()
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
	cbody := (io.Closer)(rsp.Body)
	rbody := (io.Reader)(rsp.Body)

	var limit int64

	limit = rsp.ContentLength
	if limit = rsp.ContentLength; limit == -1 {
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
