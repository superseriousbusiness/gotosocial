package instance

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// InstanceUpdatePATCHHandler swagger:operation PATCH /api/v1/instance instanceUpdate
//
// Update your instance information and/or upload a new avatar/header for the instance.
//
// This requires admin permissions on the instance.
//
// ---
// tags:
// - instance
//
// consumes:
// - multipart/form-data
//
// produces:
// - application/json
//
// parameters:
// - name: title
//   in: formData
//   description: Title to use for the instance.
//   type: string
//   maximum: 40
//   allowEmptyValue: true
// - name: contact_username
//   in: formData
//   description: |-
//     Username of the contact account.
//     This must be the username of an instance admin.
//   type: string
//   allowEmptyValue: true
// - name: contact_email
//   in: formData
//   description: Email address to use as the instance contact.
//   type: string
//   allowEmptyValue: true
// - name: short_description
//   in: formData
//   description: Short description of the instance.
//   type: string
//   maximum: 500
//   allowEmptyValue: true
// - name: description
//   in: formData
//   description: Longer description of the instance.
//   type: string
//   maximum: 5000
//   allowEmptyValue: true
// - name: terms
//   in: formData
//   description: Terms and conditions of the instance.
//   type: string
//   maximum: 5000
//   allowEmptyValue: true
// - name: avatar
//   in: formData
//   description: Avatar of the instance.
//   type: file
// - name: header
//   in: formData
//   description: Header of the instance.
//   type: file
//
// security:
// - OAuth2 Bearer:
//   - admin
//
// responses:
//   '200':
//     description: "The newly updated instance."
//     schema:
//       "$ref": "#/definitions/instance"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
func (m *Module) InstanceUpdatePATCHHandler(c *gin.Context) {
	l := logrus.WithField("func", "InstanceUpdatePATCHHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	// only admins can update instance settings
	if !authed.User.Admin {
		l.Debug("user is not an admin so cannot update instance settings")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not an admin"})
		return
	}

	l.Debug("parsing request form")
	form := &model.InstanceSettingsUpdateRequest{}
	if err := c.ShouldBind(&form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l.Debugf("parsed form: %+v", form)

	// if everything on the form is nil, then nothing has been set and we shouldn't continue
	if form.Title == nil && form.ContactUsername == nil && form.ContactEmail == nil && form.ShortDescription == nil && form.Description == nil && form.Terms == nil && form.Avatar == nil && form.Header == nil {
		l.Debugf("could not parse form from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty form submitted"})
		return
	}

	i, errWithCode := m.processor.InstancePatch(c.Request.Context(), form)
	if errWithCode != nil {
		l.Debugf("error with instance patch request: %s", errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, i)
}
