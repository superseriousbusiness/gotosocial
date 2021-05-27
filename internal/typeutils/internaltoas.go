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
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
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
		avatar := &gtsmodel.MediaAttachment{}
		if err := c.db.GetByID(a.AvatarMediaAttachmentID, avatar); err != nil {
			return nil, err
		}

		iconProperty := streams.NewActivityStreamsIconProperty()

		iconImage := streams.NewActivityStreamsImage()

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
		header := &gtsmodel.MediaAttachment{}
		if err := c.db.GetByID(a.HeaderMediaAttachmentID, header); err != nil {
			return nil, err
		}

		headerProperty := streams.NewActivityStreamsImageProperty()

		headerImage := streams.NewActivityStreamsImage()

		mediaType := streams.NewActivityStreamsMediaTypeProperty()
		mediaType.Set(header.File.ContentType)
		headerImage.SetActivityStreamsMediaType(mediaType)

		headerURLProperty := streams.NewActivityStreamsUrlProperty()
		headerURL, err := url.Parse(header.URL)
		if err != nil {
			return nil, err
		}
		headerURLProperty.AppendIRI(headerURL)
		headerImage.SetActivityStreamsUrl(headerURLProperty)

		headerProperty.AppendActivityStreamsImage(headerImage)
	}

	return person, nil
}

