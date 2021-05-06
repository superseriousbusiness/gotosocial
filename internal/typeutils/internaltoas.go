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

package typeutils

import (
	"crypto/x509"
	"encoding/pem"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)



// Converts a gts model account into an Activity Streams person type, following
// the spec laid out for mastodon here: https://docs.joinmastodon.org/spec/activitypub/
func (c *converter) AccountToAS(a *gtsmodel.Account) (vocab.ActivityStreamsPerson, error) {
	person := streams.NewActivityStreamsPerson()

	// id should be the activitypub URI of this user
	// something like https://example.org/users/example_user
	profileIDURI, err := url.Parse(a.URI)
	if err != nil {
		return nil, err
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(profileIDURI)
	person.SetJSONLDId(idProp)

	// following
	// The URI for retrieving a list of accounts this user is following
	followingURI, err := url.Parse(a.FollowingURI)
	if err != nil {
		return nil, err
	}
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingProp.SetIRI(followingURI)
	person.SetActivityStreamsFollowing(followingProp)

	// followers
	// The URI for retrieving a list of this user's followers
	followersURI, err := url.Parse(a.FollowersURI)
	if err != nil {
		return nil, err
	}
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersProp.SetIRI(followersURI)
	person.SetActivityStreamsFollowers(followersProp)

	// inbox
	// the activitypub inbox of this user for accepting messages
	inboxURI, err := url.Parse(a.InboxURI)
	if err != nil {
		return nil, err
	}
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(inboxURI)
	person.SetActivityStreamsInbox(inboxProp)

	// outbox
	// the activitypub outbox of this user for serving messages
	outboxURI, err := url.Parse(a.OutboxURI)
	if err != nil {
		return nil, err
	}
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(outboxURI)
	person.SetActivityStreamsOutbox(outboxProp)

	// featured posts
	// Pinned posts.
	featuredURI, err := url.Parse(a.FeaturedCollectionURI)
	if err != nil {
		return nil, err
	}
	featuredProp := streams.NewTootFeaturedProperty()
	featuredProp.SetIRI(featuredURI)
	person.SetTootFeatured(featuredProp)

	// featuredTags
	// NOT IMPLEMENTED

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(a.Username)
	person.SetActivityStreamsPreferredUsername(preferredUsernameProp)

	// name
	// Used as profile display name.
	nameProp := streams.NewActivityStreamsNameProperty()
	if a.Username != "" {
		nameProp.AppendXMLSchemaString(a.DisplayName)
	} else {
		nameProp.AppendXMLSchemaString(a.Username)
	}
	person.SetActivityStreamsName(nameProp)

	// summary
	// Used as profile bio.
	if a.Note != "" {
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(a.Note)
		person.SetActivityStreamsSummary(summaryProp)
	}

	// url
	// Used as profile link.
	profileURL, err := url.Parse(a.URL)
	if err != nil {
		return nil, err
	}
	urlProp := streams.NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(profileURL)
	person.SetActivityStreamsUrl(urlProp)

	// manuallyApprovesFollowers
	// Will be shown as a locked account.
	// TODO: NOT IMPLEMENTED **YET** -- this needs to be added as an activitypub extension to https://github.com/go-fed/activity, see https://github.com/go-fed/activity/tree/master/astool

	// discoverable
	// Will be shown in the profile directory.
	discoverableProp := streams.NewTootDiscoverableProperty()
	discoverableProp.Set(a.Discoverable)
	person.SetTootDiscoverable(discoverableProp)

	// devices
	// NOT IMPLEMENTED, probably won't implement

	// alsoKnownAs
	// Required for Move activity.
	// TODO: NOT IMPLEMENTED **YET** -- this needs to be added as an activitypub extension to https://github.com/go-fed/activity, see https://github.com/go-fed/activity/tree/master/astool

	// publicKey
	// Required for signatures.
	publicKeyProp := streams.NewW3IDSecurityV1PublicKeyProperty()

	// create the public key
	publicKey := streams.NewW3IDSecurityV1PublicKey()

	// set ID for the public key
	publicKeyIDProp := streams.NewJSONLDIdProperty()
	publicKeyURI, err := url.Parse(a.PublicKeyURI)
	if err != nil {
		return nil, err
	}
	publicKeyIDProp.SetIRI(publicKeyURI)
	publicKey.SetJSONLDId(publicKeyIDProp)

	// set owner for the public key
	publicKeyOwnerProp := streams.NewW3IDSecurityV1OwnerProperty()
	publicKeyOwnerProp.SetIRI(profileIDURI)
	publicKey.SetW3IDSecurityV1Owner(publicKeyOwnerProp)

	// set the pem key itself
	encodedPublicKey, err := x509.MarshalPKIXPublicKey(a.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: encodedPublicKey,
	})
	publicKeyPEMProp := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	publicKeyPEMProp.Set(string(publicKeyBytes))
	publicKey.SetW3IDSecurityV1PublicKeyPem(publicKeyPEMProp)

	// append the public key to the public key property
	publicKeyProp.AppendW3IDSecurityV1PublicKey(publicKey)

	// set the public key property on the Person
	person.SetW3IDSecurityV1PublicKey(publicKeyProp)

	// tag
	// TODO: Any tags used in the summary of this profile

	// attachment
	// Used for profile fields.
	// TODO: The PropertyValue type has to be added: https://schema.org/PropertyValue

	// endpoints
	// NOT IMPLEMENTED -- this is for shared inbox which we don't use

	// icon
	// Used as profile avatar.
	if a.AvatarMediaAttachmentID != "" {
		iconProperty := streams.NewActivityStreamsIconProperty()

		iconImage := streams.NewActivityStreamsImage()

		avatar := &gtsmodel.MediaAttachment{}
		if err := c.db.GetByID(a.AvatarMediaAttachmentID, avatar); err != nil {
			return nil, err
		}

		mediaType := streams.NewActivityStreamsMediaTypeProperty()
		mediaType.Set(avatar.File.ContentType)
		iconImage.SetActivityStreamsMediaType(mediaType)

		avatarURLProperty := streams.NewActivityStreamsUrlProperty()
		avatarURL, err := url.Parse(avatar.URL)
		if err != nil {
			return nil, err
		}
		avatarURLProperty.AppendIRI(avatarURL)
		iconImage.SetActivityStreamsUrl(avatarURLProperty)

		iconProperty.AppendActivityStreamsImage(iconImage)
		person.SetActivityStreamsIcon(iconProperty)
	}

	// image
	// Used as profile header.
	if a.HeaderMediaAttachmentID != "" {
		iconProperty := streams.NewActivityStreamsIconProperty()

		iconImage := streams.NewActivityStreamsImage()

		header := &gtsmodel.MediaAttachment{}
		if err := c.db.GetByID(a.HeaderMediaAttachmentID, header); err != nil {
			return nil, err
		}

		mediaType := streams.NewActivityStreamsMediaTypeProperty()
		mediaType.Set(header.File.ContentType)
		iconImage.SetActivityStreamsMediaType(mediaType)

		headerURLProperty := streams.NewActivityStreamsUrlProperty()
		headerURL, err := url.Parse(header.URL)
		if err != nil {
			return nil, err
		}
		headerURLProperty.AppendIRI(headerURL)
		iconImage.SetActivityStreamsUrl(headerURLProperty)

		iconProperty.AppendActivityStreamsImage(iconImage)
	}

	return person, nil
}

func (c *converter) StatusToAS(s *gtsmodel.Status) (vocab.ActivityStreamsNote, error) {
	return nil, nil
}
