/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Converts a gts model account into an Activity Streams person type.
func (c *converter) AccountToAS(ctx context.Context, a *gtsmodel.Account) (vocab.ActivityStreamsPerson, error) {
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

	// shared inbox -- only add this if we know for sure it has one
	if a.SharedInboxURI != nil && *a.SharedInboxURI != "" {
		sharedInboxURI, err := url.Parse(*a.SharedInboxURI)
		if err != nil {
			return nil, err
		}
		endpointsProp := streams.NewActivityStreamsEndpointsProperty()
		endpoints := streams.NewActivityStreamsEndpoints()
		sharedInboxProp := streams.NewActivityStreamsSharedInboxProperty()
		sharedInboxProp.SetIRI(sharedInboxURI)
		endpoints.SetActivityStreamsSharedInbox(sharedInboxProp)
		endpointsProp.AppendActivityStreamsEndpoints(endpoints)
		person.SetActivityStreamsEndpoints(endpointsProp)
	}

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
	manuallyApprovesFollowersProp := streams.NewActivityStreamsManuallyApprovesFollowersProperty()
	manuallyApprovesFollowersProp.Set(*a.Locked)
	person.SetActivityStreamsManuallyApprovesFollowers(manuallyApprovesFollowersProp)

	// discoverable
	// Will be shown in the profile directory.
	discoverableProp := streams.NewTootDiscoverableProperty()
	discoverableProp.Set(*a.Discoverable)
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

	// tags
	tagProp := streams.NewActivityStreamsTagProperty()

	// tag -- emojis
	emojis := a.Emojis
	if len(a.EmojiIDs) > len(emojis) {
		emojis = []*gtsmodel.Emoji{}
		for _, emojiID := range a.EmojiIDs {
			emoji, err := c.db.GetEmojiByID(ctx, emojiID)
			if err != nil {
				return nil, fmt.Errorf("AccountToAS: error getting emoji %s from database: %s", emojiID, err)
			}
			emojis = append(emojis, emoji)
		}
	}
	for _, emoji := range emojis {
		asEmoji, err := c.EmojiToAS(ctx, emoji)
		if err != nil {
			return nil, fmt.Errorf("AccountToAS: error converting emoji to AS emoji: %s", err)
		}
		tagProp.AppendTootEmoji(asEmoji)
	}

	// tag -- hashtags
	// TODO

	person.SetActivityStreamsTag(tagProp)

	// attachment
	// Used for profile fields.
	// TODO: The PropertyValue type has to be added: https://schema.org/PropertyValue

	// endpoints
	// NOT IMPLEMENTED -- this is for shared inbox which we don't use

	// icon
	// Used as profile avatar.
	if a.AvatarMediaAttachmentID != "" {
		if a.AvatarMediaAttachment == nil {
			avatar, err := c.db.GetAttachmentByID(ctx, a.AvatarMediaAttachmentID)
			if err == nil {
				a.AvatarMediaAttachment = avatar
			} else {
				log.Errorf("AccountToAS: error getting Avatar with id %s: %s", a.AvatarMediaAttachmentID, err)
			}
		}

		if a.AvatarMediaAttachment != nil {
			iconProperty := streams.NewActivityStreamsIconProperty()

			iconImage := streams.NewActivityStreamsImage()

			mediaType := streams.NewActivityStreamsMediaTypeProperty()
			mediaType.Set(a.AvatarMediaAttachment.File.ContentType)
			iconImage.SetActivityStreamsMediaType(mediaType)

			avatarURLProperty := streams.NewActivityStreamsUrlProperty()
			avatarURL, err := url.Parse(a.AvatarMediaAttachment.URL)
			if err != nil {
				return nil, err
			}
			avatarURLProperty.AppendIRI(avatarURL)
			iconImage.SetActivityStreamsUrl(avatarURLProperty)

			iconProperty.AppendActivityStreamsImage(iconImage)
			person.SetActivityStreamsIcon(iconProperty)
		}
	}

	// image
	// Used as profile header.
	if a.HeaderMediaAttachmentID != "" {
		if a.HeaderMediaAttachment == nil {
			header, err := c.db.GetAttachmentByID(ctx, a.HeaderMediaAttachmentID)
			if err == nil {
				a.HeaderMediaAttachment = header
			} else {
				log.Errorf("AccountToAS: error getting Header with id %s: %s", a.HeaderMediaAttachmentID, err)
			}
		}

		if a.HeaderMediaAttachment != nil {
			headerProperty := streams.NewActivityStreamsImageProperty()

			headerImage := streams.NewActivityStreamsImage()

			mediaType := streams.NewActivityStreamsMediaTypeProperty()
			mediaType.Set(a.HeaderMediaAttachment.File.ContentType)
			headerImage.SetActivityStreamsMediaType(mediaType)

			headerURLProperty := streams.NewActivityStreamsUrlProperty()
			headerURL, err := url.Parse(a.HeaderMediaAttachment.URL)
			if err != nil {
				return nil, err
			}
			headerURLProperty.AppendIRI(headerURL)
			headerImage.SetActivityStreamsUrl(headerURLProperty)

			headerProperty.AppendActivityStreamsImage(headerImage)
			person.SetActivityStreamsImage(headerProperty)
		}
	}

	return person, nil
}

