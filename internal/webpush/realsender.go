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

package webpush

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	webpushgo "github.com/SherClockHolmes/webpush-go"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// realSender is the production Web Push sender, backed by an HTTP client, DB, and worker pool.
type realSender struct {
	httpClient *http.Client
	state      *state.State
	tc         *typeutils.Converter
}

// NewRealSender creates a Sender from an http.Client instead of an httpclient.Client.
// This should only be used by NewSender and in tests.
func NewRealSender(httpClient *http.Client, state *state.State) Sender {
	return &realSender{
		httpClient: httpClient,
		state:      state,
		tc:         typeutils.NewConverter(state),
	}
}

// TTL is an arbitrary time to ask the Web Push server to store notifications
// while waiting for the client to retrieve them.
const TTL = 48 * time.Hour

func (r *realSender) Send(
	ctx context.Context,
	notification *gtsmodel.Notification,
	filters []*gtsmodel.Filter,
	mutes *usermute.CompiledUserMuteList,
) error {
	// Load subscriptions.
	subscriptions, err := r.state.DB.GetWebPushSubscriptionsByAccountID(ctx, notification.TargetAccountID)
	if err != nil {
		return gtserror.Newf(
			"error getting Web Push subscriptions for account %s: %w",
			notification.TargetAccountID,
			err,
		)
	}
	if len(subscriptions) == 0 {
		return nil
	}

	// Subscriptions we're actually going to send to.
	relevantSubscriptions := make([]*gtsmodel.WebPushSubscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		// Check whether this subscription wants this type of notification.
		notify := false
		switch notification.NotificationType {
		case gtsmodel.NotificationFollow:
			notify = *subscription.NotifyFollow
		case gtsmodel.NotificationFollowRequest:
			notify = *subscription.NotifyFollowRequest
		case gtsmodel.NotificationMention:
			notify = *subscription.NotifyMention
		case gtsmodel.NotificationReblog:
			notify = *subscription.NotifyReblog
		case gtsmodel.NotificationFavourite:
			notify = *subscription.NotifyFavourite
		case gtsmodel.NotificationPoll:
			notify = *subscription.NotifyPoll
		case gtsmodel.NotificationStatus:
			notify = *subscription.NotifyStatus
		case gtsmodel.NotificationAdminSignup:
			notify = *subscription.NotifyAdminSignup
		case gtsmodel.NotificationAdminReport:
			notify = *subscription.NotifyAdminReport
		case gtsmodel.NotificationPendingFave:
			notify = *subscription.NotifyPendingFavourite
		case gtsmodel.NotificationPendingReply:
			notify = *subscription.NotifyPendingReply
		case gtsmodel.NotificationPendingReblog:
			notify = *subscription.NotifyPendingReblog
		default:
			log.Errorf(
				ctx,
				"notification type not supported by Web Push subscriptions: %v",
				notification.NotificationType,
			)
			continue
		}
		if !notify {
			continue
		}
		relevantSubscriptions = append(relevantSubscriptions, subscription)
	}
	if len(relevantSubscriptions) == 0 {
		return nil
	}

	// Load VAPID keys into webpush-go options struct.
	vapidKeyPair, err := r.state.DB.GetVAPIDKeyPair(ctx)
	if err != nil {
		return gtserror.Newf("error getting VAPID key pair: %w", err)
	}

	// Get API representations of notification and accounts involved.
	// This also loads the target account's settings.
	apiNotification, err := r.tc.NotificationToAPINotification(ctx, notification, filters, mutes)
	if err != nil {
		return gtserror.Newf("error converting notification %s to API representation: %w", notification.ID, err)
	}

	// Queue up a .Send() call for each relevant subscription.
	for _, subscription := range relevantSubscriptions {
		r.state.Workers.WebPush.Queue.Push(func(ctx context.Context) {
			if err := r.sendToSubscription(
				ctx,
				vapidKeyPair,
				subscription,
				notification.TargetAccount,
				apiNotification,
			); err != nil {
				log.Errorf(
					ctx,
					"error sending Web Push notification for subscription with token ID %s: %v",
					subscription.TokenID,
					err,
				)
			}
		})
	}

	return nil
}

// sendToSubscription sends a notification to a single Web Push subscription.
func (r *realSender) sendToSubscription(
	ctx context.Context,
	vapidKeyPair *gtsmodel.VAPIDKeyPair,
	subscription *gtsmodel.WebPushSubscription,
	targetAccount *gtsmodel.Account,
	apiNotification *apimodel.Notification,
) error {
	// Get the associated access token.
	token, err := r.state.DB.GetTokenByID(ctx, subscription.TokenID)
	if err != nil {
		return gtserror.Newf("error getting token %s: %w", subscription.TokenID, err)
	}

	// Create push notification payload struct.
	pushNotification := &apimodel.WebPushNotification{
		NotificationID:   apiNotification.ID,
		NotificationType: apiNotification.Type,
		Icon:             apiNotification.Account.Avatar,
		PreferredLocale:  targetAccount.Settings.Language,
		AccessToken:      token.Access,
	}

	// Set the notification title.
	displayNameOrAcct := apiNotification.Account.DisplayName
	if displayNameOrAcct == "" {
		displayNameOrAcct = apiNotification.Account.Acct
	}
	// TODO: (Vyr) improve copy
	pushNotification.Title = fmt.Sprintf("%s from %s", apiNotification.Type, displayNameOrAcct)

	// Set the notification body.
	if apiNotification.Status != nil {
		if apiNotification.Status.SpoilerText != "" {
			pushNotification.Body = apiNotification.Status.SpoilerText
		} else {
			pushNotification.Body = text.SanitizeToPlaintext(apiNotification.Status.Content)
		}
	} else {
		pushNotification.Body = text.SanitizeToPlaintext(apiNotification.Account.Note)
	}
	// TODO: (Vyr) trim this

	// Encode the push notification as JSON.
	pushNotificationBytes, err := json.Marshal(pushNotification)
	if err != nil {
		return gtserror.Newf("error encoding Web Push notification: %w", err)
	}

	// Send push notification.
	resp, err := webpushgo.SendNotificationWithContext(
		ctx,
		pushNotificationBytes,
		&webpushgo.Subscription{
			Endpoint: subscription.Endpoint,
			Keys: webpushgo.Keys{
				Auth:   subscription.Auth,
				P256dh: subscription.P256dh,
			},
		},
		&webpushgo.Options{
			HTTPClient:      r.httpClient,
			VAPIDPublicKey:  vapidKeyPair.Public,
			VAPIDPrivateKey: vapidKeyPair.Private,
			TTL:             int(TTL.Seconds()),
		},
	)
	if err != nil {
		return gtserror.Newf("error sending Web Push notification: %w", err)
	}
	// We're not going to use the response body, but we need to close it so we don't leak the connection.
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return gtserror.Newf("unexpected HTTP status received when sending Web Push notification: %s", resp.Status)
	}

	return nil
}

// gtsHttpClientRoundTripper helps wrap a GtS HTTP client back into a regular HTTP client,
// so that webpush-go can use our IP filters, bad hosts list, and retries.
type gtsHttpClientRoundTripper struct {
	httpClient *httpclient.Client
}

func (r *gtsHttpClientRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return r.httpClient.Do(request)
}
