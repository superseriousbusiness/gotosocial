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

package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// DomainKeysExpirePOSTHandler swagger:operation POST /api/v1/admin/domain_keys_expire domainKeysExpire
//
// Force expiry of cached public keys for all accounts on the given domain stored in your database.
//
// This is useful in cases where the remote domain has had to rotate their keys for whatever
// reason (security issue, data leak, routine safety procedure, etc), and your instance can no
// longer communicate with theirs properly using cached keys. A key marked as expired in this way
// will be lazily refetched next time a request is made to your instance signed by the owner of that
// key, so no further action should be required in order to reestablish communication with that domain.
//
// This endpoint is explicitly not for rotating your *own* keys, it only works for remote instances.
//
// Using this endpoint to expire keys for a domain that hasn't rotated all of their keys is not
// harmful and won't break federation, but it is pointless and will cause unnecessary requests to
// be performed.
//
//	---
//	tags:
//	- admin
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: domain
//		in: formData
//		description: |-
//			Domain to expire keys for.
//			Sample: example.org
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- admin:write
//
//	responses:
//		'202':
//			description: >-
//				Request accepted and will be processed.
//				Check the logs for progress / errors.
//			schema:
//				"$ref": "#/definitions/adminActionResponse"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'409':
//			description: >-
//				Conflict: There is already an admin action running that conflicts with this action.
//				Check the error message in the response body for more information. This is a temporary
//				error; it should be possible to process this action if you try again in a bit.
//		'500':
//			description: internal server error
func (m *Module) DomainKeysExpirePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeAdminWrite,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if !*authed.User.Admin {
		err := fmt.Errorf("user %s not an admin", authed.User.ID)
		apiutil.ErrorHandler(c, gtserror.NewErrorForbidden(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := new(apimodel.DomainKeysExpireRequest)
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateDomainKeysExpire(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	actionID, errWithCode := m.processor.Admin().DomainKeysExpire(
		c.Request.Context(),
		authed.Account,
		form.Domain,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, &apimodel.AdminActionResponse{
		ActionID: actionID,
	})
}

func validateDomainKeysExpire(form *apimodel.DomainKeysExpireRequest) error {
	form.Domain = strings.TrimSpace(form.Domain)
	if form.Domain == "" {
		return errors.New("no domain given")
	}

	if form.Domain == config.GetHost() || form.Domain == config.GetAccountDomain() {
		return errors.New("provided domain was this domain, but must be a remote domain")
	}

	return nil
}
