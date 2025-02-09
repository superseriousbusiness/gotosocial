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
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	webpushgo "github.com/SherClockHolmes/webpush-go"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// realSender is the production Web Push sender,
// backed by an HTTP client, DB, and worker pool.
type realSender struct {
	httpClient *http.Client
	state      *state.State
	converter  *typeutils.Converter
}

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

	// Subscriptions we're actually going to send to.
	relevantSubscriptions := slices.DeleteFunc(
		subscriptions,
		func(subscription *gtsmodel.WebPushSubscription) bool {
			return r.shouldSkipSubscription(ctx, notification, subscription)
		},
	)
	if len(relevantSubscriptions) == 0 {
		return nil
	}

	// Get VAPID keys.
	vapidKeyPair, err := r.state.DB.GetVAPIDKeyPair(ctx)
	if err != nil {
		return gtserror.Newf("error getting VAPID key pair: %w", err)
	}

	// Get target account settings.
	targetAccountSettings, err := r.state.DB.GetAccountSettings(ctx, notification.TargetAccountID)
	if err != nil {
		return gtserror.Newf("error getting settings for account %s: %w", notification.TargetAccountID, err)
	}

	// Get API representations of notification and accounts involved.
	apiNotification, err := r.converter.NotificationToAPINotification(ctx, notification, filters, mutes)
	if err != nil {
		return gtserror.Newf("error converting notification %s to API representation: %w", notification.ID, err)
	}

	// Queue up a .Send() call for each relevant subscription.
	for _, subscription := range relevantSubscriptions {
		r.state.Workers.WebPush.Queue.Push(func(ctx context.Context) {
			if err := r.sendToSubscription(
				ctx,
				vapidKeyPair,
				targetAccountSettings,
				subscription,
				notification,
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

// shouldSkipSubscription returns true if this subscription is not relevant to this notification.
func (r *realSender) shouldSkipSubscription(
	ctx context.Context,
	notification *gtsmodel.Notification,
	subscription *gtsmodel.WebPushSubscription,
) bool {
	// Remove subscriptions that don't want this type of notification.
	if !subscription.NotificationFlags.Get(notification.NotificationType) {
		return true
	}

	// Check against subscription's notification policy.
	switch subscription.Policy {
	case gtsmodel.WebPushNotificationPolicyAll:
		// Allow notifications from any account.
		return false

	case gtsmodel.WebPushNotificationPolicyFollowed:
		// Allow if the subscription account follows the notifying account.
		isFollowing, err := r.state.DB.IsFollowing(ctx, subscription.AccountID, notification.OriginAccountID)
		if err != nil {
			log.Errorf(
				ctx,
				"error checking whether account %s follows account %s: %v",
				subscription.AccountID,
				notification.OriginAccountID,
				err,
			)
			return true
		}
		return !isFollowing

	case gtsmodel.WebPushNotificationPolicyFollower:
		// Allow if the notifying account follows the subscription account.
		isFollowing, err := r.state.DB.IsFollowing(ctx, notification.OriginAccountID, subscription.AccountID)
		if err != nil {
			log.Errorf(
				ctx,
				"error checking whether account %s follows account %s: %v",
				notification.OriginAccountID,
				subscription.AccountID,
				err,
			)
			return true
		}
		return !isFollowing

	case gtsmodel.WebPushNotificationPolicyNone:
		// This subscription doesn't want any push notifications.
		return true

	default:
		log.Errorf(
			ctx,
			"unknown Web Push notification policy for subscription with token ID %s: %d",
			subscription.TokenID,
			subscription.Policy,
		)
		return true
	}
}

// sendToSubscription sends a notification to a single Web Push subscription.
func (r *realSender) sendToSubscription(
	ctx context.Context,
	vapidKeyPair *gtsmodel.VAPIDKeyPair,
	targetAccountSettings *gtsmodel.AccountSettings,
	subscription *gtsmodel.WebPushSubscription,
	notification *gtsmodel.Notification,
	apiNotification *apimodel.Notification,
) error {
	const (
		// TTL is an arbitrary time to ask the Web Push server to store notifications
		// while waiting for the client to retrieve them.
		TTL = 48 * time.Hour

		// recordSize limits how big our notifications can be once padding is applied.
		// To be polite to applications that need to relay them over services like APNS,
		// which has a max message size of 4 kB, we set this comfortably smaller.
		recordSize = 2048

		// responseBodyMaxLen limits how much of the Web Push server response we read for error messages.
		responseBodyMaxLen = 1024
	)

	// Get the associated access token.
	token, err := r.state.DB.GetTokenByID(ctx, subscription.TokenID)
	if err != nil {
		return gtserror.Newf("error getting token %s: %w", subscription.TokenID, err)
	}

	// Create push notification payload struct.
	pushNotification := &apimodel.WebPushNotification{
		NotificationID:   apiNotification.ID,
		NotificationType: apiNotification.Type,
		Title:            formatNotificationTitle(ctx, subscription, notification, apiNotification),
		Body:             formatNotificationBody(apiNotification),
		Icon:             apiNotification.Account.Avatar,
		PreferredLocale:  targetAccountSettings.Language,
		AccessToken:      token.Access,
	}

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
			RecordSize:      recordSize,
			Subscriber:      "https://" + config.GetHost(),
			VAPIDPublicKey:  vapidKeyPair.Public,
			VAPIDPrivateKey: vapidKeyPair.Private,
			TTL:             int(TTL.Seconds()),
		},
	)
	if err != nil {
		return gtserror.Newf("error sending Web Push notification: %w", err)
	}
	defer resp.Body.Close()

	switch {
	// All good, delivered.
	case resp.StatusCode >= 200 && resp.StatusCode <= 299:
		return nil

	// Temporary outage or some other delivery issue.
	case resp.StatusCode == http.StatusRequestTimeout ||
		resp.StatusCode == http.StatusRequestEntityTooLarge ||
		resp.StatusCode == http.StatusTooManyRequests ||
		(resp.StatusCode >= 500 && resp.StatusCode <= 599):

		// Try to get the response body.
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, responseBodyMaxLen))
		if err != nil {
			return gtserror.Newf("error reading Web Push server response: %w", err)
		}

		// Return the error with its response body.
		return gtserror.Newf(
			"unexpected HTTP status %s received when sending Web Push notification: %s",
			resp.Status,
			string(bodyBytes),
		)

	// Some serious error that indicates auth problems, not a Web Push server, etc.
	// We should not send any more notifications to this subscription. Try to delete it.
	default:
		err := r.state.DB.DeleteWebPushSubscriptionByTokenID(ctx, subscription.TokenID)
		if err != nil {
			return gtserror.Newf(
				"received HTTP status %s but failed to delete subscription: %s",
				resp.Status,
				err,
			)
		}

		log.Infof(
			ctx,
			"Deleted Web Push subscription with token ID %s because push server sent HTTP status %s",
			subscription.TokenID, resp.Status,
		)
		return nil
	}
}

