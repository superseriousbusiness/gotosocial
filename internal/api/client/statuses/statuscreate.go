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

package statuses

import (
	"errors"
	"fmt"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/form/v4"
)

// StatusCreatePOSTHandler swagger:operation POST /api/v1/statuses statusCreate
//
// Create a new status using the given form field parameters.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
//
// The 'interaction_policy' field can be used to set an interaction policy for this status.
//
// If submitting using form data, use the following pattern to set an interaction policy:
//
// `interaction_policy[INTERACTION_TYPE][CONDITION][INDEX]=Value`
//
// For example: `interaction_policy[can_reply][always][0]=author`
//
// Using `curl` this might look something like:
//
// `curl -F 'interaction_policy[can_reply][always][0]=author' -F 'interaction_policy[can_reply][always][1]=followers' [... other form fields ...]`
//
// The JSON equivalent would be:
//
// `curl -H 'Content-Type: application/json' -d '{"interaction_policy":{"can_reply":{"always":["author","followers"]}} [... other json fields ...]}'`
//
// The server will perform some normalization on the submitted policy so that you can't submit something totally invalid.
//
//	---
//	tags:
//	- statuses
//
//	consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	parameters:
//	-
//		name: status
//		x-go-name: Status
//		description: |-
//			Text content of the status.
//			If media_ids is provided, this becomes optional.
//			Attaching a poll is optional while status is provided.
//		type: string
//		in: formData
//	-
//		name: media_ids[]
//		x-go-name: MediaIDs
//		description: |-
//			Array of Attachment ids to be attached as media.
//			If provided, status becomes optional, and poll cannot be used.
//
//			If the status is being submitted as a form, the key is 'media_ids[]',
//			but if it's json or xml, the key is 'media_ids'.
//		type: array
//		items:
//			type: string
//		in: formData
//		collectionFormat: multi
//		uniqueItems: true
//	-
//		name: poll[options][]
//		x-go-name: PollOptions
//		description: |-
//			Array of possible poll answers.
//			If provided, media_ids cannot be used, and poll[expires_in] must be provided.
//		type: array
//		items:
//			type: string
//		in: formData
//		collectionFormat: multi
//		uniqueItems: true
//	-
//		name: poll[expires_in]
//		x-go-name: PollExpiresIn
//		description: |-
//			Duration the poll should be open, in seconds.
//			If provided, media_ids cannot be used, and poll[options] must be provided.
//		type: integer
//		format: int64
//		in: formData
//	-
//		name: poll[multiple]
//		x-go-name: PollMultiple
//		description: Allow multiple choices on this poll.
//		type: boolean
//		default: false
//		in: formData
//	-
//		name: poll[hide_totals]
//		x-go-name: PollHideTotals
//		description: Hide vote counts until the poll ends.
//		type: boolean
//		default: true
//		in: formData
//	-
//		name: in_reply_to_id
//		x-go-name: InReplyToID
//		description: ID of the status being replied to, if status is a reply.
//		type: string
//		in: formData
//	-
//		name: sensitive
//		x-go-name: Sensitive
//		description: Status and attached media should be marked as sensitive.
//		type: boolean
//		in: formData
//	-
//		name: spoiler_text
//		x-go-name: SpoilerText
//		description: |-
//			Text to be shown as a warning or subject before the actual content.
//			Statuses are generally collapsed behind this field.
//		type: string
//		in: formData
//	-
//		name: visibility
//		x-go-name: Visibility
//		description: Visibility of the posted status.
//		type: string
//		enum:
//			- public
//			- unlisted
//			- private
//			- mutuals_only
//			- direct
//		in: formData
//	-
//		name: local_only
//		x-go-name: LocalOnly
//		description: >-
//			If set to true, this status will be "local only" and will NOT be federated beyond the local timeline(s).
//			If set to false (default), this status will be federated to your followers beyond the local timeline(s).
//		type: boolean
//		in: formData
//		default: false
//	-
//		name: federated
//		x-go-name: Federated
//		description: >-
//			***DEPRECATED***. Included for back compat only. Only used if set and local_only is not yet.
//			If set to true, this status will be federated beyond the local timeline(s).
//			If set to false, this status will NOT be federated beyond the local timeline(s).
//		in: formData
//		type: boolean
//	-
//		name: scheduled_at
//		x-go-name: ScheduledAt
//		description: |-
//			ISO 8601 Datetime at which to schedule a status.
//
//			Providing this parameter with a *future* time will cause ScheduledStatus to be returned instead of Status.
//			Must be at least 5 minutes in the future.
//			This feature isn't implemented yet.
//
//			Providing this parameter with a *past* time will cause the status to be backdated,
//			and will not push it to the user's followers. This is intended for importing old statuses.
//		type: string
//		format: date-time
//		in: formData
//	-
//		name: language
//		x-go-name: Language
//		description: ISO 639 language code for this status.
//		type: string
//		in: formData
//	-
//		name: content_type
//		x-go-name: ContentType
//		description: Content type to use when parsing this status.
//		type: string
//		enum:
//			- text/plain
//			- text/markdown
//		in: formData
//	-
//		name: interaction_policy[can_favourite][always][0]
//		in: formData
//		description: Nth entry for interaction_policy.can_favourite.always.
//		type: string
//	-
//		name: interaction_policy[can_favourite][with_approval][0]
//		in: formData
//		description: Nth entry for interaction_policy.can_favourite.with_approval.
//		type: string
//	-
//		name: interaction_policy[can_reply][always][0]
//		in: formData
//		description: Nth entry for interaction_policy.can_reply.always.
//		type: string
//	-
//		name: interaction_policy[can_reply][with_approval][0]
//		in: formData
//		description: Nth entry for interaction_policy.can_reply.with_approval.
//		type: string
//	-
//		name: interaction_policy[can_reblog][always][0]
//		in: formData
//		description: Nth entry for interaction_policy.can_reblog.always.
//		type: string
//	-
//		name: interaction_policy[can_reblog][with_approval][0]
//		in: formData
//		description: Nth entry for interaction_policy.can_reblog.with_approval.
//		type: string
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer:
//		- write:statuses
//
//	responses:
//		'200':
//			description: "The newly created status."
//			schema:
//				"$ref": "#/definitions/status"
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
//		'500':
//			description: internal server error
//		'501':
//			description: scheduled_at was set, but this feature is not yet implemented
func (m *Module) StatusCreatePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
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

	form, errWithCode := parseStatusCreateForm(c)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// DO NOT COMMIT THIS UNCOMMENTED, IT WILL CAUSE MASS CHAOS.
	// this is being left in as an ode to kim's shitposting.
	//
	// user := authed.Account.DisplayName
	// if user == "" {
	// 	user = authed.Account.Username
	// }
	// form.Status += "\n\nsent from " + user + "'s iphone\n"

	apiStatus, errWithCode := m.processor.Status().Create(
		c.Request.Context(),
		authed.Account,
		authed.Application,
		form,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiStatus)
}

