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
	"strconv"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// StatusCreatePOSTHandler swagger:operation POST /api/v1/statuses statusCreate
//
// Create a new status.
//
// The parameters can also be given in the body of the request, as JSON, if the content-type is set to 'application/json'.
// The parameters can also be given in the body of the request, as XML, if the content-type is set to 'application/xml'.
//
//	---
//	tags:
//	- statuses
//
//	consumes:
//	- application/json
//	- application/xml
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
//		name: media_ids
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
//			Providing this parameter will cause ScheduledStatus to be returned instead of Status.
//			Must be at least 5 minutes in the future.
//
//			This feature isn't implemented yet.
//		type: string
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
func (m *Module) StatusCreatePOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
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

	form := &apimodel.StatusCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
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

	if err := validateNormalizeCreateStatus(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

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

	c.JSON(http.StatusOK, apiStatus)
}

// validateNormalizeCreateStatus checks the form
// for disallowed combinations of attachments and
// overlength inputs.
//
// Side effect: normalizes the post's language tag.
func validateNormalizeCreateStatus(form *apimodel.StatusCreateRequest) error {
	hasStatus := form.Status != ""
	hasMedia := len(form.MediaIDs) != 0
	hasPoll := form.Poll != nil

	if !hasStatus && !hasMedia && !hasPoll {
		return errors.New("no status, media, or poll provided")
	}

	if hasMedia && hasPoll {
		return errors.New("can't post media + poll in same status")
	}

	maxChars := config.GetStatusesMaxChars()
	if length := len([]rune(form.Status)) + len([]rune(form.SpoilerText)); length > maxChars {
		return fmt.Errorf("status too long, %d characters provided (including spoiler/content warning) but limit is %d", length, maxChars)
	}

	maxMediaFiles := config.GetStatusesMediaMaxFiles()
	if len(form.MediaIDs) > maxMediaFiles {
		return fmt.Errorf("too many media files attached to status, %d attached but limit is %d", len(form.MediaIDs), maxMediaFiles)
	}

	if form.Poll != nil {
		if err := validateNormalizeCreatePoll(form); err != nil {
			return err
		}
	}

	if form.Language != "" {
		language, err := validate.Language(form.Language)
		if err != nil {
			return err
		}
		form.Language = language
	}

	// Check if the deprecated "federated" field was
	// set in lieu of "local_only", and use it if so.
	if form.LocalOnly == nil && form.Federated != nil { // nolint:staticcheck
		form.LocalOnly = util.Ptr(!*form.Federated) // nolint:staticcheck
	}

	return nil
}

func validateNormalizeCreatePoll(form *apimodel.StatusCreateRequest) error {
	maxPollOptions := config.GetStatusesPollMaxOptions()
	maxPollChars := config.GetStatusesPollOptionMaxChars()

	// Normalize poll expiry if necessary.
	// If we parsed this as JSON, expires_in
	// may be either a float64 or a string.
	if ei := form.Poll.ExpiresInI; ei != nil {
		switch e := ei.(type) {
		case float64:
			form.Poll.ExpiresIn = int(e)

		case string:
			expiresIn, err := strconv.Atoi(e)
			if err != nil {
				return fmt.Errorf("could not parse expires_in value %s as integer: %w", e, err)
			}

			form.Poll.ExpiresIn = expiresIn

		default:
			return fmt.Errorf("could not parse expires_in type %T as integer", ei)
		}
	}

	if len(form.Poll.Options) == 0 {
		return errors.New("poll with no options")
	}

	if len(form.Poll.Options) > maxPollOptions {
		return fmt.Errorf("too many poll options provided, %d provided but limit is %d", len(form.Poll.Options), maxPollOptions)
	}

	for _, p := range form.Poll.Options {
		if length := len([]rune(p)); length > maxPollChars {
			return fmt.Errorf("poll option too long, %d characters provided but limit is %d", length, maxPollChars)
		}
	}

	return nil
}
