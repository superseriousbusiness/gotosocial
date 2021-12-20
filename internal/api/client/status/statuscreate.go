/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package status

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// StatusCreatePOSTHandler swagger:operation POST /api/v1/statuses statusCreate
//
// Create a new status.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
// ---
// tags:
// - statuses
//
// consumes:
// - application/json
// - application/xml
// - application/x-www-form-urlencoded
//
// produces:
// - application/json
//
// security:
// - OAuth2 Bearer:
//   - write:statuses
//
// responses:
//   '200':
//     description: "The newly created status."
//     schema:
//       "$ref": "#/definitions/status"
//   '401':
//      description: unauthorized
//   '400':
//      description: bad request
//   '404':
//      description: not found
//   '500':
//      description: internal error
func (m *Module) StatusCreatePOSTHandler(c *gin.Context) {
	l := logrus.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	// First check this user/account is permitted to post new statuses.
	// There's no point continuing otherwise.
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled, not yet approved, or suspended"})
		return
	}

	// extract the status create form from the request context
	l.Debugf("parsing request form: %s", c.Request.Form)
	form := &model.AdvancedStatusCreateForm{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}
	l.Debugf("handling status request form: %+v", form)

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateStatus(form); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiStatus, err := m.processor.StatusCreate(c.Request.Context(), authed, form)
	if err != nil {
		l.Debugf("error processing status create: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	c.JSON(http.StatusOK, apiStatus)
}

func validateCreateStatus(form *model.AdvancedStatusCreateForm) error {
	// validate that, structurally, we have a valid status/post
	if form.Status == "" && form.MediaIDs == nil && form.Poll == nil {
		return errors.New("no status, media, or poll provided")
	}

	if form.MediaIDs != nil && form.Poll != nil {
		return errors.New("can't post media + poll in same status")
	}

	keys := config.Keys
	maxChars := viper.GetInt(keys.StatusesMaxChars)
	maxMediaFiles := viper.GetInt(keys.StatusesMediaMaxFiles)
	maxPollOptions := viper.GetInt(keys.StatusesPollMaxOptions)
	maxPollChars := viper.GetInt(keys.StatusesPollOptionMaxChars)
	maxCwChars := viper.GetInt(keys.StatusesCWMaxChars)

	// validate status
	if form.Status != "" {
		if len(form.Status) > maxChars {
			return fmt.Errorf("status too long, %d characters provided but limit is %d", len(form.Status), maxChars)
		}
	}

	// validate media attachments
	if len(form.MediaIDs) > maxMediaFiles {
		return fmt.Errorf("too many media files attached to status, %d attached but limit is %d", len(form.MediaIDs), maxMediaFiles)
	}

	// validate poll
	if form.Poll != nil {
		if form.Poll.Options == nil {
			return errors.New("poll with no options")
		}
		if len(form.Poll.Options) > maxPollOptions {
			return fmt.Errorf("too many poll options provided, %d provided but limit is %d", len(form.Poll.Options), maxPollOptions)
		}
		for _, p := range form.Poll.Options {
			if len(p) > maxPollChars {
				return fmt.Errorf("poll option too long, %d characters provided but limit is %d", len(p), maxPollChars)
			}
		}
	}

	// validate spoiler text/cw
	if form.SpoilerText != "" {
		if len(form.SpoilerText) > maxCwChars {
			return fmt.Errorf("content-warning/spoilertext too long, %d characters provided but limit is %d", len(form.SpoilerText), maxCwChars)
		}
	}

	// validate post language
	if form.Language != "" {
		if err := validate.Language(form.Language); err != nil {
			return err
		}
	}

	return nil
}
