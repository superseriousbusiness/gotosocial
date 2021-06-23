package instance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *Module) InstanceUpdatePATCHHandler(c *gin.Context) {
	l := m.log.WithField("func", "InstanceUpdatePATCHHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// only admins can update instance settings
	if !authed.User.Admin {
		l.Debug("user is not an admin so cannot update instance settings")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not an admin"})
		return
	}

	l.Debugf("parsing request form %s", c.Request.Form)
	form := &model.InstanceSettingsUpdateRequest{}
	if err := c.ShouldBind(&form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// if everything on the form is nil, then nothing has been set and we shouldn't continue
	if form.SiteTitle == nil && form.SiteContactUsername == nil && form.SiteContactEmail == nil && form.SiteShortDescription == nil && form.SiteDescription == nil && form.SiteTerms == nil && form.Avatar == nil && form.Header == nil {
		l.Debugf("could not parse form from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty form submitted"})
		return
	}

	i, errWithCode := m.processor.InstancePatch(form)
	if errWithCode != nil {
		l.Debugf("error with instance patch request: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, i)
}
