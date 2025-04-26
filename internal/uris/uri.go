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

package uris

import (
	"fmt"
	"net/url"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
)

const (
	UsersPath        = "users"         // UsersPath is for serving users info
	StatusesPath     = "statuses"      // StatusesPath is for serving statuses
	InboxPath        = "inbox"         // InboxPath represents the activitypub inbox location
	OutboxPath       = "outbox"        // OutboxPath represents the activitypub outbox location
	FollowersPath    = "followers"     // FollowersPath represents the activitypub followers location
	FollowingPath    = "following"     // FollowingPath represents the activitypub following location
	LikedPath        = "liked"         // LikedPath represents the activitypub liked location
	CollectionsPath  = "collections"   // CollectionsPath represents the activitypub collections location
	FeaturedPath     = "featured"      // FeaturedPath represents the activitypub featured location
	PublicKeyPath    = "main-key"      // PublicKeyPath is for serving an account's public key
	FollowPath       = "follow"        // FollowPath used to generate the URI for an individual follow or follow request
	UpdatePath       = "updates"       // UpdatePath is used to generate the URI for an account update
	BlocksPath       = "blocks"        // BlocksPath is used to generate the URI for a block
	MovesPath        = "moves"         // MovesPath is used to generate the URI for a move
	ReportsPath      = "reports"       // ReportsPath is used to generate the URI for a report/flag
	ConfirmEmailPath = "confirm_email" // ConfirmEmailPath is used to generate the URI for an email confirmation link
	FileserverPath   = "fileserver"    // FileserverPath is a path component for serving attachments + media
	EmojiPath        = "emoji"         // EmojiPath represents the activitypub emoji location
	TagsPath         = "tags"          // TagsPath represents the activitypub tags location
	AcceptsPath      = "accepts"       // AcceptsPath represents the activitypub Accept's location
	RejectsPath      = "rejects"       // RejectsPath represents the activitypub Reject's location
)

// UserURIs contains a bunch of UserURIs and URLs for a user, host, account, etc.
type UserURIs struct {
	// The web URL of the instance host, eg https://example.org
	HostURL string
	// The web URL of the user, eg., https://example.org/@example_user
	UserURL string
	// The web URL for statuses of this user, eg., https://example.org/@example_user/statuses
	StatusesURL string

	// The activitypub URI of this user, eg., https://example.org/users/example_user
	UserURI string
	// The activitypub URI for this user's statuses, eg., https://example.org/users/example_user/statuses
	StatusesURI string
	// The activitypub URI for this user's activitypub inbox, eg., https://example.org/users/example_user/inbox
	InboxURI string
	// The activitypub URI for this user's activitypub outbox, eg., https://example.org/users/example_user/outbox
	OutboxURI string
	// The activitypub URI for this user's followers, eg., https://example.org/users/example_user/followers
	FollowersURI string
	// The activitypub URI for this user's following, eg., https://example.org/users/example_user/following
	FollowingURI string
	// The activitypub URI for this user's liked posts eg., https://example.org/users/example_user/liked
	LikedURI string
	// The activitypub URI for this user's featured collections, eg., https://example.org/users/example_user/collections/featured
	FeaturedCollectionURI string
	// The URI for this user's public key, eg., https://example.org/users/example_user/publickey
	PublicKeyURI string
}

// GenerateURIForFollow returns the AP URI for a new follow -- something like:
// https://example.org/users/whatever_user/follow/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForFollow(username string, thisFollowID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, FollowPath, thisFollowID)
}

// GenerateURIForLike returns the AP URI for a new like/fave -- something like:
// https://example.org/users/whatever_user/liked/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForLike(username string, thisFavedID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, LikedPath, thisFavedID)
}

// GenerateURIForUpdate returns the AP URI for a new update activity -- something like:
// https://example.org/users/whatever_user#updates/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForUpdate(username string, thisUpdateID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s#%s/%s", protocol, host, UsersPath, username, UpdatePath, thisUpdateID)
}

// GenerateURIForBlock returns the AP URI for a new block activity -- something like:
// https://example.org/users/whatever_user/blocks/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForBlock(username string, thisBlockID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, BlocksPath, thisBlockID)
}

// GenerateURIForMove returns the AP URI for a new Move activity -- something like:
// https://example.org/users/whatever_user/moves/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForMove(username string, thisMoveID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, MovesPath, thisMoveID)
}

// GenerateURIForReport returns the API URI for a new Flag activity -- something like:
// https://example.org/reports/01GP3AWY4CRDVRNZKW0TEAMB5R
//
// This path specifically doesn't contain any info about the user who did the reporting,
// to protect their privacy.
func GenerateURIForReport(thisReportID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s", protocol, host, ReportsPath, thisReportID)
}

