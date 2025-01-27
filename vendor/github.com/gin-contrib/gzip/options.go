package gzip

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// DefaultExcludedExtentions is a predefined list of file extensions that should be excluded from gzip compression.
	// These extensions typically represent image files that are already compressed
	// and do not benefit from additional compression.
	DefaultExcludedExtentions = NewExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
	})
	// ErrUnsupportedContentEncoding is an error that indicates the content encoding
	// is not supported by the application.
	ErrUnsupportedContentEncoding = errors.New("unsupported content encoding")
)

// Option is an interface that defines a method to apply a configuration
// to a given config instance. Implementations of this interface can be
// used to modify the configuration settings of the logger.
type Option interface {
	apply(*config)
}

// Ensures that optionFunc implements the Option interface at compile time.
// If optionFunc does not implement Option, a compile-time error will occur.
var _ Option = (*optionFunc)(nil)

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

type config struct {
	excludedExtensions     ExcludedExtensions
	excludedPaths          ExcludedPaths
	excludedPathesRegexs   ExcludedPathesRegexs
	decompressFn           func(c *gin.Context)
	decompressOnly         bool
	customShouldCompressFn func(c *gin.Context) bool
}

// WithExcludedExtensions returns an Option that sets the ExcludedExtensions field of the Options struct.
// Parameters:
//   - args: []string - A slice of file extensions to exclude from gzip compression.
func WithExcludedExtensions(args []string) Option {
	return optionFunc(func(o *config) {
		o.excludedExtensions = NewExcludedExtensions(args)
	})
}

// WithExcludedPaths returns an Option that sets the ExcludedPaths field of the Options struct.
// Parameters:
//   - args: []string - A slice of paths to exclude from gzip compression.
func WithExcludedPaths(args []string) Option {
	return optionFunc(func(o *config) {
		o.excludedPaths = NewExcludedPaths(args)
	})
}

// WithExcludedPathsRegexs returns an Option that sets the ExcludedPathesRegexs field of the Options struct.
// Parameters:
//   - args: []string - A slice of regex patterns to exclude paths from gzip compression.
func WithExcludedPathsRegexs(args []string) Option {
	return optionFunc(func(o *config) {
		o.excludedPathesRegexs = NewExcludedPathesRegexs(args)
	})
}

// WithDecompressFn returns an Option that sets the DecompressFn field of the Options struct.
// Parameters:
//   - decompressFn: func(c *gin.Context) - A function to handle decompression of incoming requests.
func WithDecompressFn(decompressFn func(c *gin.Context)) Option {
	return optionFunc(func(o *config) {
		o.decompressFn = decompressFn
	})
}

// WithDecompressOnly is an option that configures the gzip middleware to only
// decompress incoming requests without compressing the responses. When this
// option is enabled, the middleware will set the DecompressOnly field of the
// Options struct to true.
func WithDecompressOnly() Option {
	return optionFunc(func(o *config) {
		o.decompressOnly = true
	})
}

// WithCustomShouldCompressFn returns an Option that sets the CustomShouldCompressFn field of the Options struct.
// Parameters:
//   - fn: func(c *gin.Context) bool - A function to determine if a request should be compressed.
//     The function should return true if the request should be compressed, false otherwise.
//     If the function returns false, the middleware will not compress the response.
//     If the function is nil, the middleware will use the default logic to determine
//     if the response should be compressed.
//
// Returns:
//   - Option - An option that sets the CustomShouldCompressFn field of the Options struct.
//
// Example:
//
//	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithCustomShouldCompressFn(func(c *gin.Context) bool {
//		return c.Request.URL.Path != "/no-compress"
//	})))
func WithCustomShouldCompressFn(fn func(c *gin.Context) bool) Option {
	return optionFunc(func(o *config) {
		o.customShouldCompressFn = fn
	})
}

// Using map for better lookup performance
type ExcludedExtensions map[string]struct{}

// NewExcludedExtensions creates a new ExcludedExtensions map from a slice of file extensions.
// Parameters:
//   - extensions: []string - A slice of file extensions to exclude from gzip compression.
//
// Returns:
//   - ExcludedExtensions - A map of excluded file extensions.
func NewExcludedExtensions(extensions []string) ExcludedExtensions {
	res := make(ExcludedExtensions, len(extensions))
	for _, e := range extensions {
		res[e] = struct{}{}
	}
	return res
}

