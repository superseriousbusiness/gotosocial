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

package interactionpolicies

import (
	"fmt"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/form/v4"
)

// PoliciesDefaultsPATCHHandler swagger:operation PATCH /api/v1/interaction_policies/defaults policiesDefaultsUpdate
//
// Update default interaction policies per visibility level for new statuses created by you.
//
// If submitting using form data, use the following pattern:
//
// `VISIBILITY[INTERACTION_TYPE][CONDITION][INDEX]=Value`
//
// For example: `public[can_reply][always][0]=author`
//
// Using `curl` this might look something like:
//
// `curl -F 'public[can_reply][always][0]=author' -F 'public[can_reply][always][1]=followers'`
//
// The JSON equivalent would be:
//
// `curl -H 'Content-Type: application/json' -d '{"public":{"can_reply":{"always":["author","followers"]}}}'`
//
// Any visibility level left unspecified in the request body will be returned to the default.
//
// Ie., in the example above, "public" would be updated, but "unlisted", "private", and "direct" would be reset to defaults.
//
// The server will perform some normalization on submitted policies so that you can't submit totally invalid policies.
//
//	---
//	tags:
//	- interaction_policies
//
//	consumes:
//	- multipart/form-data
//	- application/x-www-form-urlencoded
//	- application/json
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: public[can_favourite][always][0]
//		in: formData
//		description: Nth entry for public.can_favourite.always.
//		type: string
//	-
//		name: public[can_favourite][with_approval][0]
//		in: formData
//		description: Nth entry for public.can_favourite.with_approval.
//		type: string
//	-
//		name: public[can_reply][always][0]
//		in: formData
//		description: Nth entry for public.can_reply.always.
//		type: string
//	-
//		name: public[can_reply][with_approval][0]
//		in: formData
//		description: Nth entry for public.can_reply.with_approval.
//		type: string
//	-
//		name: public[can_reblog][always][0]
//		in: formData
//		description: Nth entry for public.can_reblog.always.
//		type: string
//	-
//		name: public[can_reblog][with_approval][0]
//		in: formData
//		description: Nth entry for public.can_reblog.with_approval.
//		type: string
//
//	-
//		name: unlisted[can_favourite][always][0]
//		in: formData
//		description: Nth entry for unlisted.can_favourite.always.
//		type: string
//	-
//		name: unlisted[can_favourite][with_approval][0]
//		in: formData
//		description: Nth entry for unlisted.can_favourite.with_approval.
//		type: string
//	-
//		name: unlisted[can_reply][always][0]
//		in: formData
//		description: Nth entry for unlisted.can_reply.always.
//		type: string
//	-
//		name: unlisted[can_reply][with_approval][0]
//		in: formData
//		description: Nth entry for unlisted.can_reply.with_approval.
//		type: string
//	-
//		name: unlisted[can_reblog][always][0]
//		in: formData
//		description: Nth entry for unlisted.can_reblog.always.
//		type: string
//	-
//		name: unlisted[can_reblog][with_approval][0]
//		in: formData
//		description: Nth entry for unlisted.can_reblog.with_approval.
//		type: string
//
//	-
//		name: private[can_favourite][always][0]
//		in: formData
//		description: Nth entry for private.can_favourite.always.
//		type: string
//	-
//		name: private[can_favourite][with_approval][0]
//		in: formData
//		description: Nth entry for private.can_favourite.with_approval.
//		type: string
//	-
//		name: private[can_reply][always][0]
//		in: formData
//		description: Nth entry for private.can_reply.always.
//		type: string
//	-
//		name: private[can_reply][with_approval][0]
//		in: formData
//		description: Nth entry for private.can_reply.with_approval.
//		type: string
//	-
//		name: private[can_reblog][always][0]
//		in: formData
//		description: Nth entry for private.can_reblog.always.
//		type: string
//	-
//		name: private[can_reblog][with_approval][0]
//		in: formData
//		description: Nth entry for private.can_reblog.with_approval.
//		type: string
//
//	-
//		name: direct[can_favourite][always][0]
//		in: formData
//		description: Nth entry for direct.can_favourite.always.
//		type: string
//	-
//		name: direct[can_favourite][with_approval][0]
//		in: formData
//		description: Nth entry for direct.can_favourite.with_approval.
//		type: string
//	-
//		name: direct[can_reply][always][0]
//		in: formData
//		description: Nth entry for direct.can_reply.always.
//		type: string
//	-
//		name: direct[can_reply][with_approval][0]
//		in: formData
//		description: Nth entry for direct.can_reply.with_approval.
//		type: string
//	-
//		name: direct[can_reblog][always][0]
//		in: formData
//		description: Nth entry for direct.can_reblog.always.
//		type: string
//	-
//		name: direct[can_reblog][with_approval][0]
//		in: formData
//		description: Nth entry for direct.can_reblog.with_approval.
//		type: string
//
//	security:
//	- OAuth2 Bearer:
//		- write:accounts
//
//	responses:
//		'200':
//			description: Updated default policies object containing a policy for each status visibility.
//			schema:
//				"$ref": "#/definitions/defaultPolicies"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'406':
//			description: not acceptable
//		'422':
//			description: unprocessable
//		'500':
//			description: internal server error
func (m *Module) PoliciesDefaultsPATCHHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteAccounts,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form, err := parseUpdatePoliciesForm(c)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Account().DefaultInteractionPoliciesUpdate(
		c.Request.Context(),
		authed.Account,
		form,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, resp)
}

