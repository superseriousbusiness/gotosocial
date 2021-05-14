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
	"strings"
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
)

// APContextKey is a type used specifically for settings values on contexts within go-fed AP request chains
type APContextKey string

const (
	// APActivity can be used to set and retrieve the actual go-fed pub.Activity within a context.
	APActivity APContextKey = "activity"
	// APAccount can be used the set and retrieve the account being interacted with
	APAccount APContextKey = "account"
	// APRequestingAccount can be used to set and retrieve the account of an incoming federation request.
	// This will often be the actor of the instance that's posting the request.
	APRequestingAccount APContextKey = "requestingAccount"
	// APRequestingActorIRI can be used to set and retrieve the actor of an incoming federation request.
	// This will usually be the owner of whatever activity is being posted.
	APRequestingActorIRI APContextKey = "requestingActorIRI"
	// APRequestingPublicKeyID can be used to set and retrieve the public key ID of an incoming federation request.
	APRequestingPublicKeyID APContextKey = "requestingPublicKeyID"
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
	publicKeyURI := fmt.Sprintf("%s#%s", userURI, PublicKeyPath)

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
	return userPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsInboxPath returns true if the given URL path corresponds to eg /users/example_username/inbox
func IsInboxPath(id *url.URL) bool {
	return inboxPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsOutboxPath returns true if the given URL path corresponds to eg /users/example_username/outbox
func IsOutboxPath(id *url.URL) bool {
	return outboxPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsInstanceActorPath returns true if the given URL path corresponds to eg /actors/example_username
func IsInstanceActorPath(id *url.URL) bool {
	return actorPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsFollowersPath returns true if the given URL path corresponds to eg /users/example_username/followers
func IsFollowersPath(id *url.URL) bool {
	return followersPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsFollowingPath returns true if the given URL path corresponds to eg /users/example_username/following
func IsFollowingPath(id *url.URL) bool {
	return followingPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsLikedPath returns true if the given URL path corresponds to eg /users/example_username/liked
func IsLikedPath(id *url.URL) bool {
	return likedPathRegex.MatchString(strings.ToLower(id.Path))
}

// IsStatusesPath returns true if the given URL path corresponds to eg /users/example_username/statuses/SOME_UUID_OF_A_STATUS
func IsStatusesPath(id *url.URL) bool {
	return statusesPathRegex.MatchString(strings.ToLower(id.Path))
}

// ParseStatusesPath returns the username and uuid from a path such as /users/example_username/statuses/SOME_UUID_OF_A_STATUS
func ParseStatusesPath(id *url.URL) (username string, uuid string, err error) {
	matches := statusesPathRegex.FindStringSubmatch(id.Path)
	if len(matches) != 3 {
		err = fmt.Errorf("expected 3 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	uuid = matches[2]
	return
}

// ParseUserPath returns the username from a path such as /users/example_username
func ParseUserPath(id *url.URL) (username string, err error) {
	matches := userPathRegex.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseInboxPath returns the username from a path such as /users/example_username/inbox
func ParseInboxPath(id *url.URL) (username string, err error) {
	matches := inboxPathRegex.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}

// ParseOutboxPath returns the username from a path such as /users/example_username/outbox
func ParseOutboxPath(id *url.URL) (username string, err error) {
	matches := outboxPathRegex.FindStringSubmatch(id.Path)
	if len(matches) != 2 {
		err = fmt.Errorf("expected 2 matches but matches length was %d", len(matches))
		return
	}
	username = matches[1]
	return
}
