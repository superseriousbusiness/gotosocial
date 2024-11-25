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

package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	/* API version keys */

	APIVersionKey = "api_version"
	APIv1         = "v1"
	APIv2         = "v2"

	/* Common keys */

	IDKey              = "id"
	LimitKey           = "limit"
	LocalKey           = "local"
	MaxIDKey           = "max_id"
	SinceIDKey         = "since_id"
	MinIDKey           = "min_id"
	UsernameKey        = "username"
	AccountIDKey       = "account_id"
	TargetAccountIDKey = "target_account_id"
	ResolvedKey        = "resolved"

	/* AP endpoint keys */

	OnlyOtherAccountsKey = "only_other_accounts"

	/* Search keys */

	SearchExcludeUnreviewedKey = "exclude_unreviewed"
	SearchFollowingKey         = "following"
	SearchLookupKey            = "acct"
	SearchOffsetKey            = "offset"
	SearchQueryKey             = "q"
	SearchResolveKey           = "resolve"
	SearchTypeKey              = "type"

	/* Tag keys */

	TagNameKey = "tag_name"

	/* Web endpoint keys */

	WebStatusIDKey = "status"

	/* Domain permission keys */

	DomainPermissionExportKey         = "export"
	DomainPermissionImportKey         = "import"
	DomainPermissionSubscriptionIDKey = "subscription_id"
	DomainPermissionPermTypeKey       = "permission_type"
	DomainPermissionDomainKey         = "domain"

	/* Admin query keys */

	AdminRemoteKey      = "remote"
	AdminActiveKey      = "active"
	AdminPendingKey     = "pending"
	AdminDisabledKey    = "disabled"
	AdminSilencedKey    = "silenced"
	AdminSuspendedKey   = "suspended"
	AdminSensitizedKey  = "sensitized"
	AdminDisplayNameKey = "display_name"
	AdminByDomainKey    = "by_domain"
	AdminEmailKey       = "email"
	AdminIPKey          = "ip"
	AdminStaffKey       = "staff"
	AdminOriginKey      = "origin"
	AdminStatusKey      = "status"
	AdminPermissionsKey = "permissions"
	AdminRoleIDsKey     = "role_ids[]"
	AdminInvitedByKey   = "invited_by"

	/* Interaction policy + request keys */

	InteractionStatusIDKey   = "status_id"
	InteractionFavouritesKey = "favourites"
	InteractionRepliesKey    = "replies"
	InteractionReblogsKey    = "reblogs"
)

/*
	Parse functions for *OPTIONAL* parameters with default values.
*/

func ParseMaxID(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func ParseSinceID(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func ParseMinID(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func ParseLimit(value string, defaultValue int, max, min int) (int, gtserror.WithCode) {
	i, err := parseInt(value, defaultValue, max, min, LimitKey)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func ParseLocal(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, LocalKey)
}

func ParseResolved(value string, defaultValue *bool) (*bool, gtserror.WithCode) {
	return parseBoolPtr(value, defaultValue, ResolvedKey)
}

func ParseSearchExcludeUnreviewed(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, SearchExcludeUnreviewedKey)
}

func ParseSearchFollowing(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, SearchFollowingKey)
}

func ParseSearchOffset(value string, defaultValue int, max, min int) (int, gtserror.WithCode) {
	return parseInt(value, defaultValue, max, min, SearchOffsetKey)
}

func ParseSearchResolve(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, SearchResolveKey)
}

func ParseDomainPermissionExport(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, DomainPermissionExportKey)
}

func ParseDomainPermissionImport(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, DomainPermissionImportKey)
}

func ParseOnlyOtherAccounts(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, OnlyOtherAccountsKey)
}

func ParseAdminRemote(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminRemoteKey)
}

func ParseAdminActive(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminActiveKey)
}

func ParseAdminPending(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminPendingKey)
}

func ParseAdminDisabled(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminDisabledKey)
}

func ParseAdminSilenced(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminSilencedKey)
}

func ParseAdminSuspended(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminSuspendedKey)
}

func ParseAdminStaff(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, AdminStaffKey)
}

func ParseInteractionFavourites(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, InteractionFavouritesKey)
}

func ParseInteractionReplies(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, InteractionRepliesKey)
}

func ParseInteractionReblogs(value string, defaultValue bool) (bool, gtserror.WithCode) {
	return parseBool(value, defaultValue, InteractionReblogsKey)
}