func (c *converter) StatusToAS(s *gtsmodel.Status) (vocab.ActivityStreamsNote, error) {
	// ensure prerequisites here before we get stuck in

	// check if author account is already attached to status and attach it if not
	// if we can't retrieve this, bail here already because we can't attribute the status to anyone
	if s.GTSAuthorAccount == nil {
		a := &gtsmodel.Account{}
		if err := c.db.GetByID(s.AccountID, a); err != nil {
			return nil, fmt.Errorf("StatusToAS: error retrieving author account from db: %s", err)
		}
		s.GTSAuthorAccount = a
	}

	// create the Note!
	status := streams.NewActivityStreamsNote()

	// id
	statusURI, err := url.Parse(s.URI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.URI, err)
	}
	statusIDProp := streams.NewJSONLDIdProperty()
	statusIDProp.SetIRI(statusURI)
	status.SetJSONLDId(statusIDProp)

	// type
	// will be set automatically by go-fed

	// summary aka cw
	statusSummaryProp := streams.NewActivityStreamsSummaryProperty()
	statusSummaryProp.AppendXMLSchemaString(s.ContentWarning)
	status.SetActivityStreamsSummary(statusSummaryProp)

	// inReplyTo
	if s.InReplyToID != "" {
		// fetch the replied status if we don't have it on hand already
		if s.GTSReplyToStatus == nil {
			rs := &gtsmodel.Status{}
			if err := c.db.GetByID(s.InReplyToID, rs); err != nil {
				return nil, fmt.Errorf("StatusToAS: error retrieving replied-to status from db: %s", err)
			}
			s.GTSReplyToStatus = rs
		}
		rURI, err := url.Parse(s.GTSReplyToStatus.URI)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.GTSReplyToStatus.URI, err)
		}

		inReplyToProp := streams.NewActivityStreamsInReplyToProperty()
		inReplyToProp.AppendIRI(rURI)
		status.SetActivityStreamsInReplyTo(inReplyToProp)
	}

	// published
	publishedProp := streams.NewActivityStreamsPublishedProperty()
	publishedProp.Set(s.CreatedAt)
	status.SetActivityStreamsPublished(publishedProp)

	// url
	if s.URL != "" {
		sURL, err := url.Parse(s.URL)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.URL, err)
		}

		urlProp := streams.NewActivityStreamsUrlProperty()
		urlProp.AppendIRI(sURL)
		status.SetActivityStreamsUrl(urlProp)
	}

	// attributedTo
	authorAccountURI, err := url.Parse(s.GTSAuthorAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.GTSAuthorAccount.URI, err)
	}
	attributedToProp := streams.NewActivityStreamsAttributedToProperty()
	attributedToProp.AppendIRI(authorAccountURI)
	status.SetActivityStreamsAttributedTo(attributedToProp)

	// tags
	tagProp := streams.NewActivityStreamsTagProperty()

	// tag -- mentions
	for _, m := range s.GTSMentions {
		asMention, err := c.MentionToAS(m)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error converting mention to AS mention: %s", err)
		}
		tagProp.AppendActivityStreamsMention(asMention)
	}

	// tag -- emojis
	// TODO

	// tag -- hashtags
	// TODO

	status.SetActivityStreamsTag(tagProp)

	// parse out some URIs we need here
	authorFollowersURI, err := url.Parse(s.GTSAuthorAccount.FollowersURI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.GTSAuthorAccount.FollowersURI, err)
	}

	publicURI, err := url.Parse(asPublicURI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", asPublicURI, err)
	}

	// to and cc
	toProp := streams.NewActivityStreamsToProperty()
	ccProp := streams.NewActivityStreamsCcProperty()
	switch s.Visibility {
	case gtsmodel.VisibilityDirect:
		// if DIRECT, then only mentioned users should be added to TO, and nothing to CC
		for _, m := range s.GTSMentions {
			iri, err := url.Parse(m.GTSAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.GTSAccount.URI, err)
			}
			toProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityMutualsOnly:
		// TODO
	case gtsmodel.VisibilityFollowersOnly:
		// if FOLLOWERS ONLY then we want to add followers to TO, and mentions to CC
		toProp.AppendIRI(authorFollowersURI)
		for _, m := range s.GTSMentions {
			iri, err := url.Parse(m.GTSAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.GTSAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityUnlocked:
		// if UNLOCKED, we want to add followers to TO, and public and mentions to CC
		toProp.AppendIRI(authorFollowersURI)
		ccProp.AppendIRI(publicURI)
		for _, m := range s.GTSMentions {
			iri, err := url.Parse(m.GTSAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.GTSAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityPublic:
		// if PUBLIC, we want to add public to TO, and followers and mentions to CC
		toProp.AppendIRI(publicURI)
		ccProp.AppendIRI(authorFollowersURI)
		for _, m := range s.GTSMentions {
			iri, err := url.Parse(m.GTSAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.GTSAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	}
	status.SetActivityStreamsTo(toProp)
	status.SetActivityStreamsCc(ccProp)

	// conversation
	// TODO

	// content -- the actual post itself
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(s.Content)
	status.SetActivityStreamsContent(contentProp)

	// attachment
	attachmentProp := streams.NewActivityStreamsAttachmentProperty()
	for _, a := range s.GTSMediaAttachments {
		doc, err := c.AttachmentToAS(a)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error converting attachment: %s", err)
		}
		attachmentProp.AppendActivityStreamsDocument(doc)
	}
	status.SetActivityStreamsAttachment(attachmentProp)

	// replies
	// TODO

	return status, nil
}

func (c *converter) FollowToAS(f *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (vocab.ActivityStreamsFollow, error) {
	// parse out the various URIs we need for this
	// origin account (who's doing the follow)
	originAccountURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing origin account uri: %s", err)
	}
	originActor := streams.NewActivityStreamsActorProperty()
	originActor.AppendIRI(originAccountURI)

	// target account (who's being followed)
	targetAccountURI, err := url.Parse(targetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing target account uri: %s", err)
	}

	// uri of the follow activity itself
	followURI, err := url.Parse(f.URI)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing follow uri: %s", err)
	}

	// start preparing the follow activity
	follow := streams.NewActivityStreamsFollow()

	// set the actor
	follow.SetActivityStreamsActor(originActor)

	// set the id
	followIDProp := streams.NewJSONLDIdProperty()
	followIDProp.SetIRI(followURI)
	follow.SetJSONLDId(followIDProp)

	// set the object
	followObjectProp := streams.NewActivityStreamsObjectProperty()
	followObjectProp.AppendIRI(targetAccountURI)
	follow.SetActivityStreamsObject(followObjectProp)

	// set the To property
	followToProp := streams.NewActivityStreamsToProperty()
	followToProp.AppendIRI(targetAccountURI)
	follow.SetActivityStreamsTo(followToProp)

	return follow, nil
}

func (c *converter) MentionToAS(m *gtsmodel.Mention) (vocab.ActivityStreamsMention, error) {
	if m.GTSAccount == nil {
		a := &gtsmodel.Account{}
		if err := c.db.GetWhere([]db.Where{{Key: "target_account_id", Value: m.TargetAccountID}}, a); err != nil {
			return nil, fmt.Errorf("MentionToAS: error getting target account from db: %s", err)
		}
		m.GTSAccount = a
	}

	// create the mention
	mention := streams.NewActivityStreamsMention()

	// href -- this should be the URI of the mentioned user
	hrefProp := streams.NewActivityStreamsHrefProperty()
	hrefURI, err := url.Parse(m.GTSAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("MentionToAS: error parsing uri %s: %s", m.GTSAccount.URI, err)
	}
	hrefProp.SetIRI(hrefURI)
	mention.SetActivityStreamsHref(hrefProp)

	// name -- this should be the namestring of the mentioned user, something like @whatever@example.org
	var domain string
	if m.GTSAccount.Domain == "" {
		domain = c.config.Host
	} else {
		domain = m.GTSAccount.Domain
	}
	username := m.GTSAccount.Username
	nameString := fmt.Sprintf("@%s@%s", username, domain)
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(nameString)
	mention.SetActivityStreamsName(nameProp)

	return mention, nil
}

func (c *converter) AttachmentToAS(a *gtsmodel.MediaAttachment) (vocab.ActivityStreamsDocument, error) {
	// type -- Document
	doc := streams.NewActivityStreamsDocument()

	// mediaType aka mime content type
	mediaTypeProp := streams.NewActivityStreamsMediaTypeProperty()
	mediaTypeProp.Set(a.File.ContentType)
	doc.SetActivityStreamsMediaType(mediaTypeProp)

	// url -- for the original image not the thumbnail
	urlProp := streams.NewActivityStreamsUrlProperty()
	imageURL, err := url.Parse(a.URL)
	if err != nil {
		return nil, fmt.Errorf("AttachmentToAS: error parsing uri %s: %s", a.URL, err)
	}
	urlProp.AppendIRI(imageURL)
	doc.SetActivityStreamsUrl(urlProp)

	// name -- aka image description
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(a.Description)
	doc.SetActivityStreamsName(nameProp)

	// blurhash
	blurProp := streams.NewTootBlurhashProperty()
	blurProp.Set(a.Blurhash)
	doc.SetTootBlurhash(blurProp)

	// focalpoint
	// TODO

	return doc, nil
}

/*
	We want to end up with something like this:

	{
	"@context": "https://www.w3.org/ns/activitystreams",
	"actor": "https://ondergrond.org/users/dumpsterqueer",
	"id": "https://ondergrond.org/users/dumpsterqueer#likes/44584",
	"object": "https://testingtesting123.xyz/users/gotosocial_test_account/statuses/771aea80-a33d-4d6d-8dfd-57d4d2bfcbd4",
	"type": "Like"
	}
*/
func (c *converter) FaveToAS(f *gtsmodel.StatusFave) (vocab.ActivityStreamsLike, error) {
	// check if targetStatus is already pinned to this fave, and fetch it if not
	if f.GTSStatus == nil {
		s := &gtsmodel.Status{}
		if err := c.db.GetByID(f.StatusID, s); err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching target status from database: %s", err)
		}
		f.GTSStatus = s
	}

	// check if the targetAccount is already pinned to this fave, and fetch it if not
	if f.GTSTargetAccount == nil {
		a := &gtsmodel.Account{}
		if err := c.db.GetByID(f.TargetAccountID, a); err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching target account from database: %s", err)
		}
		f.GTSTargetAccount = a
	}

	// check if the faving account is already pinned to this fave, and fetch it if not
	if f.GTSFavingAccount == nil {
		a := &gtsmodel.Account{}
		if err := c.db.GetByID(f.AccountID, a); err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching faving account from database: %s", err)
		}
		f.GTSFavingAccount = a
	}

	// create the like
	like := streams.NewActivityStreamsLike()

	// set the actor property to the fave-ing account's URI
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI, err := url.Parse(f.GTSFavingAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.GTSFavingAccount.URI, err)
	}
	actorProp.AppendIRI(actorIRI)
	like.SetActivityStreamsActor(actorProp)

	// set the ID property to the fave's URI
	idProp := streams.NewJSONLDIdProperty()
	idIRI, err := url.Parse(f.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.URI, err)
	}
	idProp.Set(idIRI)
	like.SetJSONLDId(idProp)

	// set the object property to the target status's URI
	objectProp := streams.NewActivityStreamsObjectProperty()
	statusIRI, err := url.Parse(f.GTSStatus.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.GTSStatus.URI, err)
	}
	objectProp.AppendIRI(statusIRI)
	like.SetActivityStreamsObject(objectProp)

	// set the TO property to the target account's IRI
	toProp := streams.NewActivityStreamsToProperty()
	toIRI, err := url.Parse(f.GTSTargetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.GTSTargetAccount.URI, err)
	}
	toProp.AppendIRI(toIRI)
	like.SetActivityStreamsTo(toProp)

	return like, nil
}
