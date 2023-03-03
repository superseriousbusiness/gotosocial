package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/go-fed/httpsig"
)

var (
	// this mimics an untyped error returned by httpsig when no signature is present;
	// define it here so that we can use it to decide what to log without hitting
	// performance too hard
	noSignatureError    = fmt.Sprintf("neither %q nor %q have signature parameters", httpsig.Signature, httpsig.Authorization)
	signatureHeader     = string(httpsig.Signature)
	authorizationHeader = string(httpsig.Authorization)
)

// SignatureCheck returns a gin middleware for checking http signatures.
//
// The middleware first checks whether an incoming http request has been http-signed with a well-formed signature.
//
// If so, it will check if the domain that signed the request is permitted to access the server, using the provided isURIBlocked function.
//
// If it is permitted, the handler will set the key verifier and the signature in the gin context for use down the line.
//
// If the domain is blocked, the middleware will abort the request chain instead with http code 403 forbidden.
//
// In case of an error, the request will be aborted with http code 500 internal server error.
func SignatureCheck(isURIBlocked func(context.Context, *url.URL) (bool, db.Error)) func(*gin.Context) {
	return func(c *gin.Context) {
		// Acquire ctx from gin request.
		ctx := c.Request.Context()

		// create the verifier from the request, this will error if the request wasn't signed
		verifier, err := httpsig.NewVerifier(c.Request)
		if err != nil {
			// Something went wrong, so we need to return regardless, but only actually
			// *abort* the request with 401 if a signature was present but malformed
			if err.Error() != noSignatureError {
				log.Debugf(ctx, "http signature was present but invalid: %s", err)
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}

		// The request was signed!
		// The key ID should be given in the signature so that we know where to fetch it from the remote server.
		// This will be something like https://example.org/users/whatever_requesting_user#main-key
		requestingPublicKeyIDString := verifier.KeyId()
		requestingPublicKeyID, err := url.Parse(requestingPublicKeyIDString)
		if err != nil {
			log.Debugf(ctx, "http signature requesting public key id %s could not be parsed as a url: %s", requestingPublicKeyIDString, err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if requestingPublicKeyID == nil {
			// Key can sometimes be nil, according to url parse function:
			// 'Trying to parse a hostname and path without a scheme is invalid but may not necessarily return an error, due to parsing ambiguities'
			log.Debugf(ctx, "http signature requesting public key id %s was nil after parsing as a url", requestingPublicKeyIDString)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// we managed to parse the url!
		// if the domain is blocked we want to bail as early as possible
		if blocked, err := isURIBlocked(c.Request.Context(), requestingPublicKeyID); err != nil {
			log.Errorf(ctx, "could not tell if domain %s was blocked or not: %s", requestingPublicKeyID.Host, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		} else if blocked {
			log.Infof(ctx, "domain %s is blocked", requestingPublicKeyID.Host)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// assume signature was set on Signature header (most common behavior),
		// but fall back to Authorization header if necessary
		var signature string
		if s := c.GetHeader(signatureHeader); s != "" {
			signature = s
		} else {
			signature = c.GetHeader(authorizationHeader)
		}

		// set the verifier and signature on the context here to save some work further down the line
		c.Set(string(ap.ContextRequestingPublicKeyVerifier), verifier)
		c.Set(string(ap.ContextRequestingPublicKeySignature), signature)
	}
}