// Converts a gts model account into a VERY MINIMAL Activity Streams person type.
//
// The returned account will just have the Type, Username, PublicKey, and ID properties set.
func (c *converter) AccountToASMinimal(ctx context.Context, a *gtsmodel.Account) (vocab.ActivityStreamsPerson, error) {
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

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(a.Username)
	person.SetActivityStreamsPreferredUsername(preferredUsernameProp)

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

	return person, nil
}

func (c *converter) StatusToAS(ctx context.Context, s *gtsmodel.Status) (vocab.ActivityStreamsNote, error) {
	// ensure prerequisites here before we get stuck in

	// check if author account is already attached to status and attach it if not
	// if we can't retrieve this, bail here already because we can't attribute the status to anyone
	if s.Account == nil {
		a, err := c.db.GetAccountByID(ctx, s.AccountID)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error retrieving author account from db: %s", err)
		}
		s.Account = a
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
		if s.InReplyTo == nil {
			rs, err := c.db.GetStatusByID(ctx, s.InReplyToID)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error getting replied to status %s: %s", s.InReplyToID, err)
			}
			s.InReplyTo = rs
		}
		rURI, err := url.Parse(s.InReplyTo.URI)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.InReplyTo.URI, err)
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
	authorAccountURI, err := url.Parse(s.Account.URI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.Account.URI, err)
	}
	attributedToProp := streams.NewActivityStreamsAttributedToProperty()
	attributedToProp.AppendIRI(authorAccountURI)
	status.SetActivityStreamsAttributedTo(attributedToProp)

	// tags
	tagProp := streams.NewActivityStreamsTagProperty()

	// tag -- mentions
	mentions := s.Mentions
	if len(s.MentionIDs) > len(mentions) {
		mentions = []*gtsmodel.Mention{}
		for _, mentionID := range s.MentionIDs {
			mention, err := c.db.GetMention(ctx, mentionID)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error getting mention %s from database: %s", mentionID, err)
			}
			mentions = append(mentions, mention)
		}
	}
	for _, m := range mentions {
		asMention, err := c.MentionToAS(ctx, m)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error converting mention to AS mention: %s", err)
		}
		tagProp.AppendActivityStreamsMention(asMention)
	}

	// tag -- emojis
	emojis := s.Emojis
	if len(s.EmojiIDs) > len(emojis) {
		emojis = []*gtsmodel.Emoji{}
		for _, emojiID := range s.EmojiIDs {
			emoji, err := c.db.GetEmojiByID(ctx, emojiID)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error getting emoji %s from database: %s", emojiID, err)
			}
			emojis = append(emojis, emoji)
		}
	}
	for _, emoji := range emojis {
		asEmoji, err := c.EmojiToAS(ctx, emoji)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error converting emoji to AS emoji: %s", err)
		}
		tagProp.AppendTootEmoji(asEmoji)
	}

	// tag -- hashtags
	// TODO

	status.SetActivityStreamsTag(tagProp)

	// parse out some URIs we need here
	authorFollowersURI, err := url.Parse(s.Account.FollowersURI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", s.Account.FollowersURI, err)
	}

	publicURI, err := url.Parse(pub.PublicActivityPubIRI)
	if err != nil {
		return nil, fmt.Errorf("StatusToAS: error parsing url %s: %s", pub.PublicActivityPubIRI, err)
	}

	// to and cc
	toProp := streams.NewActivityStreamsToProperty()
	ccProp := streams.NewActivityStreamsCcProperty()
	switch s.Visibility {
	case gtsmodel.VisibilityDirect:
		// if DIRECT, then only mentioned users should be added to TO, and nothing to CC
		for _, m := range s.Mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.TargetAccount.URI, err)
			}
			toProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityMutualsOnly:
		// TODO
	case gtsmodel.VisibilityFollowersOnly:
		// if FOLLOWERS ONLY then we want to add followers to TO, and mentions to CC
		toProp.AppendIRI(authorFollowersURI)
		for _, m := range s.Mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.TargetAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityUnlocked:
		// if UNLOCKED, we want to add followers to TO, and public and mentions to CC
		toProp.AppendIRI(authorFollowersURI)
		ccProp.AppendIRI(publicURI)
		for _, m := range s.Mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.TargetAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityPublic:
		// if PUBLIC, we want to add public to TO, and followers and mentions to CC
		toProp.AppendIRI(publicURI)
		ccProp.AppendIRI(authorFollowersURI)
		for _, m := range s.Mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.TargetAccount.URI, err)
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

	// attachments
	attachmentProp := streams.NewActivityStreamsAttachmentProperty()
	attachments := s.Attachments
	if len(s.AttachmentIDs) > len(attachments) {
		attachments = []*gtsmodel.MediaAttachment{}
		for _, attachmentID := range s.AttachmentIDs {
			attachment, err := c.db.GetAttachmentByID(ctx, attachmentID)
			if err != nil {
				return nil, fmt.Errorf("StatusToAS: error getting attachment %s from database: %s", attachmentID, err)
			}
			attachments = append(attachments, attachment)
		}
	}
	for _, a := range attachments {
		doc, err := c.AttachmentToAS(ctx, a)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error converting attachment: %s", err)
		}
		attachmentProp.AppendActivityStreamsDocument(doc)
	}
	status.SetActivityStreamsAttachment(attachmentProp)

	// replies
	repliesCollection, err := c.StatusToASRepliesCollection(ctx, s, false)
	if err != nil {
		return nil, fmt.Errorf("error creating repliesCollection: %s", err)
	}

	repliesProp := streams.NewActivityStreamsRepliesProperty()
	repliesProp.SetActivityStreamsCollection(repliesCollection)
	status.SetActivityStreamsReplies(repliesProp)

	// sensitive
	sensitiveProp := streams.NewActivityStreamsSensitiveProperty()
	sensitiveProp.AppendXMLSchemaBoolean(*s.Sensitive)
	status.SetActivityStreamsSensitive(sensitiveProp)

	return status, nil
}