// GenerateURIForEmailConfirm returns a link for email confirmation -- something like:
// https://example.org/confirm_email?token=490e337c-0162-454f-ac48-4b22bb92a205
func GenerateURIForEmailConfirm(token string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s?token=%s", protocol, host, ConfirmEmailPath, token)
}

// GenerateURIForAccept returns the AP URI for a new Accept activity -- something like:
// https://example.org/users/whatever_user/accepts/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForAccept(username string, thisAcceptID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, AcceptsPath, thisAcceptID)
}

// GenerateURIForReject returns the AP URI for a new Reject activity -- something like:
// https://example.org/users/whatever_user/rejects/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForReject(username string, thisRejectID string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, RejectsPath, thisRejectID)
}

// GenerateURIsForAccount throws together a bunch of URIs for the given username, with the given protocol and host.
func GenerateURIsForAccount(username string) *UserURIs {
	protocol := config.GetProtocol()
	host := config.GetHost()

	// The below URLs are used for serving web requests
	hostURL := fmt.Sprintf("%s://%s", protocol, host)
	userURL := fmt.Sprintf("%s/@%s", hostURL, username)
	statusesURL := fmt.Sprintf("%s/%s", userURL, StatusesPath)

	// the below URIs are used in ActivityPub and Webfinger
	userURI := fmt.Sprintf("%s/%s/%s", hostURL, UsersPath, username)
	statusesURI := fmt.Sprintf("%s/%s", userURI, StatusesPath)
	inboxURI := fmt.Sprintf("%s/%s", userURI, InboxPath)
	outboxURI := fmt.Sprintf("%s/%s", userURI, OutboxPath)
	followersURI := fmt.Sprintf("%s/%s", userURI, FollowersPath)
	followingURI := fmt.Sprintf("%s/%s", userURI, FollowingPath)
	likedURI := fmt.Sprintf("%s/%s", userURI, LikedPath)
	collectionURI := fmt.Sprintf("%s/%s/%s", userURI, CollectionsPath, FeaturedPath)
	publicKeyURI := fmt.Sprintf("%s/%s", userURI, PublicKeyPath)

	return &UserURIs{
		HostURL:     hostURL,
		UserURL:     userURL,
		StatusesURL: statusesURL,

		UserURI:               userURI,
		StatusesURI:           statusesURI,
		InboxURI:              inboxURI,
		OutboxURI:             outboxURI,
		FollowersURI:          followersURI,
		FollowingURI:          followingURI,
		LikedURI:              likedURI,
		FeaturedCollectionURI: collectionURI,
		PublicKeyURI:          publicKeyURI,
	}
}

// URIForAttachment generates a URI for
// an attachment/emoji/header etc.
//
// Will produce something like:
//
//	"https://example.org/fileserver/01FPST95B8FC3HG3AGCDKPQNQ2/attachment/original/01FPST9QK4V5XWS3F9Z4F2G1X7.gif"
func URIForAttachment(
	accountID string,
	mediaType string,
	mediaSize string,
	mediaID string,
	extension string,
) string {
	const format = "%s://%s/%s/%s/%s/%s/%s.%s"

	return fmt.Sprintf(
		format,
		config.GetProtocol(),
		config.GetHost(),
		FileserverPath,
		accountID,
		mediaType,
		mediaSize,
		mediaID,
		extension,
	)
}

// StoragePathForAttachment generates a storage
// path for an attachment/emoji/header etc.
//
// Will produce something like:
//
//	"01FPST95B8FC3HG3AGCDKPQNQ2/attachment/original/01FPST9QK4V5XWS3F9Z4F2G1X7.gif"
func StoragePathForAttachment(
	accountID string,
	mediaType string,
	mediaSize string,
	mediaID string,
	extension string,
) string {
	const format = "%s/%s/%s/%s.%s"

	return fmt.Sprintf(
		format,
		accountID,
		mediaType,
		mediaSize,
		mediaID,
		extension,
	)
}

// URIForEmoji generates an
// ActivityPub URI for an emoji.
//
// Will produce something like:
//
//	"https://example.org/emoji/01FPST9QK4V5XWS3F9Z4F2G1X7"
func URIForEmoji(emojiID string) string {
	const format = "%s://%s/%s/%s"

	return fmt.Sprintf(
		format,
		config.GetProtocol(),
		config.GetHost(),
		EmojiPath,
		emojiID,
	)
}

// URIForTag generates an activitypub uri for a tag.
func URIForTag(name string) string {
	protocol := config.GetProtocol()
	host := config.GetHost()
	return fmt.Sprintf("%s://%s/%s/%s", protocol, host, TagsPath, strings.ToLower(name))
}

// IsUserPath returns true if the given URL path corresponds to eg /users/example_username
func IsUserPath(id *url.URL) bool {
	return regexes.UserPath.MatchString(id.Path)
}

// IsUserWebPath returns true if the given URL path corresponds to eg /@example_username
func IsUserWebPath(id *url.URL) bool {
	return regexes.UserWebPath.MatchString(id.Path)
}

