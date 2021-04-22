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
)

const (
	// UsersPath is for serving users info
	UsersPath       = "users"
	// StatusesPath is for serving statuses
	StatusesPath    = "statuses"
	// InboxPath represents the webfinger inbox location
	InboxPath       = "inbox"
	// OutboxPath represents the webfinger outbox location
	OutboxPath      = "outbox"
	// FollowersPath represents the webfinger followers location
	FollowersPath   = "followers"
	// CollectionsPath represents the webfinger collections location
	CollectionsPath = "collections"
	// FeaturedPath represents the webfinger featured location
	FeaturedPath    = "featured"
)

// UserURIs contains a bunch of UserURIs and URLs for a user, host, account, etc.
type UserURIs struct {
	// The web URL of the instance host, eg https://example.org
	HostURL     string
	// The web URL of the user, eg., https://example.org/@example_user
	UserURL     string
	// The web URL for statuses of this user, eg., https://example.org/@example_user/statuses
	StatusesURL string

	// The webfinger URI of this user, eg., https://example.org/users/example_user
	UserURI       string
	// The webfinger URI for this user's statuses, eg., https://example.org/users/example_user/statuses
	StatusesURI   string
	// The webfinger URI for this user's activitypub inbox, eg., https://example.org/users/example_user/inbox
	InboxURI      string
	// The webfinger URI for this user's activitypub outbox, eg., https://example.org/users/example_user/outbox
	OutboxURI     string
	// The webfinger URI for this user's followers, eg., https://example.org/users/example_user/followers
	FollowersURI  string
	// The webfinger URI for this user's featured collections, eg., https://example.org/users/example_user/collections/featured
	CollectionURI string
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
	collectionURI := fmt.Sprintf("%s/%s/%s", userURI, CollectionsPath, FeaturedPath)
	return &UserURIs{
		HostURL:     hostURL,
		UserURL:     userURL,
		StatusesURL: statusesURL,

		UserURI:       userURI,
		StatusesURI:   statusesURI,
		InboxURI:      inboxURI,
		OutboxURI:     outboxURI,
		FollowersURI:  followersURI,
		CollectionURI: collectionURI,
	}
}

func ParseActivityPubRequestURL(id *url.URL) error {
	return nil
}