// Contains checks if a given file extension is in the ExcludedExtensions map.
// Parameters:
//   - target: string - The file extension to check.
//
// Returns:
//   - bool - True if the extension is excluded, false otherwise.
func (e ExcludedExtensions) Contains(target string) bool {
	_, ok := e[target]
	return ok
}

type ExcludedPaths []string

// NewExcludedPaths creates a new ExcludedPaths slice from a slice of paths.
// Parameters:
//   - paths: []string - A slice of paths to exclude from gzip compression.
//
// Returns:
//   - ExcludedPaths - A slice of excluded paths.
func NewExcludedPaths(paths []string) ExcludedPaths {
	return ExcludedPaths(paths)
}

// Contains checks if a given request URI starts with any of the excluded paths.
// Parameters:
//   - requestURI: string - The request URI to check.
//
// Returns:
//   - bool - True if the URI starts with an excluded path, false otherwise.
func (e ExcludedPaths) Contains(requestURI string) bool {
	for _, path := range e {
		if strings.HasPrefix(requestURI, path) {
			return true
		}
	}
	return false
}

type ExcludedPathesRegexs []*regexp.Regexp

// NewExcludedPathesRegexs creates a new ExcludedPathesRegexs slice from a slice of regex patterns.
// Parameters:
//   - regexs: []string - A slice of regex patterns to exclude paths from gzip compression.
//
// Returns:
//   - ExcludedPathesRegexs - A slice of excluded path regex patterns.
func NewExcludedPathesRegexs(regexs []string) ExcludedPathesRegexs {
	result := make(ExcludedPathesRegexs, len(regexs))
	for i, reg := range regexs {
		result[i] = regexp.MustCompile(reg)
	}
	return result
}

// Contains checks if a given request URI matches any of the excluded path regex patterns.
// Parameters:
//   - requestURI: string - The request URI to check.
//
// Returns:
//   - bool - True if the URI matches an excluded path regex pattern, false otherwise.
func (e ExcludedPathesRegexs) Contains(requestURI string) bool {
	for _, reg := range e {
		if reg.MatchString(requestURI) {
			return true
		}
	}
	return false
}

// DefaultDecompressHandle is a middleware function for the Gin framework that
// decompresses the request body if it is gzip encoded. It checks if the request
// body is nil and returns immediately if it is. Otherwise, it attempts to create
// a new gzip reader from the request body. If an error occurs during this process,
// it aborts the request with a 400 Bad Request status and the error. If successful,
// it removes the "Content-Encoding" and "Content-Length" headers from the request
// and replaces the request body with the decompressed reader.
//
// Parameters:
//   - c: *gin.Context - The Gin context for the current request.
func DefaultDecompressHandle(c *gin.Context) {
	if c.Request.Body == nil {
		return
	}

	contentEncodingField := strings.Split(strings.ToLower(c.GetHeader("Content-Encoding")), ",")
	if len(contentEncodingField) == 0 { // nothing to decompress
		c.Next()

		return
	}

	toClose := make([]io.Closer, 0, len(contentEncodingField))
	defer func() {
		for i := len(toClose); i > 0; i-- {
			toClose[i-1].Close()
		}
	}()

	// parses multiply gzips like
	// Content-Encoding: gzip, gzip, gzip
	// allowed by RFC
	for i := 0; i < len(contentEncodingField); i++ {
		trimmedValue := strings.TrimSpace(contentEncodingField[i])

		if trimmedValue == "" {
			continue
		}

		if trimmedValue != "gzip" {
			// According to RFC 7231, Section 3.1.2.2:
			// https://www.rfc-editor.org/rfc/rfc7231#section-3.1.2.2
			// An origin server MAY respond with a status code of 415 (Unsupported
			// Media Type) if a representation in the request message has a content
			// coding that is not acceptable.
			_ = c.AbortWithError(http.StatusUnsupportedMediaType, ErrUnsupportedContentEncoding)
		}

		r, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)

			return
		}

		toClose = append(toClose, c.Request.Body)

		c.Request.Body = r
	}

	c.Request.Header.Del("Content-Encoding")
	c.Request.Header.Del("Content-Length")

	c.Next()
}