// IsInboxPath returns true if the given URL path corresponds to eg /users/example_username/inbox
func IsInboxPath(id *url.URL) bool {
	return regexes.InboxPath.MatchString(id.Path)
}

// IsOutboxPath returns true if the given URL path corresponds to eg /users/example_username/outbox
func IsOutboxPath(id *url.URL) bool {
	return regexes.OutboxPath.MatchString(id.Path)
}

// IsFollowersPath returns true if the given URL path corresponds to eg /users/example_username/followers
func IsFollowersPath(id *url.URL) bool {
	return regexes.FollowersPath.MatchString(id.Path)
}

// IsFollowingPath returns true if the given URL path corresponds to eg /users/example_username/following
func IsFollowingPath(id *url.URL) bool {
	return regexes.FollowingPath.MatchString(id.Path)
}

// IsFollowPath returns true if the given URL path corresponds to eg /users/example_username/follow/SOME_ULID_OF_A_FOLLOW
func IsFollowPath(id *url.URL) bool {
	return regexes.FollowPath.MatchString(id.Path)
}

// IsLikedPath returns true if the given URL path corresponds to eg /users/example_username/liked
func IsLikedPath(id *url.URL) bool {
	return regexes.LikedPath.MatchString(id.Path)
}

// IsLikePath returns true if the given URL path corresponds to eg /users/example_username/liked/SOME_ULID_OF_A_STATUS
func IsLikePath(id *url.URL) bool {
	return regexes.LikePath.MatchString(id.Path)
}

// IsStatusesPath returns true if the given URL path corresponds to eg /users/example_username/statuses/SOME_ULID_OF_A_STATUS
func IsStatusesPath(id *url.URL) bool {
	return regexes.StatusesPath.MatchString(id.Path)
}

// IsPublicKeyPath returns true if the given URL path corresponds to eg /users/example_username/main-key
func IsPublicKeyPath(id *url.URL) bool {
	return regexes.PublicKeyPath.MatchString(id.Path)
}

// IsBlockPath returns true if the given URL path corresponds to eg /users/example_username/blocks/SOME_ULID_OF_A_BLOCK
func IsBlockPath(id *url.URL) bool {
	return regexes.BlockPath.MatchString(id.Path)
}

// IsReportPath returns true if the given URL path corresponds to eg /reports/SOME_ULID_OF_A_REPORT
func IsReportPath(id *url.URL) bool {
	return regexes.ReportPath.MatchString(id.Path)
}

// IsAcceptsPath returns true if the given URL path corresponds to eg /users/example_username/accepts/SOME_ULID_OF_AN_ACCEPT
func IsAcceptsPath(id *url.URL) bool {
	return regexes.AcceptsPath.MatchString(id.Path)
}

// ParseStatusesPath returns the username and ulid from a path such as /users/example_username/statuses/SOME_ULID_OF_A_STATUS
func ParseStatusesPath(id *url.URL) (username string, ulid string, err error) {
	matches := regexes.StatusesPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	ulid = matches[2]
	return
}

// ParseUserPath returns the username from a path such as /users/example_username
func ParseUserPath(id *url.URL) (username string, err error) {
	matches := regexes.UserPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseUserPath returns the username from a path such as /@example_username
func ParseUserWebPath(id *url.URL) (username string, err error) {
	matches := regexes.UserWebPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseInboxPath returns the username from a path such as /users/example_username/inbox
func ParseInboxPath(id *url.URL) (username string, err error) {
	matches := regexes.InboxPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseOutboxPath returns the username from a path such as /users/example_username/outbox
func ParseOutboxPath(id *url.URL) (username string, err error) {
	matches := regexes.OutboxPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseFollowersPath returns the username from a path such as /users/example_username/followers
func ParseFollowersPath(id *url.URL) (username string, err error) {
	matches := regexes.FollowersPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseFollowingPath returns the username from a path such as /users/example_username/following
func ParseFollowingPath(id *url.URL) (username string, err error) {
	matches := regexes.FollowingPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseLikedPath returns the username and ulid from a path such as /users/example_username/liked/SOME_ULID_OF_A_STATUS
func ParseLikedPath(id *url.URL) (username string, ulid string, err error) {
	matches := regexes.LikePath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	ulid = matches[2]
	return
}

// ParseBlockPath returns the username and ulid from a path such as /users/example_username/blocks/SOME_ULID_OF_A_BLOCK
func ParseBlockPath(id *url.URL) (username string, ulid string, err error) {
	matches := regexes.BlockPath.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	ulid = matches[2]
	return
}

// ParseReportPath returns the ulid from a path such as /reports/SOME_ULID_OF_A_REPORT
func ParseReportPath(id *url.URL) (ulid string, err error) {
	matches := regexes.ReportPath.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	ulid = matches[1]
	return
}