// intPolicyFormBinding satisfies gin's binding.Binding interface.
// Should only be used specifically for multipart/form-data MIME type.
type intPolicyFormBinding struct {
	visibility string
}

func (i intPolicyFormBinding) Name() string {
	return i.visibility
}

func (intPolicyFormBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	// Change default namespace prefix and suffix to
	// allow correct parsing of the field attributes.
	decoder := form.NewDecoder()
	decoder.SetNamespacePrefix("[")
	decoder.SetNamespaceSuffix("]")

	return decoder.Decode(obj, req.Form)
}

// customBind does custom form binding for
// each visibility in the form data.
func customBind(
	c *gin.Context,
	form *apimodel.UpdateInteractionPoliciesRequest,
) error {
	for _, vis := range []string{
		"Direct",
		"Private",
		"Unlisted",
		"Public",
	} {
		if err := c.ShouldBindWith(
			form,
			intPolicyFormBinding{
				visibility: vis,
			},
		); err != nil {
			return fmt.Errorf("custom form binding failed: %w", err)
		}
	}

	return nil
}

func parseUpdatePoliciesForm(c *gin.Context) (*apimodel.UpdateInteractionPoliciesRequest, error) {
	form := new(apimodel.UpdateInteractionPoliciesRequest)

	switch ct := c.ContentType(); ct {
	case binding.MIMEJSON:
		// Just bind with default json binding.
		if err := c.ShouldBindWith(form, binding.JSON); err != nil {
			return nil, err
		}

	case binding.MIMEPOSTForm:
		// Bind with default form binding first.
		if err := c.ShouldBindWith(form, binding.FormPost); err != nil {
			return nil, err
		}

		// Now do custom binding.
		if err := customBind(c, form); err != nil {
			return nil, err
		}

	case binding.MIMEMultipartPOSTForm:
		// Bind with default form binding first.
		if err := c.ShouldBindWith(form, binding.FormMultipart); err != nil {
			return nil, err
		}

		// Now do custom binding.
		if err := customBind(c, form); err != nil {
			return nil, err
		}

	default:
		err := fmt.Errorf(
			"content-type %s not supported for this endpoint; supported content-types are %s, %s, %s",
			ct, binding.MIMEJSON, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm,
		)
		return nil, err
	}

	return form, nil
}
