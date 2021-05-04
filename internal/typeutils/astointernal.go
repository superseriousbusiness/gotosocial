package typeutils

import (
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) ASPersonToAccount(person vocab.ActivityStreamsPerson) (*gtsmodel.Account, error) {

	// acct := &gtsmodel.Account{
	// 	URI:                     "",
	// 	URL:                     "",
	// 	ID:                      "",
	// 	Username:                "",
	// 	Domain:                  "",
	// 	AvatarMediaAttachmentID: "",
	// 	AvatarRemoteURL:         "",
	// 	HeaderMediaAttachmentID: "",
	// 	HeaderRemoteURL:         "",
	// 	DisplayName:             "",
	// 	Fields:                  nil,
	// 	Note:                    "",
	// 	Memorial:                false,
	// 	MovedToAccountID:        "",
	// 	CreatedAt:               time.Time{},
	// 	UpdatedAt:               time.Time{},
	// 	Bot:                     false,
	// 	Reason:                  "",
	// 	Locked:                  false,
	// 	Discoverable:            true,
	// 	Privacy:                 "",
	// 	Sensitive:               false,
	// 	Language:                "",
	// 	LastWebfingeredAt:       time.Now(),
	// 	InboxURI:                "",
	// 	OutboxURI:               "",
	// 	FollowingURI:            "",
	// 	FollowersURI:            "",
	// 	FeaturedCollectionURI:   "",
	// 	ActorType:               gtsmodel.ActivityStreamsPerson,
	// 	AlsoKnownAs:             "",
	// 	PrivateKey:              nil,
	// 	PublicKey:               nil,
	// 	PublicKeyURI:            "",
	// 	SensitizedAt:            time.Time{},
	// 	SilencedAt:              time.Time{},
	// 	SuspendedAt:             time.Time{},
	// 	HideCollections:         false,
	// 	SuspensionOrigin:        "",
	// }

	// // ID
	// // Generate a new uuid for our particular database.
	// // This is distinct from the AP ID of the person.
	// id := uuid.NewString()
	// acct.ID = id

	// // Username
	// // We need this one so bail if it's not set.
	// username := person.GetActivityStreamsPreferredUsername()
	// if username == nil || username.GetXMLSchemaString() == "" {
	// 	return nil, errors.New("preferredusername was empty")
	// }
	// acct.Username = username.GetXMLSchemaString()

	// // Domain
	// // We need this one as well
	// acct.Domain = domain

	return nil, nil
}
