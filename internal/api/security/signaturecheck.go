package security

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-fed/httpsig"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (m *Module) SignatureCheck(c *gin.Context) {
	l := m.log.WithField("func", "DomainBlockChecker")

	// set this extra field for signature validation
	c.Request.Header.Set("host", m.config.Host)

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
			blockedDomain, err := m.blockedDomain(requestingPublicKeyID.Host)
			if err != nil {
				l.Errorf("could not tell if domain %s was blocked or not: %s", requestingPublicKeyID.Host, err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if blockedDomain {
				l.Infof("domain %s is blocked", requestingPublicKeyID.Host)
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			// set the verifier on the context here to save some work further down the line
			c.Set(string(util.APRequestingPublicKeyVerifier), verifier)
		}
	}
}

func (m *Module) blockedDomain(host string) (bool, error) {
	b := &gtsmodel.DomainBlock{}
	err := m.db.GetWhere([]db.Where{{Key: "domain", Value: host, CaseInsensitive: true}}, b)
	if err == nil {
		// block exists
		return true, nil
	}

	if _, ok := err.(db.ErrNoEntries); ok {
		// there are no entries so there's no block
		return false, nil
	}

	// there's an actual error
	return false, err
}