// formatNotificationTitle creates a title for a Web Push notification from the notification type and account's name.
func formatNotificationTitle(
	ctx context.Context,
	subscription *gtsmodel.WebPushSubscription,
	notification *gtsmodel.Notification,
	apiNotification *apimodel.Notification,
) string {
	displayNameOrAcct := apiNotification.Account.DisplayName
	if displayNameOrAcct == "" {
		displayNameOrAcct = apiNotification.Account.Acct
	}

	switch notification.NotificationType {
	case gtsmodel.NotificationFollow:
		return fmt.Sprintf("%s followed you", displayNameOrAcct)
	case gtsmodel.NotificationFollowRequest:
		return fmt.Sprintf("%s requested to follow you", displayNameOrAcct)
	case gtsmodel.NotificationMention:
		return fmt.Sprintf("%s mentioned you", displayNameOrAcct)
	case gtsmodel.NotificationReblog:
		return fmt.Sprintf("%s boosted your post", displayNameOrAcct)
	case gtsmodel.NotificationFavourite:
		return fmt.Sprintf("%s faved your post", displayNameOrAcct)
	case gtsmodel.NotificationPoll:
		if subscription.AccountID == notification.TargetAccountID {
			return "Your poll has ended"
		} else {
			return fmt.Sprintf("%s's poll has ended", displayNameOrAcct)
		}
	case gtsmodel.NotificationStatus:
		return fmt.Sprintf("%s posted", displayNameOrAcct)
	case gtsmodel.NotificationAdminSignup:
		return fmt.Sprintf("%s requested to sign up", displayNameOrAcct)
	case gtsmodel.NotificationPendingFave:
		return fmt.Sprintf("%s faved your post, which requires your approval", displayNameOrAcct)
	case gtsmodel.NotificationPendingReply:
		return fmt.Sprintf("%s mentioned you, which requires your approval", displayNameOrAcct)
	case gtsmodel.NotificationPendingReblog:
		return fmt.Sprintf("%s boosted your post, which requires your approval", displayNameOrAcct)
	case gtsmodel.NotificationAdminReport:
		return fmt.Sprintf("%s submitted a report", displayNameOrAcct)
	case gtsmodel.NotificationUpdate:
		return fmt.Sprintf("%s updated their post", displayNameOrAcct)
	default:
		log.Warnf(ctx, "Unknown notification type: %d", notification.NotificationType)
		return fmt.Sprintf(
			"%s did something (unknown notification type %d)",
			displayNameOrAcct,
			notification.NotificationType,
		)
	}
}

// formatNotificationBody creates a body for a Web Push notification,
// from the CW or beginning of the body text of the status, if there is one,
// or the beginning of the bio text of the related account.
func formatNotificationBody(apiNotification *apimodel.Notification) string {
	// bodyMaxLen is a polite maximum length for a Web Push notification's body text, in bytes. Note that this isn't
	// limited per se, but Web Push servers may reject anything with a total request body size over 4k,
	// and we set a lower max size above for compatibility with mobile push systems.
	const bodyMaxLen = 1500

	var body string
	if apiNotification.Status != nil {
		if apiNotification.Status.SpoilerText != "" {
			body = apiNotification.Status.SpoilerText
		} else {
			body = text.SanitizeToPlaintext(apiNotification.Status.Content)
		}
	} else {
		body = text.SanitizeToPlaintext(apiNotification.Account.Note)
	}
	return firstNBytesTrimSpace(body, bodyMaxLen)
}

// firstNBytesTrimSpace returns the first N bytes of a string, trimming leading and trailing whitespace.
func firstNBytesTrimSpace(s string, n int) string {
	return strings.TrimSpace(text.FirstNBytesByWords(strings.TrimSpace(s), n))
}
