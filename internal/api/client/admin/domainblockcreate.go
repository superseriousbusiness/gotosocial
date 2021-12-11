package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainBlocksPOSTHandler swagger:operation POST /api/v1/admin/domain_blocks domainBlockCreate
//
// Create one or more domain blocks, from a string or a file.
//
// Note that you have two options when using this endpoint: either you can set `import` to true
// and upload a file containing multiple domain blocks, JSON-formatted, or you can leave import as
// false, and just add one domain block.
//
// The format of the json file should be something like: `[{"domain":"example.org"},{"domain":"whatever.com","public_comment":"they smell"}]`
//
// ---
// tags:
// - admin
//
// consumes:
// - multipart/form-data
//
// produces:
// - application/json
//
// parameters:
// - name: import
//   in: query
//   description: |-
//     Signal that a list of domain blocks is being imported as a file.
//     If set to true, then 'domains' must be present as a JSON-formatted file.
//     If set to false, then 'domains' will be ignored, and 'domain' must be present.
//   type: boolean
// - name: domains
//   in: formData
//   description: |-
//     JSON-formatted list of domain blocks to import.
//     This is only used if `import` is set to true.
//   type: file
// - name: domain
//   in: formData
//   description: |-
//     Single domain to block.
//     Used only if `import` is not true.
//   type: string
// - name: obfuscate
//   in: formData
//   description: |-
//     Obfuscate the name of the domain when serving it publicly.
//     Eg., 'example.org' becomes something like 'ex***e.org'.
//     Used only if `import` is not true.
//   type: boolean
// - name: public_comment
//   in: formData
//   description: |-
//     Public comment about this domain block.
//     Will be displayed alongside the domain block if you choose to share blocks.
//     Used only if `import` is not true.
//   type: string
// - name: private_comment
//   in: formData
//   description: |-
//     Private comment about this domain block. Will only be shown to other admins, so this
//     is a useful way of internally keeping track of why a certain domain ended up blocked.
//     Used only if `import` is not true.
//   type: string
//
// security:
// - OAuth2 Bearer:
//   - admin
//
// responses:
//   '200':
//     description: |-
//       The newly created domain block, if `import` != `true`.
//       Note that if a list has been imported, then an `array` of newly created domain blocks will be returned instead.
//     schema:
//       "$ref": "#/definitions/domainBlock"
//   '403':
//      description: forbidden
//   '400':
//      description: bad request
func (m *Module) DomainBlocksPOSTHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func":        "DomainBlocksPOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})

	// make sure we're authed with an admin account
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if !authed.User.Admin {
		l.Debugf("user %s not an admin", authed.User.ID)
		c.JSON(http.StatusForbidden, gin.H{"error": "not an admin"})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	imp := false
	importString := c.Query(ImportQueryKey)
	if importString != "" {
		i, err := strconv.ParseBool(importString)
		if err != nil {
			l.Debugf("error parsing import string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse import query param"})
			return
		}
		imp = i
	}

	// extract the media create form from the request context
	l.Tracef("parsing request form: %+v", c.Request.Form)
	form := &model.DomainBlockCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		l.Debugf("error parsing form %+v: %s", c.Request.Form, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not parse form: %s", err)})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateDomainBlock(form, imp); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if imp {
		// we're importing multiple blocks
		domainBlocks, err := m.processor.AdminDomainBlocksImport(c.Request.Context(), authed, form)
		if err != nil {
			l.Debugf("error importing domain blocks: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, domainBlocks)
	} else {
		// we're just creating one block
		domainBlock, err := m.processor.AdminDomainBlockCreate(c.Request.Context(), authed, form)
		if err != nil {
			l.Debugf("error creating domain block: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, domainBlock)
	}
}

func validateCreateDomainBlock(form *model.DomainBlockCreateRequest, imp bool) error {
	if imp {
		if form.Domains.Size == 0 {
			return errors.New("import was specified but list of domains is empty")
		}
	} else {
		// add some more validation here later if necessary
		if form.Domain == "" {
			return errors.New("empty domain provided")
		}
	}

	return nil
}