func (c *converter) FollowToAS(ctx context.Context, f *gtsmodel.Follow, originAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (vocab.ActivityStreamsFollow, error) {
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

func (c *converter) MentionToAS(ctx context.Context, m *gtsmodel.Mention) (vocab.ActivityStreamsMention, error) {
	if m.TargetAccount == nil {
		a, err := c.db.GetAccountByID(ctx, m.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("MentionToAS: error getting target account from db: %s", err)
		}
		m.TargetAccount = a
	}

	// create the mention
	mention := streams.NewActivityStreamsMention()

	// href -- this should be the URI of the mentioned user
	hrefProp := streams.NewActivityStreamsHrefProperty()
	hrefURI, err := url.Parse(m.TargetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("MentionToAS: error parsing uri %s: %s", m.TargetAccount.URI, err)
	}
	hrefProp.SetIRI(hrefURI)
	mention.SetActivityStreamsHref(hrefProp)

	// name -- this should be the namestring of the mentioned user, something like @whatever@example.org
	var domain string
	if m.TargetAccount.Domain == "" {
		accountDomain := config.GetAccountDomain()
		if accountDomain == "" {
			accountDomain = config.GetHost()
		}
		domain = accountDomain
	} else {
		domain = m.TargetAccount.Domain
	}
	username := m.TargetAccount.Username
	nameString := fmt.Sprintf("@%s@%s", username, domain)
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString(nameString)
	mention.SetActivityStreamsName(nameProp)

	return mention, nil
}

/*
	 we're making something like this:
		{
			"id": "https://example.com/emoji/123",
			"type": "Emoji",
			"name": ":kappa:",
			"icon": {
				"type": "Image",
				"mediaType": "image/png",
				"url": "https://example.com/files/kappa.png"
			}
		}
*/
func (c *converter) EmojiToAS(ctx context.Context, e *gtsmodel.Emoji) (vocab.TootEmoji, error) {
	// create the emoji
	emoji := streams.NewTootEmoji()

	// set the ID property to the blocks's URI
	idProp := streams.NewJSONLDIdProperty()
	idIRI, err := url.Parse(e.URI)
	if err != nil {
		return nil, fmt.Errorf("EmojiToAS: error parsing uri %s: %s", e.URI, err)
	}
	idProp.Set(idIRI)
	emoji.SetJSONLDId(idProp)

	nameProp := streams.NewActivityStreamsNameProperty()
	nameString := fmt.Sprintf(":%s:", e.Shortcode)
	nameProp.AppendXMLSchemaString(nameString)
	emoji.SetActivityStreamsName(nameProp)

	iconProperty := streams.NewActivityStreamsIconProperty()
	iconImage := streams.NewActivityStreamsImage()

	mediaType := streams.NewActivityStreamsMediaTypeProperty()
	mediaType.Set(e.ImageContentType)
	iconImage.SetActivityStreamsMediaType(mediaType)

	emojiURLProperty := streams.NewActivityStreamsUrlProperty()
	emojiURL, err := url.Parse(e.ImageURL)
	if err != nil {
		return nil, fmt.Errorf("EmojiToAS: error parsing url %s: %s", e.ImageURL, err)
	}
	emojiURLProperty.AppendIRI(emojiURL)
	iconImage.SetActivityStreamsUrl(emojiURLProperty)

	iconProperty.AppendActivityStreamsImage(iconImage)
	emoji.SetActivityStreamsIcon(iconProperty)

	updatedProp := streams.NewActivityStreamsUpdatedProperty()
	updatedProp.Set(e.ImageUpdatedAt)
	emoji.SetActivityStreamsUpdated(updatedProp)

	return emoji, nil
}

func (c *converter) AttachmentToAS(ctx context.Context, a *gtsmodel.MediaAttachment) (vocab.ActivityStreamsDocument, error) {
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
func (c *converter) FaveToAS(ctx context.Context, f *gtsmodel.StatusFave) (vocab.ActivityStreamsLike, error) {
	// check if targetStatus is already pinned to this fave, and fetch it if not
	if f.Status == nil {
		s, err := c.db.GetStatusByID(ctx, f.StatusID)
		if err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching target status from database: %s", err)
		}
		f.Status = s
	}

	// check if the targetAccount is already pinned to this fave, and fetch it if not
	if f.TargetAccount == nil {
		a, err := c.db.GetAccountByID(ctx, f.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching target account from database: %s", err)
		}
		f.TargetAccount = a
	}

	// check if the faving account is already pinned to this fave, and fetch it if not
	if f.Account == nil {
		a, err := c.db.GetAccountByID(ctx, f.AccountID)
		if err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching faving account from database: %s", err)
		}
		f.Account = a
	}

	// create the like
	like := streams.NewActivityStreamsLike()

	// set the actor property to the fave-ing account's URI
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI, err := url.Parse(f.Account.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.Account.URI, err)
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
	statusIRI, err := url.Parse(f.Status.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.Status.URI, err)
	}
	objectProp.AppendIRI(statusIRI)
	like.SetActivityStreamsObject(objectProp)

	// set the TO property to the target account's IRI
	toProp := streams.NewActivityStreamsToProperty()
	toIRI, err := url.Parse(f.TargetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("FaveToAS: error parsing uri %s: %s", f.TargetAccount.URI, err)
	}
	toProp.AppendIRI(toIRI)
	like.SetActivityStreamsTo(toProp)

	return like, nil
}

