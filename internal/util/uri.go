/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package util

import (
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

const (
	// UsersPath is for serving users info
	UsersPath = "users"
	// ActorsPath is for serving actors info
	ActorsPath = "actors"
	// StatusesPath is for serving statuses
	StatusesPath = "statuses"
	// InboxPath represents the webfinger inbox location
	InboxPath = "inbox"
	// OutboxPath represents the webfinger outbox location
	OutboxPath = "outbox"
	// FollowersPath represents the webfinger followers location
	FollowersPath = "followers"
	// FollowingPath represents the webfinger following location
	FollowingPath = "following"
	// LikedPath represents the webfinger liked location
	LikedPath = "liked"
	// CollectionsPath represents the webfinger collections location
	CollectionsPath = "collections"
	// FeaturedPath represents the webfinger featured location
	FeaturedPath = "featured"
	// PublicKeyPath is for serving an account's public key
	PublicKeyPath = "main-key"
	// FollowPath used to generate the URI for an individual follow or follow request
	FollowPath = "follow"
	// UpdatePath is used to generate the URI for an account update
	UpdatePath = "updates"
	// BlocksPath is used to generate the URI for a block
	BlocksPath = "blocks"
	// ConfirmEmailPath is used to generate the URI for an email confirmation link
	ConfirmEmailPath = "confirm_email"
)

// APContextKey is a type used specifically for settings values on contexts within go-fed AP request chains
type APContextKey string

const (
	// APActivity can be used to set and retrieve the actual go-fed pub.Activity within a context.
	APActivity APContextKey = "activity"
	// APReceivingAccount can be used the set and retrieve the account being interacted with / receiving an activity in their inbox.
	APReceivingAccount APContextKey = "account"
	// APRequestingAccount can be used to set and retrieve the account of an incoming federation request.
	// This will often be the actor of the instance that's posting the request.
	APRequestingAccount APContextKey = "requestingAccount"
	// APRequestingActorIRI can be used to set and retrieve the actor of an incoming federation request.
	// This will usually be the owner of whatever activity is being posted.
	APRequestingActorIRI APContextKey = "requestingActorIRI"
	// APRequestingPublicKeyVerifier can be used to set and retrieve the public key verifier of an incoming federation request.
	APRequestingPublicKeyVerifier APContextKey = "requestingPublicKeyVerifier"
	// APRequestingPublicKeySignature can be used to set and retrieve the value of the signature header of an incoming federation request.
	APRequestingPublicKeySignature APContextKey = "requestingPublicKeySignature"
	// APFromFederatorChanKey can be used to pass a pointer to the fromFederator channel into the federator for use in callbacks.
	APFromFederatorChanKey APContextKey = "fromFederatorChan"
)

type ginContextKey struct{}

// GinContextKey is used solely for setting and retrieving the gin context from a context.Context
var GinContextKey = &ginContextKey{}

// UserURIs contains a bunch of UserURIs and URLs for a user, host, account, etc.
type UserURIs struct {
	// The web URL of the instance host, eg https://example.org
	HostURL string
	// The web URL of the user, eg., https://example.org/@example_user
	UserURL string
	// The web URL for statuses of this user, eg., https://example.org/@example_user/statuses
	StatusesURL string

	// The webfinger URI of this user, eg., https://example.org/users/example_user
	UserURI string
	// The webfinger URI for this user's statuses, eg., https://example.org/users/example_user/statuses
	StatusesURI string
	// The webfinger URI for this user's activitypub inbox, eg., https://example.org/users/example_user/inbox
	InboxURI string
	// The webfinger URI for this user's activitypub outbox, eg., https://example.org/users/example_user/outbox
	OutboxURI string
	// The webfinger URI for this user's followers, eg., https://example.org/users/example_user/followers
	FollowersURI string
	// The webfinger URI for this user's following, eg., https://example.org/users/example_user/following
	FollowingURI string
	// The webfinger URI for this user's liked posts eg., https://example.org/users/example_user/liked
	LikedURI string
	// The webfinger URI for this user's featured collections, eg., https://example.org/users/example_user/collections/featured
	CollectionURI string
	// The URI for this user's public key, eg., https://example.org/users/example_user/publickey
	PublicKeyURI string
}

// GenerateURIForFollow returns the AP URI for a new follow -- something like:
// https://example.org/users/whatever_user/follow/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForFollow(username string, protocol string, host string, thisFollowID string) string {
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, FollowPath, thisFollowID)
}

// GenerateURIForLike returns the AP URI for a new like/fave -- something like:
// https://example.org/users/whatever_user/liked/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForLike(username string, protocol string, host string, thisFavedID string) string {
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, LikedPath, thisFavedID)
}

// GenerateURIForUpdate returns the AP URI for a new update activity -- something like:
// https://example.org/users/whatever_user#updates/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForUpdate(username string, protocol string, host string, thisUpdateID string) string {
	return fmt.Sprintf("%s://%s/%s/%s#%s/%s", protocol, host, UsersPath, username, UpdatePath, thisUpdateID)
}

// GenerateURIForBlock returns the AP URI for a new block activity -- something like:
// https://example.org/users/whatever_user/blocks/01F7XTH1QGBAPMGF49WJZ91XGC
func GenerateURIForBlock(username string, protocol string, host string, thisBlockID string) string {
	return fmt.Sprintf("%s://%s/%s/%s/%s/%s", protocol, host, UsersPath, username, BlocksPath, thisBlockID)
}

// GenerateURIForEmailConfirm returns a link for email confirmation -- something like:
// https://example.org/confirm_email?token=490e337c-0162-454f-ac48-4b22bb92a205
func GenerateURIForEmailConfirm(protocol string, host string, token string) string {
	return fmt.Sprintf("%s://%s/%s?token=%s", protocol, host, ConfirmEmailPath, token)
}

// GenerateURIsForAccount throws together a bunch of URIs for the given username, with the given protocol and host.
func GenerateURIsForAccount(username string, protocol string, host string) *UserURIs {
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

		UserURI:       userURI,
		StatusesURI:   statusesURI,
		InboxURI:      inboxURI,
		OutboxURI:     outboxURI,
		FollowersURI:  followersURI,
		FollowingURI:  followingURI,
		LikedURI:      likedURI,
		CollectionURI: collectionURI,
		PublicKeyURI:  publicKeyURI,
	}
}

// IsUserPath returns true if the given URL path corresponds to eg /users/example_username
func IsUserPath(id *url.URL) bool {
	return regexes.UserPath.MatchString(id.Path)
}

// IsInboxPath returns true if the given URL path corresponds to eg /users/example_username/inbox
func IsInboxPath(id *url.URL) bool {
	return regexes.InboxPath.MatchString(id.Path)
}

// IsOutboxPath returns true if the given URL path corresponds to eg /users/example_username/outbox
func IsOutboxPath(id *url.URL) bool {
	return regexes.OutboxPath.MatchString(id.Path)
}

// IsInstanceActorPath returns true if the given URL path corresponds to eg /actors/example_username
func IsInstanceActorPath(id *url.URL) bool {
	return regexes.ActorPath.MatchString(id.Path)
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
