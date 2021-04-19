/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package media

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	mastotypes "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (m *MediaModule) MediaCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.MustAuth(c, true, true, true, true) // posting new media is serious business so we want *everything*
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// First check this user/account is permitted to create media
	// There's no point continuing otherwise.
	if authed.User.Disabled || !authed.User.Approved || !authed.Account.SuspendedAt.IsZero() {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled, not yet approved, or suspended"})
		return
	}

	// extract the media create form from the request context
	l.Tracef("parsing request form: %s", c.Request.Form)
	form := &mastotypes.AttachmentRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateMedia(form, m.config.MediaConfig); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// open the attachment and extract the bytes from it
	f, err := form.File.Open()
	if err != nil {
		l.Debugf("error opening attachment: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not open provided attachment: %s", err)})
		return
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		l.Debugf("error reading attachment: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not read provided attachment: %s", err)})
		return
	}
	if size == 0 {
		l.Debug("could not read provided attachment: size 0 bytes")
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read provided attachment: size 0 bytes"})
		return
	}

	// allow the mediaHandler to work its magic of processing the attachment bytes, and putting them in whatever storage backend we're using
	attachment, err := m.mediaHandler.ProcessLocalAttachment(buf.Bytes(), authed.Account.ID)
	if err != nil {
		l.Debugf("error reading attachment: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not process attachment: %s", err)})
		return
	}

	// now we need to add extra fields that the attachment processor doesn't know (from the form)
	// TODO: handle this inside mediaHandler.ProcessAttachment (just pass more params to it)

	// first description
	attachment.Description = form.Description

	// now parse the focus parameter
	// TODO: tidy this up into a separate function and just return an error so all the c.JSON and return calls are obviated
	var focusx, focusy float32
	if form.Focus != "" {
		spl := strings.Split(form.Focus, ",")
		if len(spl) != 2 {
			l.Debugf("improperly formatted focus %s", form.Focus)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("improperly formatted focus %s", form.Focus)})
			return
		}
		xStr := spl[0]
		yStr := spl[1]
		if xStr == "" || yStr == "" {
			l.Debugf("improperly formatted focus %s", form.Focus)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("improperly formatted focus %s", form.Focus)})
			return
		}
		fx, err := strconv.ParseFloat(xStr, 32)
		if err != nil {
			l.Debugf("improperly formatted focus %s: %s", form.Focus, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("improperly formatted focus %s", form.Focus)})
			return
		}
		if fx > 1 || fx < -1 {
			l.Debugf("improperly formatted focus %s", form.Focus)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("improperly formatted focus %s", form.Focus)})
			return
		}
		focusx = float32(fx)
		fy, err := strconv.ParseFloat(yStr, 32)
		if err != nil {
			l.Debugf("improperly formatted focus %s: %s", form.Focus, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("improperly formatted focus %s", form.Focus)})
			return
		}
		if fy > 1 || fy < -1 {
			l.Debugf("improperly formatted focus %s", form.Focus)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("improperly formatted focus %s", form.Focus)})
			return
		}
		focusy = float32(fy)
	}
	attachment.FileMeta.Focus.X = focusx
	attachment.FileMeta.Focus.Y = focusy

	// prepare the frontend representation now -- if there are any errors here at least we can bail without
	// having already put something in the database and then having to clean it up again (eugh)
	mastoAttachment, err := m.mastoConverter.AttachmentToMasto(attachment)
	if err != nil {
		l.Debugf("error parsing media attachment to frontend type: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error parsing media attachment to frontend type: %s", err)})
		return
	}

	// now we can confidently put the attachment in the database
	if err := m.db.Put(attachment); err != nil {
		l.Debugf("error storing media attachment in db: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error storing media attachment in db: %s", err)})
		return
	}

	// and return its frontend representation
	c.JSON(http.StatusAccepted, mastoAttachment)
}

func validateCreateMedia(form *mastotypes.AttachmentRequest, config *config.MediaConfig) error {
	// check there actually is a file attached and it's not size 0
	if form.File == nil || form.File.Size == 0 {
		return errors.New("no attachment given")
	}

	// a very superficial check to see if no size limits are exceeded
	// we still don't actually know which media types we're dealing with but the other handlers will go into more detail there
	maxSize := config.MaxVideoSize
	if config.MaxImageSize > maxSize {
		maxSize = config.MaxImageSize
	}
	if form.File.Size > int64(maxSize) {
		return fmt.Errorf("file size limit exceeded: limit is %d bytes but attachment was %d bytes", maxSize, form.File.Size)
	}

	if len(form.Description) < config.MinDescriptionChars || len(form.Description) > config.MaxDescriptionChars {
		return fmt.Errorf("image description length must be between %d and %d characters (inclusive), but provided image description was %d chars", config.MinDescriptionChars, config.MaxDescriptionChars, len(form.Description))
	}

	// TODO: validate focus here

	return nil
}