func (c *converter) BoostToAS(ctx context.Context, boostWrapperStatus *gtsmodel.Status, boostingAccount *gtsmodel.Account, boostedAccount *gtsmodel.Account) (vocab.ActivityStreamsAnnounce, error) {
	// the boosted status is probably pinned to the boostWrapperStatus but double check to make sure
	if boostWrapperStatus.BoostOf == nil {
		b, err := c.db.GetStatusByID(ctx, boostWrapperStatus.BoostOfID)
		if err != nil {
			return nil, fmt.Errorf("BoostToAS: error getting status with ID %s from the db: %s", boostWrapperStatus.BoostOfID, err)
		}
		boostWrapperStatus.BoostOf = b
	}

	// create the announce
	announce := streams.NewActivityStreamsAnnounce()

	// set the actor
	boosterURI, err := url.Parse(boostingAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("BoostToAS: error parsing uri %s: %s", boostingAccount.URI, err)
	}
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(boosterURI)
	announce.SetActivityStreamsActor(actorProp)

	// set the ID
	boostIDURI, err := url.Parse(boostWrapperStatus.URI)
	if err != nil {
		return nil, fmt.Errorf("BoostToAS: error parsing uri %s: %s", boostWrapperStatus.URI, err)
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(boostIDURI)
	announce.SetJSONLDId(idProp)

	// set the object
	boostedStatusURI, err := url.Parse(boostWrapperStatus.BoostOf.URI)
	if err != nil {
		return nil, fmt.Errorf("BoostToAS: error parsing uri %s: %s", boostWrapperStatus.BoostOf.URI, err)
	}
	objectProp := streams.NewActivityStreamsObjectProperty()
	objectProp.AppendIRI(boostedStatusURI)
	announce.SetActivityStreamsObject(objectProp)

	// set the published time
	publishedProp := streams.NewActivityStreamsPublishedProperty()
	publishedProp.Set(boostWrapperStatus.CreatedAt)
	announce.SetActivityStreamsPublished(publishedProp)

	// set the to
	followersURI, err := url.Parse(boostingAccount.FollowersURI)
	if err != nil {
		return nil, fmt.Errorf("BoostToAS: error parsing uri %s: %s", boostingAccount.FollowersURI, err)
	}
	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(followersURI)
	announce.SetActivityStreamsTo(toProp)

	// set the cc
	ccProp := streams.NewActivityStreamsCcProperty()
	boostedAccountURI, err := url.Parse(boostedAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("BoostToAS: error parsing uri %s: %s", boostedAccount.URI, err)
	}
	ccProp.AppendIRI(boostedAccountURI)

	// maybe CC it to public depending on the boosted status visibility
	switch boostWrapperStatus.BoostOf.Visibility {
	case gtsmodel.VisibilityPublic, gtsmodel.VisibilityUnlocked:
		publicURI, err := url.Parse(pub.PublicActivityPubIRI)
		if err != nil {
			return nil, fmt.Errorf("BoostToAS: error parsing uri %s: %s", pub.PublicActivityPubIRI, err)
		}
		ccProp.AppendIRI(publicURI)
	}

	announce.SetActivityStreamsCc(ccProp)

	return announce, nil
}

/*
we want to end up with something like this:

	{
		"@context": "https://www.w3.org/ns/activitystreams",
		"actor": "https://example.org/users/some_user",
		"id":"https://example.org/users/some_user/blocks/SOME_ULID_OF_A_BLOCK",
		"object":"https://some_other.instance/users/some_other_user",
		"type":"Block"
	}
*/
func (c *converter) BlockToAS(ctx context.Context, b *gtsmodel.Block) (vocab.ActivityStreamsBlock, error) {
	if b.Account == nil {
		a, err := c.db.GetAccountByID(ctx, b.AccountID)
		if err != nil {
			return nil, fmt.Errorf("BlockToAS: error getting block owner account from database: %s", err)
		}
		b.Account = a
	}

	if b.TargetAccount == nil {
		a, err := c.db.GetAccountByID(ctx, b.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("BlockToAS: error getting block target account from database: %s", err)
		}
		b.TargetAccount = a
	}

	// create the block
	block := streams.NewActivityStreamsBlock()

	// set the actor property to the block-ing account's URI
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI, err := url.Parse(b.Account.URI)
	if err != nil {
		return nil, fmt.Errorf("BlockToAS: error parsing uri %s: %s", b.Account.URI, err)
	}
	actorProp.AppendIRI(actorIRI)
	block.SetActivityStreamsActor(actorProp)

	// set the ID property to the blocks's URI
	idProp := streams.NewJSONLDIdProperty()
	idIRI, err := url.Parse(b.URI)
	if err != nil {
		return nil, fmt.Errorf("BlockToAS: error parsing uri %s: %s", b.URI, err)
	}
	idProp.Set(idIRI)
	block.SetJSONLDId(idProp)

	// set the object property to the target account's URI
	objectProp := streams.NewActivityStreamsObjectProperty()
	targetIRI, err := url.Parse(b.TargetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("BlockToAS: error parsing uri %s: %s", b.TargetAccount.URI, err)
	}
	objectProp.AppendIRI(targetIRI)
	block.SetActivityStreamsObject(objectProp)

	// set the TO property to the target account's IRI
	toProp := streams.NewActivityStreamsToProperty()
	toIRI, err := url.Parse(b.TargetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("BlockToAS: error parsing uri %s: %s", b.TargetAccount.URI, err)
	}
	toProp.AppendIRI(toIRI)
	block.SetActivityStreamsTo(toProp)

	return block, nil
}

/*
the goal is to end up with something like this:

	{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies",
		"type": "Collection",
		"first": {
		"id": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?page=true",
		"type": "CollectionPage",
		"next": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?only_other_accounts=true&page=true",
		"partOf": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies",
		"items": []
		}
	}
*/
func (c *converter) StatusToASRepliesCollection(ctx context.Context, status *gtsmodel.Status, onlyOtherAccounts bool) (vocab.ActivityStreamsCollection, error) {
	collectionID := fmt.Sprintf("%s/replies", status.URI)
	collectionIDURI, err := url.Parse(collectionID)
	if err != nil {
		return nil, err
	}

	collection := streams.NewActivityStreamsCollection()

	// collection.id
	collectionIDProp := streams.NewJSONLDIdProperty()
	collectionIDProp.SetIRI(collectionIDURI)
	collection.SetJSONLDId(collectionIDProp)

	// first
	first := streams.NewActivityStreamsFirstProperty()
	firstPage := streams.NewActivityStreamsCollectionPage()

	// first.id
	firstPageIDProp := streams.NewJSONLDIdProperty()
	firstPageID, err := url.Parse(fmt.Sprintf("%s?page=true", collectionID))
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	firstPageIDProp.SetIRI(firstPageID)
	firstPage.SetJSONLDId(firstPageIDProp)

	// first.next
	nextProp := streams.NewActivityStreamsNextProperty()
	nextPropID, err := url.Parse(fmt.Sprintf("%s?only_other_accounts=%t&page=true", collectionID, onlyOtherAccounts))
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	nextProp.SetIRI(nextPropID)
	firstPage.SetActivityStreamsNext(nextProp)

	// first.partOf
	partOfProp := streams.NewActivityStreamsPartOfProperty()
	partOfProp.SetIRI(collectionIDURI)
	firstPage.SetActivityStreamsPartOf(partOfProp)

	first.SetActivityStreamsCollectionPage(firstPage)

	// collection.first
	collection.SetActivityStreamsFirst(first)

	return collection, nil
}

/*
the goal is to end up with something like this:

	{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?only_other_accounts=true&page=true",
		"type": "CollectionPage",
		"next": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?min_id=106720870266901180&only_other_accounts=true&page=true",
		"partOf": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies",
		"items": [
			"https://example.com/users/someone/statuses/106720752853216226",
			"https://somewhere.online/users/eeeeeeeeeep/statuses/106720870163727231"
		]
	}
*/
func (c *converter) StatusURIsToASRepliesPage(ctx context.Context, status *gtsmodel.Status, onlyOtherAccounts bool, minID string, replies map[string]*url.URL) (vocab.ActivityStreamsCollectionPage, error) {
	collectionID := fmt.Sprintf("%s/replies", status.URI)

	page := streams.NewActivityStreamsCollectionPage()

	// .id
	pageIDProp := streams.NewJSONLDIdProperty()
	pageIDString := fmt.Sprintf("%s?page=true&only_other_accounts=%t", collectionID, onlyOtherAccounts)
	if minID != "" {
		pageIDString = fmt.Sprintf("%s&min_id=%s", pageIDString, minID)
	}

	pageID, err := url.Parse(pageIDString)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	pageIDProp.SetIRI(pageID)
	page.SetJSONLDId(pageIDProp)

	// .partOf
	collectionIDURI, err := url.Parse(collectionID)
	if err != nil {
		return nil, err
	}
	partOfProp := streams.NewActivityStreamsPartOfProperty()
	partOfProp.SetIRI(collectionIDURI)
	page.SetActivityStreamsPartOf(partOfProp)

	// .items
	items := streams.NewActivityStreamsItemsProperty()
	var highestID string
	for k, v := range replies {
		items.AppendIRI(v)
		if k > highestID {
			highestID = k
		}
	}
	page.SetActivityStreamsItems(items)

	// .next
	nextProp := streams.NewActivityStreamsNextProperty()
	nextPropIDString := fmt.Sprintf("%s?only_other_accounts=%t&page=true", collectionID, onlyOtherAccounts)
	if highestID != "" {
		nextPropIDString = fmt.Sprintf("%s&min_id=%s", nextPropIDString, highestID)
	}

	nextPropID, err := url.Parse(nextPropIDString)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	nextProp.SetIRI(nextPropID)
	page.SetActivityStreamsNext(nextProp)

	return page, nil
}

/*
the goal is to end up with something like this:

	{
		"id": "https://example.org/users/whatever/outbox?page=true",
		"type": "OrderedCollectionPage",
		"next": "https://example.org/users/whatever/outbox?max_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
		"prev": "https://example.org/users/whatever/outbox?min_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
		"partOf": "https://example.org/users/whatever/outbox",
		"orderedItems": [
			"id": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7/activity",
			"type": "Create",
			"actor": "https://example.org/users/whatever",
			"published": "2021-10-18T20:06:18Z",
			"to": [
				"https://www.w3.org/ns/activitystreams#Public"
			],
			"cc": [
				"https://example.org/users/whatever/followers"
			],
			"object": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7"
		]
	}
*/
func (c *converter) StatusesToASOutboxPage(ctx context.Context, outboxID string, maxID string, minID string, statuses []*gtsmodel.Status) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	page := streams.NewActivityStreamsOrderedCollectionPage()

	// .id
	pageIDProp := streams.NewJSONLDIdProperty()
	pageID := fmt.Sprintf("%s?page=true", outboxID)
	if minID != "" {
		pageID = fmt.Sprintf("%s&minID=%s", pageID, minID)
	}
	if maxID != "" {
		pageID = fmt.Sprintf("%s&maxID=%s", pageID, maxID)
	}
	pageIDURI, err := url.Parse(pageID)
	if err != nil {
		return nil, err
	}
	pageIDProp.SetIRI(pageIDURI)
	page.SetJSONLDId(pageIDProp)

	// .partOf
	collectionIDURI, err := url.Parse(outboxID)
	if err != nil {
		return nil, err
	}
	partOfProp := streams.NewActivityStreamsPartOfProperty()
	partOfProp.SetIRI(collectionIDURI)
	page.SetActivityStreamsPartOf(partOfProp)

	// .orderedItems
	itemsProp := streams.NewActivityStreamsOrderedItemsProperty()
	var highest string
	var lowest string
	for _, s := range statuses {
		note, err := c.StatusToAS(ctx, s)
		if err != nil {
			return nil, err
		}

		create, err := c.WrapNoteInCreate(note, true)
		if err != nil {
			return nil, err
		}

		itemsProp.AppendActivityStreamsCreate(create)

		if highest == "" || s.ID > highest {
			highest = s.ID
		}
		if lowest == "" || s.ID < lowest {
			lowest = s.ID
		}
	}
	page.SetActivityStreamsOrderedItems(itemsProp)

	// .next
	if lowest != "" {
		nextProp := streams.NewActivityStreamsNextProperty()
		nextPropIDString := fmt.Sprintf("%s?page=true&max_id=%s", outboxID, lowest)
		nextPropIDURI, err := url.Parse(nextPropIDString)
		if err != nil {
			return nil, err
		}
		nextProp.SetIRI(nextPropIDURI)
		page.SetActivityStreamsNext(nextProp)
	}

	// .prev
	if highest != "" {
		prevProp := streams.NewActivityStreamsPrevProperty()
		prevPropIDString := fmt.Sprintf("%s?page=true&min_id=%s", outboxID, highest)
		prevPropIDURI, err := url.Parse(prevPropIDString)
		if err != nil {
			return nil, err
		}
		prevProp.SetIRI(prevPropIDURI)
		page.SetActivityStreamsPrev(prevProp)
	}

	return page, nil
}

