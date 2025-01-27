package gzip

import (
	"bufio"
	"compress/gzip"
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
	HuffmanOnly        = gzip.HuffmanOnly
)

func Gzip(level int, options ...Option) gin.HandlerFunc {
	return newGzipHandler(level, options...).Handle
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

func (g *gzipWriter) Flush() {
	_ = g.writer.Flush()
	g.ResponseWriter.Flush()
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

// Ensure gzipWriter implements the http.Hijacker interface.
// This will cause a compile-time error if gzipWriter does not implement all methods of the http.Hijacker interface.
var _ http.Hijacker = (*gzipWriter)(nil)

// Hijack allows the caller to take over the connection from the HTTP server.
// After a call to Hijack, the HTTP server library will not do anything else with the connection.
// It becomes the caller's responsibility to manage and close the connection.
//
// It returns the underlying net.Conn, a buffered reader/writer for the connection, and an error
// if the ResponseWriter does not support the Hijacker interface.
func (g *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := g.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}
