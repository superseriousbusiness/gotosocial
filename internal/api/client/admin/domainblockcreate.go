package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// DomainBlocksPOSTHandler deals with the creation of a new domain block.
func (m *Module) DomainBlocksPOSTHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
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
		domainBlocks, err := m.processor.AdminDomainBlocksImport(authed, form)
		if err != nil {
			l.Debugf("error importing domain blocks: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, domainBlocks)
	} else {
		// we're just creating one block
		domainBlock, err := m.processor.AdminDomainBlockCreate(authed, form)
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