/*
we want something that looks like this:

	{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id": "https://example.org/users/whatever/outbox",
		"type": "OrderedCollection",
		"first": "https://example.org/users/whatever/outbox?page=true"
	}
*/
func (c *converter) OutboxToASCollection(ctx context.Context, outboxID string) (vocab.ActivityStreamsOrderedCollection, error) {
	collection := streams.NewActivityStreamsOrderedCollection()

	collectionIDProp := streams.NewJSONLDIdProperty()
	outboxIDURI, err := url.Parse(outboxID)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s", outboxID)
	}
	collectionIDProp.SetIRI(outboxIDURI)
	collection.SetJSONLDId(collectionIDProp)

	collectionFirstProp := streams.NewActivityStreamsFirstProperty()
	collectionFirstPropID := fmt.Sprintf("%s?page=true", outboxID)
	collectionFirstPropIDURI, err := url.Parse(collectionFirstPropID)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s", collectionFirstPropID)
	}
	collectionFirstProp.SetIRI(collectionFirstPropIDURI)
	collection.SetActivityStreamsFirst(collectionFirstProp)

	return collection, nil
}

func (c *converter) ReportToASFlag(ctx context.Context, r *gtsmodel.Report) (vocab.ActivityStreamsFlag, error) {
	flag := streams.NewActivityStreamsFlag()

	flagIDProp := streams.NewJSONLDIdProperty()
	idURI, err := url.Parse(r.URI)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %w", r.URI, err)
	}
	flagIDProp.SetIRI(idURI)
	flag.SetJSONLDId(flagIDProp)

	// for privacy, set the actor as the INSTANCE ACTOR,
	// not as the actor who created the report
	instanceAccount, err := c.db.GetInstanceAccount(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error getting instance account: %w", err)
	}
	instanceAccountIRI, err := url.Parse(instanceAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %w", instanceAccount.URI, err)
	}
	flagActorProp := streams.NewActivityStreamsActorProperty()
	flagActorProp.AppendIRI(instanceAccountIRI)
	flag.SetActivityStreamsActor(flagActorProp)

	// content should be the comment submitted when the report was created
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(r.Comment)
	flag.SetActivityStreamsContent(contentProp)

	// set at least the target account uri as the object of the flag
	objectProp := streams.NewActivityStreamsObjectProperty()
	targetAccountURI, err := url.Parse(r.TargetAccount.URI)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %w", r.TargetAccount.URI, err)
	}
	objectProp.AppendIRI(targetAccountURI)
	// also set status URIs if they were provided with the report
	for _, s := range r.Statuses {
		statusURI, err := url.Parse(s.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing url %s: %w", s.URI, err)
		}
		objectProp.AppendIRI(statusURI)
	}
	flag.SetActivityStreamsObject(objectProp)

	return flag, nil
}
