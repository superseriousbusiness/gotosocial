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
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// PushSubscriptionPOSTHandler swagger:operation POST /api/v1/push/subscription pushSubscriptionPost
//
// Create a new Web Push subscription for the current access token, or replace the existing one.
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
//		name: subscription[endpoint]
//		in: formData
//		type: string
//		required: true
//		minLength: 1
//		description: The URL to which Web Push notifications will be sent.
//	-
//		name: subscription[keys][auth]
//		in: formData
//		type: string
//		required: true
//		minLength: 1
//		description: The auth secret, a Base64 encoded string of 16 bytes of random data.
//	-
//		name: subscription[keys][p256dh]
//		in: formData
//		type: string
//		required: true
//		minLength: 1
//		description: The user agent public key, a Base64 encoded string of a public key from an ECDH keypair using the prime256v1 curve.
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
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) PushSubscriptionPOSTHandler(c *gin.Context) {
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

	form := &apimodel.WebPushSubscriptionCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := validateNormalizeCreate(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiSubscription, errWithCode := m.processor.Push().CreateOrReplace(c, authed.Account.ID, authed.Token.GetAccess(), form)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiSubscription)
}

// validateNormalizeCreate checks subscription endpoint format and keys decodability,
// and copies form fields to their canonical JSON equivalents.
func validateNormalizeCreate(request *apimodel.WebPushSubscriptionCreateRequest) error {
	if request.Subscription == nil {
		request.Subscription = &apimodel.WebPushSubscriptionRequestSubscription{}
	}

	// Normalize and validate endpoint URL.
	if request.SubscriptionEndpoint != nil {
		request.Subscription.Endpoint = *request.SubscriptionEndpoint
	}

	if request.Subscription.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	endpointURL, err := url.Parse(request.Subscription.Endpoint)
	if err != nil {
		return errors.New("endpoint must be a valid URL")
	}
	if endpointURL.Scheme != "https" {
		return errors.New("endpoint must be an https:// URL")
	}
	if endpointURL.Host == "" {
		return errors.New("endpoint URL must have a host")
	}
	if endpointURL.Fragment != "" {
		return errors.New("endpoint URL must not have a fragment")
	}

	// Normalize and validate auth secret.
	if request.SubscriptionKeysAuth != nil {
		request.Subscription.Keys.Auth = *request.SubscriptionKeysAuth
	}

	authBytes, err := base64DecodeAny("auth", request.Subscription.Keys.Auth)
	if err != nil {
		return err
	}
	if len(authBytes) != 16 {
		return fmt.Errorf("auth must be 16 bytes long, got %d", len(authBytes))
	}

	// Normalize and validate public key.
	if request.SubscriptionKeysP256dh != nil {
		request.Subscription.Keys.P256dh = *request.SubscriptionKeysP256dh
	}

	p256dhBytes, err := base64DecodeAny("p256dh", request.Subscription.Keys.P256dh)
	if err != nil {
		return err
	}
	_, err = ecdh.P256().NewPublicKey(p256dhBytes)
	if err != nil {
		return fmt.Errorf("p256dh must be a valid public key on the NIST P-256 curve: %w", err)
	}

	return validateNormalizeUpdate(&request.WebPushSubscriptionUpdateRequest)
}

// base64DecodeAny tries decoding a string with standard and URL alphabets of Base64, with and without padding.
func base64DecodeAny(name string, value string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	}

	for _, encoding := range encodings {
		if bytes, err := encoding.DecodeString(value); err == nil {
			return bytes, nil
		}
	}

	return nil, fmt.Errorf("%s is not valid Base64 data", name)
}
