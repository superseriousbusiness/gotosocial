package util

import "fmt"

type URIs struct {
	HostURL       string
	UserURL       string
	UserURI       string
	InboxURL      string
	OutboxURL     string
	FollowersURL  string
	CollectionURL string
}

func GenerateURIs(username string, protocol string, host string) *URIs {
	hostURL := fmt.Sprintf("%s://%s", protocol, host)
	userURL := fmt.Sprintf("%s/@%s", hostURL, username)
	userURI := fmt.Sprintf("%s/users/%s", hostURL, username)
	inboxURL := fmt.Sprintf("%s/inbox", userURI)
	outboxURL := fmt.Sprintf("%s/outbox", userURI)
	followersURL := fmt.Sprintf("%s/followers", userURI)
	collectionURL := fmt.Sprintf("%s/collections/featured", userURI)
	return &URIs{
		HostURL:       hostURL,
		UserURL:       userURL,
		UserURI:       userURI,
		InboxURL:      inboxURL,
		OutboxURL:     outboxURL,
		FollowersURL:  followersURL,
		CollectionURL: collectionURL,
	}
}
