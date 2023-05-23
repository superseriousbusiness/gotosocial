// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package users

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// InboxPOSTHandler deals with incoming POST requests to an actor's inbox.
// Eg., POST to https://example.org/users/whatever/inbox.
func (m *Module) InboxPOSTHandler(c *gin.Context) {
	_, err := m.processor.Fedi().InboxPost(apiutil.TransferSignatureContext(c), c.Writer, c.Request)
	if err != nil {
		if errWithCode := new(gtserror.WithCode); errors.As(err, errWithCode) {
			// An errWithCode was returned to us, we can be nice and
			// try to return something informative to the caller.
			apiutil.ErrorHandler(c, *errWithCode, m.processor.InstanceGetV1)
		} else {
			// Something else went wrong, and someone forgot to return
			// an errWithCode! It's chill though. Log the error but don't
			// return it to the caller, to avoid leaking internals.
			log.WithContext(c.Request.Context()).Errorf("returning Bad Request to caller, err was: %q", err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err), m.processor.InstanceGetV1)
		}
	} else {
		// Inbox POST body was Accepted for processing.
		c.JSON(http.StatusAccepted, gin.H{"status": http.StatusText(http.StatusAccepted)})
	}
}
