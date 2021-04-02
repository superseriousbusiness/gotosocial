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

import "fmt"

type URIs struct {
	HostURL     string
	UserURL     string
	StatusesURL string

	UserURI       string
	StatusesURI   string
	InboxURI      string
	OutboxURI     string
	FollowersURI  string
	CollectionURI string
}

func GenerateURIs(username string, protocol string, host string) *URIs {
	hostURL := fmt.Sprintf("%s://%s", protocol, host)
	userURL := fmt.Sprintf("%s/@%s", hostURL, username)
	statusesURL := fmt.Sprintf("%s/statuses", userURL)

	userURI := fmt.Sprintf("%s/users/%s", hostURL, username)
	statusesURI := fmt.Sprintf("%s/statuses", userURI)
	inboxURI := fmt.Sprintf("%s/inbox", userURI)
	outboxURI := fmt.Sprintf("%s/outbox", userURI)
	followersURI := fmt.Sprintf("%s/followers", userURI)
	collectionURI := fmt.Sprintf("%s/collections/featured", userURI)
	return &URIs{
		HostURL:       hostURL,
		UserURL:       userURL,
		StatusesURL:   statusesURL,
		
		UserURI:       userURI,
		StatusesURI:   statusesURI,
		InboxURI:      inboxURI,
		OutboxURI:     outboxURI,
		FollowersURI:  followersURI,
		CollectionURI: collectionURI,
	}
}
