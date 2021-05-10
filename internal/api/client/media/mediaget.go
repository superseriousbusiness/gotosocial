package media

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

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

// MediaGETHandler allows the owner of an attachment to get information about that attachment before it's used in a status.
func (m *Module) MediaGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	attachmentID := c.GetString(IDKey)
	if attachmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no attachment ID given in request"})
		return
	}

	attachment, errWithCode := m.processor.MediaGet(authed, attachmentID)
	if err != nil {
		c.JSON(errWithCode.Code(),gin.H{"error":  errWithCode.Safe()})
		return
	}

	c.JSON(http.StatusOK, attachment)
}