// intPolicyFormBinding satisfies gin's binding.Binding interface.
// Should only be used specifically for multipart/form-data MIME type.
type intPolicyFormBinding struct{}

func (i intPolicyFormBinding) Name() string {
	return "InteractionPolicy"
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

func parseStatusCreateForm(c *gin.Context) (*apimodel.StatusCreateRequest, gtserror.WithCode) {
	form := new(apimodel.StatusCreateRequest)

	switch ct := c.ContentType(); ct {
	case binding.MIMEJSON:
		// Just bind with default json binding.
		if err := c.ShouldBindWith(form, binding.JSON); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

	case binding.MIMEPOSTForm:
		// Bind with default form binding first.
		if err := c.ShouldBindWith(form, binding.FormPost); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

		// Now do custom binding.
		intReqForm := new(apimodel.StatusInteractionPolicyForm)
		if err := c.ShouldBindWith(intReqForm, intPolicyFormBinding{}); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

		form.InteractionPolicy = intReqForm.InteractionPolicy

	case binding.MIMEMultipartPOSTForm:
		// Bind with default form binding first.
		if err := c.ShouldBindWith(form, binding.FormMultipart); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

		// Now do custom binding.
		intReqForm := new(apimodel.StatusInteractionPolicyForm)
		if err := c.ShouldBindWith(intReqForm, intPolicyFormBinding{}); err != nil {
			return nil, gtserror.NewErrorBadRequest(
				err,
				err.Error(),
			)
		}

		form.InteractionPolicy = intReqForm.InteractionPolicy

	default:
		text := fmt.Sprintf("content-type %s not supported for this endpoint; supported content-types are %s, %s, %s",
			ct, binding.MIMEJSON, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm)
		return nil, gtserror.NewErrorNotAcceptable(errors.New(text), text)
	}

	// Check if the deprecated "federated" field was
	// set in lieu of "local_only", and use it if so.
	if form.LocalOnly == nil && form.Federated != nil { // nolint:staticcheck
		form.LocalOnly = util.Ptr(!*form.Federated) // nolint:staticcheck
	}

	// Normalize poll expiry time if a poll was given.
	if form.Poll != nil && form.Poll.ExpiresInI != nil {

		// If we parsed this as JSON, expires_in
		// may be either a float64 or a string.
		expiresIn, err := apiutil.ParseDuration(
			form.Poll.ExpiresInI,
			"expires_in",
		)
		if err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		form.Poll.ExpiresIn = util.PtrOrZero(expiresIn)
	}

	return form, nil
}
