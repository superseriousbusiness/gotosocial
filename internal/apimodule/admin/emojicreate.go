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

package admin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	mastotypes "github.com/superseriousbusiness/gotosocial/internal/mastotypes/mastomodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (m *Module) emojiCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func":        "emojiCreatePOSTHandler",
		"request_uri": c.Request.RequestURI,
		"user_agent":  c.Request.UserAgent(),
		"origin_ip":   c.ClientIP(),
	})

	// make sure we're authed with an admin account
	authed, err := oauth.MustAuth(c, true, true, true, true) // posting a status is serious business so we want *everything*
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

	// extract the media create form from the request context
	l.Tracef("parsing request form: %+v", c.Request.Form)
	form := &mastotypes.EmojiCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		l.Debugf("error parsing form %+v: %s", c.Request.Form, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not parse form: %s", err)})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateEmoji(form); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// open the emoji and extract the bytes from it
	f, err := form.Image.Open()
	if err != nil {
		l.Debugf("error opening emoji: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not open provided emoji: %s", err)})
		return
	}
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, f)
	if err != nil {
		l.Debugf("error reading emoji: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not read provided emoji: %s", err)})
		return
	}
	if size == 0 {
		l.Debug("could not read provided emoji: size 0 bytes")
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read provided emoji: size 0 bytes"})
		return
	}

	// allow the mediaHandler to work its magic of processing the emoji bytes, and putting them in whatever storage backend we're using
	emoji, err := m.mediaHandler.ProcessLocalEmoji(buf.Bytes(), form.Shortcode)
	if err != nil {
		l.Debugf("error reading emoji: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not process emoji: %s", err)})
		return
	}

	mastoEmoji, err := m.mastoConverter.EmojiToMasto(emoji)
	if err != nil {
		l.Debugf("error converting emoji to mastotype: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not convert emoji: %s", err)})
		return
	}

	if err := m.db.Put(emoji); err != nil {
		l.Debugf("database error while processing emoji: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("database error while processing emoji: %s", err)})
		return
	}

	c.JSON(http.StatusOK, mastoEmoji)
}

func validateCreateEmoji(form *mastotypes.EmojiCreateRequest) error {
	// check there actually is an image attached and it's not size 0
	if form.Image == nil || form.Image.Size == 0 {
		return errors.New("no emoji given")
	}

	// a very superficial check to see if the media size limit is exceeded
	if form.Image.Size > media.EmojiMaxBytes {
		return fmt.Errorf("file size limit exceeded: limit is %d bytes but emoji was %d bytes", media.EmojiMaxBytes, form.Image.Size)
	}

	return util.ValidateEmojiShortcode(form.Shortcode)
}