func ParseNotificationType(value string) (gtsmodel.NotificationType, gtserror.WithCode) {
	switch strings.ToLower(value) {
	case "follow":
		return gtsmodel.NotificationFollow, nil
	case "follow_request":
		return gtsmodel.NotificationFollowRequest, nil
	case "mention":
		return gtsmodel.NotificationMention, nil
	case "reblog":
		return gtsmodel.NotificationReblog, nil
	case "favourite":
		return gtsmodel.NotificationFave, nil
	case "poll":
		return gtsmodel.NotificationPoll, nil
	case "status":
		return gtsmodel.NotificationStatus, nil
	case "admin.sign_up":
		return gtsmodel.NotificationSignup, nil
	case "pending.favourite":
		return gtsmodel.NotificationPendingFave, nil
	case "pending.reply":
		return gtsmodel.NotificationPendingReply, nil
	case "pending.reblog":
		return gtsmodel.NotificationPendingReblog, nil
	default:
		text := fmt.Sprintf("unrecognized notification type %s", value)
		return 0, gtserror.NewErrorBadRequest(errors.New(text), text)
	}
}

func ParseNotificationTypes(values []string) ([]gtsmodel.NotificationType, gtserror.WithCode) {
	if len(values) == 0 {
		return nil, nil
	}
	ntypes := make([]gtsmodel.NotificationType, len(values))
	for i, value := range values {
		ntype, errWithCode := ParseNotificationType(value)
		if errWithCode != nil {
			return nil, errWithCode
		}
		ntypes[i] = ntype
	}
	return ntypes, nil
}

/*
	Parse functions for *REQUIRED* parameters.
*/

func ParseAPIVersion(value string, availableVersion ...string) (string, gtserror.WithCode) {
	key := APIVersionKey

	if value == "" {
		return "", requiredError(key)
	}

	for _, av := range availableVersion {
		if value == av {
			return value, nil
		}
	}

	err := fmt.Errorf(
		"invalid API version, valid versions for this path are [%s]",
		strings.Join(availableVersion, ", "),
	)
	return "", gtserror.NewErrorBadRequest(err, err.Error())
}

func ParseID(value string) (string, gtserror.WithCode) {
	key := IDKey

	if value == "" {
		return "", requiredError(key)
	}

	return value, nil
}

func ParseSearchLookup(value string) (string, gtserror.WithCode) {
	key := SearchLookupKey

	if value == "" {
		return "", requiredError(key)
	}

	return value, nil
}

func ParseSearchQuery(value string) (string, gtserror.WithCode) {
	key := SearchQueryKey

	if value == "" {
		return "", requiredError(key)
	}

	return value, nil
}

func ParseTagName(value string) (string, gtserror.WithCode) {
	key := TagNameKey

	if value == "" {
		return "", requiredError(key)
	}

	return value, nil
}

func ParseUsername(value string) (string, gtserror.WithCode) {
	key := UsernameKey

	if value == "" {
		return "", requiredError(key)
	}

	return value, nil
}

func ParseWebStatusID(value string) (string, gtserror.WithCode) {
	key := WebStatusIDKey

	if value == "" {
		return "", requiredError(key)
	}

	return value, nil
}

/*
	Internal functions
*/

func parseBool(value string, defaultValue bool, key string) (bool, gtserror.WithCode) {
	if value == "" {
		return defaultValue, nil
	}

	i, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue, parseError(key, value, defaultValue, err)
	}

	return i, nil
}

func parseBoolPtr(value string, defaultValue *bool, key string) (*bool, gtserror.WithCode) {
	if value == "" {
		return defaultValue, nil
	}

	i, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue, parseError(key, value, defaultValue, err)
	}

	return &i, nil
}

func parseInt(value string, defaultValue int, max int, min int, key string) (int, gtserror.WithCode) {
	if value == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue, parseError(key, value, defaultValue, err)
	}

	if i > max {
		i = max
	} else if i < min {
		i = min
	}

	return i, nil
}

// parseError returns gtserror.WithCode set to 400 Bad Request, to indicate
// to the caller that a key was set to a value that could not be parsed.
func parseError(key string, value, defaultValue any, err error) gtserror.WithCode {
	err = fmt.Errorf("error parsing key %s with value %s as %T: %w", key, value, defaultValue, err)
	return gtserror.NewErrorBadRequest(err, err.Error())
}

// requiredError returns gtserror.WithCode set to 400 Bad Request, to indicate
// to the caller a required key value was not provided, or was empty.
func requiredError(key string) gtserror.WithCode {
	err := fmt.Errorf("required key %s was not set or had empty value", key)
	return gtserror.NewErrorBadRequest(err, err.Error())
}
