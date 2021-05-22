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

package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AccountUpdateCredentialsPATCHHandler allows a user to modify their account/profile settings.
// It should be served as a PATCH at /api/v1/accounts/update_credentials
//
// TODO: this can be optimized massively by building up a picture of what we want the new account
// details to be, and then inserting it all in the database at once. As it is, we do queries one-by-one
// which is not gonna make the database very happy when lots of requests are going through.
// This way it would also be safer because the update won't happen until *all* the fields are validated.
// Otherwise we risk doing a partial update and that's gonna cause probllleeemmmsss.
func (m *Module) AccountUpdateCredentialsPATCHHandler(c *gin.Context) {
	l := m.log.WithField("func", "accountUpdateCredentialsPATCHHandler")
	authed, err := oauth.Authed(c, true, false, false, true)
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("retrieved account %+v", authed.Account.ID)

	l.Debugf("parsing request form %s", c.Request.Form)
	form := &model.UpdateCredentialsRequest{}
	if err := c.ShouldBind(&form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// if everything on the form is nil, then nothing has been set and we shouldn't continue
	if form.Discoverable == nil && form.Bot == nil && form.DisplayName == nil && form.Note == nil && form.Avatar == nil && form.Header == nil && form.Locked == nil && form.Source == nil && form.FieldsAttributes == nil {
		l.Debugf("could not parse form from request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty form submitted"})
		return
	}

	acctSensitive, err := m.processor.AccountUpdate(authed, form)
	if err != nil {
		l.Debugf("could not update account: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l.Tracef("conversion successful, returning OK and mastosensitive account %+v", acctSensitive)
	c.JSON(http.StatusOK, acctSensitive)
}
