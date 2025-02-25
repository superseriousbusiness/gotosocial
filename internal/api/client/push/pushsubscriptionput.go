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

package push

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// PushSubscriptionPUTHandler swagger:operation PUT /api/v1/push/subscription pushSubscriptionPut
//
// Update the Web Push subscription for the current access token.
// Only which notifications you receive can be updated.
//
//	---
//	tags:
//	- push
//
//	consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: data[alerts][follow]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when someone has followed you?
//	-
//		name: data[alerts][follow_request]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when someone has requested to follow you?
//	-
//		name: data[alerts][favourite]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a status you created has been favourited by someone else?
//	-
//		name: data[alerts][mention]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when someone else has mentioned you in a status?
//	-
//		name: data[alerts][reblog]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a status you created has been boosted by someone else?
//	-
//		name: data[alerts][poll]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a poll you voted in or created has ended?
//	-
//		name: data[alerts][status]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a subscribed account posts a status?
//	-
//		name: data[alerts][update]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a status you interacted with has been edited?
//	-
//		name: data[alerts][admin.sign_up]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a new user has signed up?
//	-
//		name: data[alerts][admin.report]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a new report has been filed?
//	-
//		name: data[alerts][pending.favourite]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a fave is pending?
//	-
//		name: data[alerts][pending.reply]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a reply is pending?
//	-
//		name: data[alerts][pending.reblog]
//		in: formData
//		type: boolean
//		default: false
//		description: Receive a push notification when a boost is pending?
//	-
//		name: data[policy]
//		in: formData
//		type: string
//		enum:
//			- all
//			- followed
//			- follower
//			- none
//		default: all
//		description: Which accounts to receive push notifications from.
//
//	security:
//	- OAuth2 Bearer:
//		- push
//
//	responses:
//		'200':
//			description: Web Push subscription for current access token.
//			schema:
//				"$ref": "#/definitions/webPushSubscription"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: This access token doesn't have an associated subscription.
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) PushSubscriptionPUTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopePush,
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

	form := &apimodel.WebPushSubscriptionUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateNormalizeUpdate(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiSubscription, errWithCode := m.processor.Push().Update(c, authed.Token.GetAccess(), form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiSubscription)
}

// validateNormalizeUpdate copies form fields to their canonical JSON equivalents
// and sets defaults for fields that have them.
func validateNormalizeUpdate(request *apimodel.WebPushSubscriptionUpdateRequest) error {
	if request.Data == nil {
		request.Data = &apimodel.WebPushSubscriptionRequestData{}
	}

	if request.Data.Alerts == nil {
		request.Data.Alerts = &apimodel.WebPushSubscriptionAlerts{}
	}

	if request.DataAlertsFollow != nil {
		request.Data.Alerts.Follow = *request.DataAlertsFollow
	}
	if request.DataAlertsFollowRequest != nil {
		request.Data.Alerts.FollowRequest = *request.DataAlertsFollowRequest
	}
	if request.DataAlertsMention != nil {
		request.Data.Alerts.Mention = *request.DataAlertsMention
	}
	if request.DataAlertsReblog != nil {
		request.Data.Alerts.Reblog = *request.DataAlertsReblog
	}
	if request.DataAlertsPoll != nil {
		request.Data.Alerts.Poll = *request.DataAlertsPoll
	}
	if request.DataAlertsStatus != nil {
		request.Data.Alerts.Status = *request.DataAlertsStatus
	}
	if request.DataAlertsUpdate != nil {
		request.Data.Alerts.Update = *request.DataAlertsUpdate
	}
	if request.DataAlertsAdminSignup != nil {
		request.Data.Alerts.AdminSignup = *request.DataAlertsAdminSignup
	}
	if request.DataAlertsAdminReport != nil {
		request.Data.Alerts.AdminReport = *request.DataAlertsAdminReport
	}
	if request.DataAlertsPendingFavourite != nil {
		request.Data.Alerts.PendingFavourite = *request.DataAlertsPendingFavourite
	}
	if request.DataAlertsPendingReply != nil {
		request.Data.Alerts.PendingReply = *request.DataAlertsPendingReply
	}
	if request.DataAlertsPendingReblog != nil {
		request.Data.Alerts.Reblog = *request.DataAlertsPendingReblog
	}

	if request.DataPolicy != nil {
		request.Data.Policy = request.DataPolicy
	}
	if request.Data.Policy == nil {
		request.Data.Policy = util.Ptr(apimodel.WebPushNotificationPolicyAll)
	}

	return nil
}
