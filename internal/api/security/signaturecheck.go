package security

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// SignatureCheck checks whether an incoming http request has been signed. If so, it will check if the domain
// that signed the request is permitted to access the server. If it is permitted, the handler will set the key
// verifier and the signature in the gin context for use down the line.
func (m *Module) SignatureCheck(c *gin.Context) {
	l := logrus.WithField("func", "DomainBlockChecker")

	// create the verifier from the request
	// if the request is signed, it will have a signature header
	verifier, err := httpsig.NewVerifier(c.Request)
	if err == nil {
		// the request was signed!

		// The key ID should be given in the signature so that we know where to fetch it from the remote server.
		// This will be something like https://example.org/users/whatever_requesting_user#main-key
		requestingPublicKeyID, err := url.Parse(verifier.KeyId())
		if err == nil && requestingPublicKeyID != nil {
			// we managed to parse the url!

			// if the domain is blocked we want to bail as early as possible
			blocked, err := m.db.IsURIBlocked(c.Request.Context(), requestingPublicKeyID)
			if err != nil {
				l.Errorf("could not tell if domain %s was blocked or not: %s", requestingPublicKeyID.Host, err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if blocked {
				l.Infof("domain %s is blocked", requestingPublicKeyID.Host)
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			// set the verifier and signature on the context here to save some work further down the line
			c.Set(string(util.APRequestingPublicKeyVerifier), verifier)
			signature := c.GetHeader("Signature")
			if signature != "" {
				c.Set(string(util.APRequestingPublicKeySignature), signature)
			}
		}
	}
}
