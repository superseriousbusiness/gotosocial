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

package typeutils

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
)

func accountableForActorType(actorType gtsmodel.AccountActorType) ap.Accountable {
	switch actorType {
	case gtsmodel.AccountActorTypeApplication:
		return streams.NewActivityStreamsApplication()
	case gtsmodel.AccountActorTypeGroup:
		return streams.NewActivityStreamsGroup()
	case gtsmodel.AccountActorTypeOrganization:
		return streams.NewActivityStreamsOrganization()
	case gtsmodel.AccountActorTypePerson:
		return streams.NewActivityStreamsPerson()
	case gtsmodel.AccountActorTypeService:
		return streams.NewActivityStreamsService()
	default:
		panic("invalid actor type")
	}
}

// AccountToAS converts a gts model
// account into an accountable.
func (c *Converter) AccountToAS(
	ctx context.Context,
	a *gtsmodel.Account,
) (ap.Accountable, error) {
	// Use appropriate underlying
	// actor type of accountable.
	accountable := accountableForActorType(a.ActorType)

	// id should be the activitypub URI of this user
	// something like https://example.org/users/example_user
	profileIDURI, err := url.Parse(a.URI)
	if err != nil {
		return nil, err
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(profileIDURI)
	accountable.SetJSONLDId(idProp)

	// published
	// The moment when the account was created.
	publishedProp := streams.NewActivityStreamsPublishedProperty()
	publishedProp.Set(a.CreatedAt)
	accountable.SetActivityStreamsPublished(publishedProp)

	// following
	// The URI for retrieving a list of accounts this user is following
	followingURI, err := url.Parse(a.FollowingURI)
	if err != nil {
		return nil, err
	}
	followingProp := streams.NewActivityStreamsFollowingProperty()
	followingProp.SetIRI(followingURI)
	accountable.SetActivityStreamsFollowing(followingProp)

	// followers
	// The URI for retrieving a list of this user's followers
	followersURI, err := url.Parse(a.FollowersURI)
	if err != nil {
		return nil, err
	}
	followersProp := streams.NewActivityStreamsFollowersProperty()
	followersProp.SetIRI(followersURI)
	accountable.SetActivityStreamsFollowers(followersProp)

	// inbox
	// the activitypub inbox of this user for accepting messages
	inboxURI, err := url.Parse(a.InboxURI)
	if err != nil {
		return nil, err
	}
	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(inboxURI)
	accountable.SetActivityStreamsInbox(inboxProp)

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
		accountable.SetActivityStreamsEndpoints(endpointsProp)
	}

	// outbox
	// the activitypub outbox of this user for serving messages
	outboxURI, err := url.Parse(a.OutboxURI)
	if err != nil {
		return nil, err
	}
	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(outboxURI)
	accountable.SetActivityStreamsOutbox(outboxProp)

	// featured posts
	// Pinned posts.
	featuredURI, err := url.Parse(a.FeaturedCollectionURI)
	if err != nil {
		return nil, err
	}
	featuredProp := streams.NewTootFeaturedProperty()
	featuredProp.SetIRI(featuredURI)
	accountable.SetTootFeatured(featuredProp)

	// featuredTags
	// NOT IMPLEMENTED

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(a.Username)
	accountable.SetActivityStreamsPreferredUsername(preferredUsernameProp)

	// name
	// Used as profile display name.
	nameProp := streams.NewActivityStreamsNameProperty()
	if a.Username != "" {
		nameProp.AppendXMLSchemaString(a.DisplayName)
	} else {
		nameProp.AppendXMLSchemaString(a.Username)
	}
	accountable.SetActivityStreamsName(nameProp)

	// summary
	// Used as profile bio.
	if a.Note != "" {
		summaryProp := streams.NewActivityStreamsSummaryProperty()
		summaryProp.AppendXMLSchemaString(a.Note)
		accountable.SetActivityStreamsSummary(summaryProp)
	}

	// url
	// Used as profile link.
	profileURL, err := url.Parse(a.URL)
	if err != nil {
		return nil, err
	}
	urlProp := streams.NewActivityStreamsUrlProperty()
	urlProp.AppendIRI(profileURL)
	accountable.SetActivityStreamsUrl(urlProp)

	// manuallyApprovesFollowers
	// Will be shown as a locked account.
	manuallyApprovesFollowersProp := streams.NewActivityStreamsManuallyApprovesFollowersProperty()
	manuallyApprovesFollowersProp.Set(*a.Locked)
	accountable.SetActivityStreamsManuallyApprovesFollowers(manuallyApprovesFollowersProp)

	// discoverable
	// Will be shown in the profile directory.
	discoverableProp := streams.NewTootDiscoverableProperty()
	discoverableProp.Set(*a.Discoverable)
	accountable.SetTootDiscoverable(discoverableProp)

	// devices
	// NOT IMPLEMENTED, probably won't implement

	// alsoKnownAs
	// Required for Move activity.
	if l := len(a.AlsoKnownAsURIs); l != 0 {
		alsoKnownAsURIs := make([]*url.URL, l)
		for i, rawURL := range a.AlsoKnownAsURIs {
			uri, err := url.Parse(rawURL)
			if err != nil {
				return nil, err
			}

			alsoKnownAsURIs[i] = uri
		}

		ap.SetAlsoKnownAs(accountable, alsoKnownAsURIs)
	}

	// movedTo
	// Required for Move activity.
	if a.MovedToURI != "" {
		movedTo, err := url.Parse(a.MovedToURI)
		if err != nil {
			return nil, err
		}

		ap.SetMovedTo(accountable, movedTo)
	}

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
	accountable.SetW3IDSecurityV1PublicKey(publicKeyProp)

	// tags
	tagProp := streams.NewActivityStreamsTagProperty()

	// tag -- emojis
	emojis := a.Emojis
	if len(a.EmojiIDs) > len(emojis) {
		emojis = []*gtsmodel.Emoji{}
		for _, emojiID := range a.EmojiIDs {
			emoji, err := c.state.DB.GetEmojiByID(ctx, emojiID)
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

	accountable.SetActivityStreamsTag(tagProp)

	// attachment
	// Used for profile fields.
	if len(a.Fields) != 0 {
		attachmentProp := streams.NewActivityStreamsAttachmentProperty()

		for _, field := range a.Fields {
			propertyValue := streams.NewSchemaPropertyValue()

			nameProp := streams.NewActivityStreamsNameProperty()
			nameProp.AppendXMLSchemaString(field.Name)
			propertyValue.SetActivityStreamsName(nameProp)

			valueProp := streams.NewSchemaValueProperty()
			valueProp.Set(field.Value)
			propertyValue.SetSchemaValue(valueProp)

			attachmentProp.AppendSchemaPropertyValue(propertyValue)
		}

		accountable.SetActivityStreamsAttachment(attachmentProp)
	}

	// endpoints
	// NOT IMPLEMENTED -- this is for shared inbox which we don't use

	// icon
	// Used as profile avatar.
	if a.AvatarMediaAttachmentID != "" {
		if a.AvatarMediaAttachment == nil {
			avatar, err := c.state.DB.GetAttachmentByID(ctx, a.AvatarMediaAttachmentID)
			if err == nil {
				a.AvatarMediaAttachment = avatar
			} else {
				log.Errorf(ctx, "error getting Avatar with id %s: %s", a.AvatarMediaAttachmentID, err)
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
			accountable.SetActivityStreamsIcon(iconProperty)
		}
	}

	// image
	// Used as profile header.
	if a.HeaderMediaAttachmentID != "" {
		if a.HeaderMediaAttachment == nil {
			header, err := c.state.DB.GetAttachmentByID(ctx, a.HeaderMediaAttachmentID)
			if err == nil {
				a.HeaderMediaAttachment = header
			} else {
				log.Errorf(ctx, "error getting Header with id %s: %s", a.HeaderMediaAttachmentID, err)
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
			accountable.SetActivityStreamsImage(headerProperty)
		}
	}

	return accountable, nil
}

// AccountToASMinimal converts a gts model account
// into an activity streams person or service.
//
// The returned account will just have the Type, Username,
// PublicKey, and ID properties set. This is suitable for
// serving to requesters to whom we want to give as little
// information as possible because we don't trust them (yet).
func (c *Converter) AccountToASMinimal(
	ctx context.Context,
	a *gtsmodel.Account,
) (ap.Accountable, error) {
	// Use appropriate underlying
	// actor type of accountable.
	accountable := accountableForActorType(a.ActorType)

	// id should be the activitypub URI of this user
	// something like https://example.org/users/example_user
	profileIDURI, err := url.Parse(a.URI)
	if err != nil {
		return nil, err
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(profileIDURI)
	accountable.SetJSONLDId(idProp)

	// preferredUsername
	// Used for Webfinger lookup. Must be unique on the domain, and must correspond to a Webfinger acct: URI.
	preferredUsernameProp := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProp.SetXMLSchemaString(a.Username)
	accountable.SetActivityStreamsPreferredUsername(preferredUsernameProp)

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
	accountable.SetW3IDSecurityV1PublicKey(publicKeyProp)

	return accountable, nil
}

// StatusToAS converts a gts model status into an ActivityStreams Statusable implementation, suitable for federation
func (c *Converter) StatusToAS(ctx context.Context, s *gtsmodel.Status) (ap.Statusable, error) {
	// Ensure the status model is fully populated.
	// The status and poll models are REQUIRED so nothing to do if this fails.
	if err := c.state.DB.PopulateStatus(ctx, s); err != nil {
		return nil, gtserror.Newf("error populating status: %w", err)
	}

	var status ap.Statusable

	if s.Poll != nil {
		// If status has poll available, we convert
		// it as an AS Question (similar to a Note).
		poll := streams.NewActivityStreamsQuestion()

		// Add required status poll data to AS Question.
		if err := c.addPollToAS(s.Poll, poll); err != nil {
			return nil, gtserror.Newf("error converting poll: %w", err)
		}

		// Set poll as status.
		status = poll
	} else {
		// Else we converter it as an AS Note.
		status = streams.NewActivityStreamsNote()
	}

	// id
	statusURI, err := url.Parse(s.URI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", s.URI, err)
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
	if s.InReplyToURI != "" {
		rURI, err := url.Parse(s.InReplyToURI)
		if err != nil {
			return nil, gtserror.Newf("error parsing url %s: %w", s.InReplyToURI, err)
		}

		inReplyToProp := streams.NewActivityStreamsInReplyToProperty()
		inReplyToProp.AppendIRI(rURI)
		status.SetActivityStreamsInReplyTo(inReplyToProp)
	}

	// Set created / updated at properties.
	ap.SetPublished(status, s.CreatedAt)
	if at := s.EditedAt; !at.IsZero() {
		ap.SetUpdated(status, at)
	}

	// url
	if s.URL != "" {
		sURL, err := url.Parse(s.URL)
		if err != nil {
			return nil, gtserror.Newf("error parsing url %s: %w", s.URL, err)
		}

		urlProp := streams.NewActivityStreamsUrlProperty()
		urlProp.AppendIRI(sURL)
		status.SetActivityStreamsUrl(urlProp)
	}

	// attributedTo
	authorAccountURI, err := url.Parse(s.Account.URI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", s.Account.URI, err)
	}
	attributedToProp := streams.NewActivityStreamsAttributedToProperty()
	attributedToProp.AppendIRI(authorAccountURI)
	status.SetActivityStreamsAttributedTo(attributedToProp)

	// tags
	tagProp := streams.NewActivityStreamsTagProperty()

	// tag -- mentions
	mentions := s.Mentions
	if len(s.MentionIDs) != len(mentions) {
		mentions, err = c.state.DB.GetMentions(ctx, s.MentionIDs)
		if err != nil {
			return nil, gtserror.Newf("error getting mentions: %w", err)
		}
	}
	for _, m := range mentions {
		asMention, err := c.MentionToAS(ctx, m)
		if err != nil {
			return nil, gtserror.Newf("error converting mention to AS mention: %w", err)
		}
		tagProp.AppendActivityStreamsMention(asMention)
	}

	// tag -- emojis
	emojis := s.Emojis
	if len(s.EmojiIDs) != len(emojis) {
		emojis, err = c.state.DB.GetEmojisByIDs(ctx, s.EmojiIDs)
		if err != nil {
			return nil, gtserror.Newf("error getting emojis from database: %w", err)
		}
	}
	for _, emoji := range emojis {
		asEmoji, err := c.EmojiToAS(ctx, emoji)
		if err != nil {
			return nil, gtserror.Newf("error converting emoji to AS emoji: %w", err)
		}
		tagProp.AppendTootEmoji(asEmoji)
	}

	// tag -- hashtags
	hashtags := s.Tags
	if len(s.TagIDs) != len(hashtags) {
		hashtags, err = c.state.DB.GetTags(ctx, s.TagIDs)
		if err != nil {
			return nil, gtserror.Newf("error getting tags: %w", err)
		}
	}
	for _, ht := range hashtags {
		asHashtag, err := c.TagToAS(ctx, ht)
		if err != nil {
			return nil, gtserror.Newf("error converting tag to AS tag: %w", err)
		}
		tagProp.AppendTootHashtag(asHashtag)
	}
	status.SetActivityStreamsTag(tagProp)

	// parse out some URIs we need here
	authorFollowersURI, err := url.Parse(s.Account.FollowersURI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", s.Account.FollowersURI, err)
	}

	publicURI, err := url.Parse(pub.PublicActivityPubIRI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", pub.PublicActivityPubIRI, err)
	}

	// to and cc
	toProp := streams.NewActivityStreamsToProperty()
	ccProp := streams.NewActivityStreamsCcProperty()
	switch s.Visibility {
	case gtsmodel.VisibilityDirect:
		// if DIRECT, then only mentioned users should be added to TO, and nothing to CC
		for _, m := range mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, gtserror.Newf("error parsing uri %s: %w", m.TargetAccount.URI, err)
			}
			toProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityMutualsOnly:
		// TODO
	case gtsmodel.VisibilityFollowersOnly:
		// if FOLLOWERS ONLY then we want to add followers to TO, and mentions to CC
		toProp.AppendIRI(authorFollowersURI)
		for _, m := range mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, gtserror.Newf("error parsing uri %s: %w", m.TargetAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityUnlocked:
		// if UNLOCKED, we want to add followers to TO, and public and mentions to CC
		toProp.AppendIRI(authorFollowersURI)
		ccProp.AppendIRI(publicURI)
		for _, m := range mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, gtserror.Newf("error parsing uri %s: %w", m.TargetAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	case gtsmodel.VisibilityPublic:
		// if PUBLIC, we want to add public to TO, and followers and mentions to CC
		toProp.AppendIRI(publicURI)
		ccProp.AppendIRI(authorFollowersURI)
		for _, m := range mentions {
			iri, err := url.Parse(m.TargetAccount.URI)
			if err != nil {
				return nil, gtserror.Newf("error parsing uri %s: %w", m.TargetAccount.URI, err)
			}
			ccProp.AppendIRI(iri)
		}
	}
	status.SetActivityStreamsTo(toProp)
	status.SetActivityStreamsCc(ccProp)

	// conversation
	// TODO

	// content -- the actual post
	// itself, plus the language
	contentProp := streams.NewActivityStreamsContentProperty()
	contentProp.AppendXMLSchemaString(s.Content)

	if s.Language != "" {
		contentProp.AppendRDFLangString(map[string]string{
			s.Language: s.Content,
		})
	}

	status.SetActivityStreamsContent(contentProp)

	// attachments
	if err := c.attachAttachments(ctx, s, status); err != nil {
		return nil, gtserror.Newf("error attaching attachments: %w", err)
	}

	// replies
	repliesCollection, err := c.StatusToASRepliesCollection(ctx, s, false)
	if err != nil {
		return nil, fmt.Errorf("error creating repliesCollection: %w", err)
	}

	repliesProp := streams.NewActivityStreamsRepliesProperty()
	repliesProp.SetActivityStreamsCollection(repliesCollection)
	status.SetActivityStreamsReplies(repliesProp)

	// sensitive
	sensitiveProp := streams.NewActivityStreamsSensitiveProperty()
	sensitiveProp.AppendXMLSchemaBoolean(*s.Sensitive)
	status.SetActivityStreamsSensitive(sensitiveProp)

	// interactionPolicy
	if ipa, ok := status.(ap.InteractionPolicyAware); ok {
		var p *gtsmodel.InteractionPolicy
		if s.InteractionPolicy != nil {
			// Use InteractionPolicy
			// set on the status.
			p = s.InteractionPolicy
		} else {
			// Fall back to default policy
			// for the status's visibility.
			p = gtsmodel.DefaultInteractionPolicyFor(s.Visibility)
		}
		policy, err := c.InteractionPolicyToASInteractionPolicy(ctx, p, s)
		if err != nil {
			return nil, fmt.Errorf("error creating interactionPolicy: %w", err)
		}

		// Set interaction policy.
		policyProp := streams.NewGoToSocialInteractionPolicyProperty()
		policyProp.AppendGoToSocialInteractionPolicy(policy)
		ipa.SetGoToSocialInteractionPolicy(policyProp)

		// Parse + set approvedBy.
		if s.ApprovedByURI != "" {
			approvedBy, err := url.Parse(s.ApprovedByURI)
			if err != nil {
				return nil, fmt.Errorf("error parsing approvedBy: %w", err)
			}

			approvedByProp := streams.NewGoToSocialApprovedByProperty()
			approvedByProp.Set(approvedBy)
			ipa.SetGoToSocialApprovedBy(approvedByProp)
		}
	}

	return status, nil
}

func (c *Converter) addPollToAS(poll *gtsmodel.Poll, dst ap.Pollable) error {
	var optionsProp interface {
		// the minimum interface for appending AS Notes
		// to an AS type options property of some kind.
		AppendActivityStreamsNote(vocab.ActivityStreamsNote)
	}

	if len(poll.Options) != len(poll.Votes) {
		return gtserror.Newf("invalid poll %s", poll.ID)
	}

	if !*poll.HideCounts {
		// Set total no. voting accounts.
		ap.SetVotersCount(dst, *poll.Voters)
	}

	if *poll.Multiple {
		// Create new multiple-choice (AnyOf) property for poll.
		anyOfProp := streams.NewActivityStreamsAnyOfProperty()
		dst.SetActivityStreamsAnyOf(anyOfProp)
		optionsProp = anyOfProp
	} else {
		// Create new single-choice (OneOf) property for poll.
		oneOfProp := streams.NewActivityStreamsOneOfProperty()
		dst.SetActivityStreamsOneOf(oneOfProp)
		optionsProp = oneOfProp
	}

	for i, name := range poll.Options {
		// Create new Note object to represent option.
		note := streams.NewActivityStreamsNote()

		// Create new name property and set the option name.
		nameProp := streams.NewActivityStreamsNameProperty()
		nameProp.AppendXMLSchemaString(name)
		note.SetActivityStreamsName(nameProp)

		if !*poll.HideCounts {
			// Create new total items property to hold the vote count.
			totalItemsProp := streams.NewActivityStreamsTotalItemsProperty()
			totalItemsProp.Set(poll.Votes[i])

			// Create new replies property with collection to encompass count.
			repliesProp := streams.NewActivityStreamsRepliesProperty()
			collection := streams.NewActivityStreamsCollection()
			collection.SetActivityStreamsTotalItems(totalItemsProp)
			repliesProp.SetActivityStreamsCollection(collection)

			// Attach the replies to Note object.
			note.SetActivityStreamsReplies(repliesProp)
		}

		// Append the note to options property.
		optionsProp.AppendActivityStreamsNote(note)
	}

	if !poll.ExpiresAt.IsZero() {
		// Set poll endTime property.
		ap.SetEndTime(dst, poll.ExpiresAt)
	}

	if !poll.ClosedAt.IsZero() {
		// Poll is closed, set closed property.
		ap.AppendClosed(dst, poll.ClosedAt)
	}

	return nil
}

// StatusToASDelete converts a gts model status into a Delete of that status, using just the
// URI of the status as object, and addressing the Delete appropriately.
func (c *Converter) StatusToASDelete(ctx context.Context, s *gtsmodel.Status) (vocab.ActivityStreamsDelete, error) {
	// Parse / fetch some information
	// we need to create the Delete.

	if s.Account == nil {
		var err error
		s.Account, err = c.state.DB.GetAccountByID(ctx, s.AccountID)
		if err != nil {
			return nil, fmt.Errorf("StatusToASDelete: error retrieving author account from db: %w", err)
		}
	}

	actorIRI, err := url.Parse(s.AccountURI)
	if err != nil {
		return nil, fmt.Errorf("StatusToASDelete: error parsing actorIRI %s: %w", s.AccountURI, err)
	}

	statusIRI, err := url.Parse(s.URI)
	if err != nil {
		return nil, fmt.Errorf("StatusToASDelete: error parsing statusIRI %s: %w", s.URI, err)
	}

	// Create a Delete.
	delete := streams.NewActivityStreamsDelete()

	// Set appropriate actor for the Delete.
	deleteActor := streams.NewActivityStreamsActorProperty()
	deleteActor.AppendIRI(actorIRI)
	delete.SetActivityStreamsActor(deleteActor)

	// Set the status IRI as the 'object' property.
	// We should avoid serializing the whole status
	// when doing a delete because it's wasteful and
	// could accidentally leak the now-deleted status.
	deleteObject := streams.NewActivityStreamsObjectProperty()
	deleteObject.AppendIRI(statusIRI)
	delete.SetActivityStreamsObject(deleteObject)

	// Address the Delete appropriately.
	toProp := streams.NewActivityStreamsToProperty()
	ccProp := streams.NewActivityStreamsCcProperty()

	// Unless the status was a direct message, we can
	// address the Delete To the ActivityPub Public URI.
	// This ensures that the Delete will have as wide an
	// audience as possible.
	//
	// Because we're using just the status URI, not the
	// whole status, it won't leak any sensitive info.
	// At worst, a remote instance becomes aware of the
	// URI for a status which is now deleted anyway.
	if s.Visibility != gtsmodel.VisibilityDirect {
		publicURI, err := url.Parse(pub.PublicActivityPubIRI)
		if err != nil {
			return nil, fmt.Errorf("StatusToASDelete: error parsing url %s: %w", pub.PublicActivityPubIRI, err)
		}
		toProp.AppendIRI(publicURI)

		actorFollowersURI, err := url.Parse(s.Account.FollowersURI)
		if err != nil {
			return nil, fmt.Errorf("StatusToASDelete: error parsing url %s: %w", s.Account.FollowersURI, err)
		}
		ccProp.AppendIRI(actorFollowersURI)
	}

	// Always include the replied-to account and any
	// mentioned accounts as addressees as well.
	//
	// Worst case scenario here is that a replied account
	// who wasn't mentioned (and perhaps didn't see the
	// message), sees that someone has now deleted a status
	// in which they were replied to but not mentioned. In
	// other words, they *might* see that someone subtooted
	// about them, but they won't know what was said.

	// Ensure mentions are populated.
	mentions := s.Mentions
	if len(s.MentionIDs) > len(mentions) {
		mentions, err = c.state.DB.GetMentions(ctx, s.MentionIDs)
		if err != nil {
			return nil, fmt.Errorf("StatusToASDelete: error getting mentions: %w", err)
		}
	}

	// Remember which accounts were mentioned
	// here to avoid duplicating them later.
	mentionedAccountIDs := make(map[string]interface{}, len(mentions))

	// For direct messages, add URI
	// to To, else just add to CC.
	var f func(*url.URL)
	if s.Visibility == gtsmodel.VisibilityDirect {
		f = toProp.AppendIRI
	} else {
		f = ccProp.AppendIRI
	}

	for _, m := range mentions {
		mentionedAccountIDs[m.TargetAccountID] = nil // Remember this ID.

		iri, err := url.Parse(m.TargetAccount.URI)
		if err != nil {
			return nil, fmt.Errorf("StatusToAS: error parsing uri %s: %s", m.TargetAccount.URI, err)
		}

		f(iri)
	}

	if s.InReplyToAccountID != "" {
		if _, ok := mentionedAccountIDs[s.InReplyToAccountID]; !ok {
			// Only address to this account if it
			// wasn't already included as a mention.
			if s.InReplyToAccount == nil {
				s.InReplyToAccount, err = c.state.DB.GetAccountByID(ctx, s.InReplyToAccountID)
				if err != nil && !errors.Is(err, db.ErrNoEntries) {
					return nil, fmt.Errorf("StatusToASDelete: db error getting account %s: %w", s.InReplyToAccountID, err)
				}
			}

			if s.InReplyToAccount != nil {
				inReplyToAccountURI, err := url.Parse(s.InReplyToAccount.URI)
				if err != nil {
					return nil, fmt.Errorf("StatusToASDelete: error parsing url %s: %w", s.InReplyToAccount.URI, err)
				}
				ccProp.AppendIRI(inReplyToAccountURI)
			}
		}
	}

	delete.SetActivityStreamsTo(toProp)
	delete.SetActivityStreamsCc(ccProp)

	return delete, nil
}

// FollowToASFollow converts a gts model Follow into an activity streams Follow, suitable for federation
func (c *Converter) FollowToAS(ctx context.Context, f *gtsmodel.Follow) (vocab.ActivityStreamsFollow, error) {
	if err := c.state.DB.PopulateFollow(ctx, f); err != nil {
		return nil, gtserror.Newf("error populating follow: %w", err)
	}

	// Parse out the various URIs we need for this
	// origin account (who's doing the follow).
	originAccountURI, err := url.Parse(f.Account.URI)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing origin account uri: %s", err)
	}
	originActor := streams.NewActivityStreamsActorProperty()
	originActor.AppendIRI(originAccountURI)

	// target account (who's being followed)
	targetAccountURI, err := url.Parse(f.TargetAccount.URI)
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

// MentionToAS converts a gts model mention into an activity streams Mention, suitable for federation
func (c *Converter) MentionToAS(ctx context.Context, m *gtsmodel.Mention) (vocab.ActivityStreamsMention, error) {
	if m.TargetAccount == nil {
		a, err := c.state.DB.GetAccountByID(ctx, m.TargetAccountID)
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

// TagToAS converts a gts model tag into a toot Hashtag, suitable for federation.
func (c *Converter) TagToAS(ctx context.Context, t *gtsmodel.Tag) (vocab.TootHashtag, error) {
	// This is probably already lowercase,
	// but let's err on the safe side.
	nameLower := strings.ToLower(t.Name)
	tagURLString := uris.URIForTag(nameLower)

	// Create the tag.
	tag := streams.NewTootHashtag()

	// `href` should be the URL of the tag.
	hrefProp := streams.NewActivityStreamsHrefProperty()
	tagURL, err := url.Parse(tagURLString)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", tagURLString, err)
	}
	hrefProp.SetIRI(tagURL)
	tag.SetActivityStreamsHref(hrefProp)

	// `name` should be the name of the tag with the # prefix.
	nameProp := streams.NewActivityStreamsNameProperty()
	nameProp.AppendXMLSchemaString("#" + nameLower)
	tag.SetActivityStreamsName(nameProp)

	return tag, nil
}

// EmojiToAS converts a gts emoji into a mastodon ns Emoji, suitable for federation.
// we're making something like this:
//
//	{
//		"id": "https://example.com/emoji/123",
//		"type": "Emoji",
//		"name": ":kappa:",
//		"icon": {
//			"type": "Image",
//			"mediaType": "image/png",
//			"url": "https://example.com/files/kappa.png"
//		}
//	}
func (c *Converter) EmojiToAS(ctx context.Context, e *gtsmodel.Emoji) (vocab.TootEmoji, error) {
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
	updatedProp.Set(e.UpdatedAt)
	emoji.SetActivityStreamsUpdated(updatedProp)

	return emoji, nil
}

// attachAttachments converts the attachments on the given status
// into Attachmentables, and appends them to the given Statusable.
func (c *Converter) attachAttachments(
	ctx context.Context,
	s *gtsmodel.Status,
	statusable ap.Statusable,
) error {
	// Ensure status attachments populated.
	if len(s.AttachmentIDs) != len(s.Attachments) {
		var err error
		s.Attachments, err = c.state.DB.GetAttachmentsByIDs(ctx, s.AttachmentIDs)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return gtserror.Newf("db error getting attachments: %w", err)
		}
	}

	// Prepare attachment property.
	attachmentProp := streams.NewActivityStreamsAttachmentProperty()
	defer statusable.SetActivityStreamsAttachment(attachmentProp)

	for _, a := range s.Attachments {

		// Use appropriate vocab.Type and
		// append function for this attachment.
		var (
			attachmentable ap.Attachmentable
			append         func()
		)
		switch a.Type {

		// png, gif, webp, jpeg, etc.
		case gtsmodel.FileTypeImage:
			t := streams.NewActivityStreamsImage()
			attachmentable = t
			append = func() { attachmentProp.AppendActivityStreamsImage(t) }

		// mp4, m4a, wmv, webm, etc.
		case gtsmodel.FileTypeVideo, gtsmodel.FileTypeGifv:
			t := streams.NewActivityStreamsVideo()
			attachmentable = t
			append = func() { attachmentProp.AppendActivityStreamsVideo(t) }

		// mp3, flac, ogg, wma, etc.
		case gtsmodel.FileTypeAudio:
			t := streams.NewActivityStreamsAudio()
			attachmentable = t
			append = func() { attachmentProp.AppendActivityStreamsAudio(t) }

		// Not sure, fall back to Document.
		default:
			t := streams.NewActivityStreamsDocument()
			attachmentable = t
			append = func() { attachmentProp.AppendActivityStreamsDocument(t) }
		}

		// `mediaType` ie., mime content type.
		ap.SetMediaType(attachmentable, a.File.ContentType)

		// URL of the media file.
		imageURL, err := url.Parse(a.URL)
		if err != nil {
			return gtserror.Newf("error parsing attachment url: %w", err)
		}
		ap.AppendURL(attachmentable, imageURL)

		// `summary` ie., media description / alt text
		ap.AppendSummary(attachmentable, a.Description)

		// `blurhash`
		ap.SetBlurhash(attachmentable, a.Blurhash)

		// Set `focalPoint` only if necessary.
		if a.FileMeta.Focus.X != 0 && a.FileMeta.Focus.Y != 0 {
			if withFocalPoint, ok := attachmentable.(ap.WithFocalPoint); ok {
				focalPointProp := streams.NewTootFocalPointProperty()
				focalPointProp.AppendXMLSchemaFloat(float64(a.FileMeta.Focus.X))
				focalPointProp.AppendXMLSchemaFloat(float64(a.FileMeta.Focus.Y))
				withFocalPoint.SetTootFocalPoint(focalPointProp)
			}
		}

		// Done, append
		// to Statusable.
		append()
	}

	statusable.SetActivityStreamsAttachment(attachmentProp)
	return nil
}

// FaveToAS converts a gts model status fave into an activityStreams LIKE, suitable for federation.
// We want to end up with something like this:
//
// {
// "@context": "https://www.w3.org/ns/activitystreams",
// "actor": "https://ondergrond.org/users/dumpsterqueer",
// "id": "https://ondergrond.org/users/dumpsterqueer#likes/44584",
// "object": "https://testingtesting123.xyz/users/gotosocial_test_account/statuses/771aea80-a33d-4d6d-8dfd-57d4d2bfcbd4",
// "type": "Like"
// }
func (c *Converter) FaveToAS(ctx context.Context, f *gtsmodel.StatusFave) (vocab.ActivityStreamsLike, error) {
	// check if targetStatus is already pinned to this fave, and fetch it if not
	if f.Status == nil {
		s, err := c.state.DB.GetStatusByID(ctx, f.StatusID)
		if err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching target status from database: %s", err)
		}
		f.Status = s
	}

	// check if the targetAccount is already pinned to this fave, and fetch it if not
	if f.TargetAccount == nil {
		a, err := c.state.DB.GetAccountByID(ctx, f.TargetAccountID)
		if err != nil {
			return nil, fmt.Errorf("FaveToAS: error fetching target account from database: %s", err)
		}
		f.TargetAccount = a
	}

	// check if the faving account is already pinned to this fave, and fetch it if not
	if f.Account == nil {
		a, err := c.state.DB.GetAccountByID(ctx, f.AccountID)
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

	// Parse + set approvedBy.
	if f.ApprovedByURI != "" {
		approvedBy, err := url.Parse(f.ApprovedByURI)
		if err != nil {
			return nil, fmt.Errorf("error parsing approvedBy: %w", err)
		}

		approvedByProp := streams.NewGoToSocialApprovedByProperty()
		approvedByProp.Set(approvedBy)
		like.SetGoToSocialApprovedBy(approvedByProp)
	}

	return like, nil
}

// BoostToAS converts a gts model boost into an activityStreams ANNOUNCE, suitable for federation
func (c *Converter) BoostToAS(ctx context.Context, boostWrapperStatus *gtsmodel.Status, boostingAccount *gtsmodel.Account, boostedAccount *gtsmodel.Account) (vocab.ActivityStreamsAnnounce, error) {
	// the boosted status is probably pinned to the boostWrapperStatus but double check to make sure
	if boostWrapperStatus.BoostOf == nil {
		b, err := c.state.DB.GetStatusByID(ctx, boostWrapperStatus.BoostOfID)
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

	// Parse + set approvedBy.
	if boostWrapperStatus.ApprovedByURI != "" {
		approvedBy, err := url.Parse(boostWrapperStatus.ApprovedByURI)
		if err != nil {
			return nil, fmt.Errorf("error parsing approvedBy: %w", err)
		}

		approvedByProp := streams.NewGoToSocialApprovedByProperty()
		approvedByProp.Set(approvedBy)
		announce.SetGoToSocialApprovedBy(approvedByProp)
	}

	return announce, nil
}

// BlockToAS converts a gts model block into an activityStreams BLOCK, suitable for federation.
// we want to end up with something like this:
//
//	{
//		"@context": "https://www.w3.org/ns/activitystreams",
//		"actor": "https://example.org/users/some_user",
//		"id":"https://example.org/users/some_user/blocks/SOME_ULID_OF_A_BLOCK",
//		"object":"https://some_other.instance/users/some_other_user",
//		"type":"Block"
//	}
func (c *Converter) BlockToAS(ctx context.Context, b *gtsmodel.Block) (vocab.ActivityStreamsBlock, error) {
	if b.Account == nil {
		a, err := c.state.DB.GetAccountByID(ctx, b.AccountID)
		if err != nil {
			return nil, fmt.Errorf("BlockToAS: error getting block owner account from database: %s", err)
		}
		b.Account = a
	}

	if b.TargetAccount == nil {
		a, err := c.state.DB.GetAccountByID(ctx, b.TargetAccountID)
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

// StatusToASRepliesCollection converts a gts model status into an activityStreams REPLIES collection.
// the goal is to end up with something like this:
//
//	{
//		"@context": "https://www.w3.org/ns/activitystreams",
//		"id": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies",
//		"type": "Collection",
//		"first": {
//		"id": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?page=true",
//		"type": "CollectionPage",
//		"next": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?only_other_accounts=true&page=true",
//		"partOf": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies",
//		"items": []
//		}
//	}
func (c *Converter) StatusToASRepliesCollection(ctx context.Context, status *gtsmodel.Status, onlyOtherAccounts bool) (vocab.ActivityStreamsCollection, error) {
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

// StatusURIsToASRepliesPage returns a collection page with appropriate next/part of pagination.
// the goal is to end up with something like this:
//
//	{
//		"@context": "https://www.w3.org/ns/activitystreams",
//		"id": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?only_other_accounts=true&page=true",
//		"type": "CollectionPage",
//		"next": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies?min_id=106720870266901180&only_other_accounts=true&page=true",
//		"partOf": "https://example.org/users/whatever/statuses/01FCNEXAGAKPEX1J7VJRPJP490/replies",
//		"items": [
//			"https://example.com/users/someone/statuses/106720752853216226",
//			"https://somewhere.online/users/eeeeeeeeeep/statuses/106720870163727231"
//		]
//	}
func (c *Converter) StatusURIsToASRepliesPage(ctx context.Context, status *gtsmodel.Status, onlyOtherAccounts bool, minID string, replies map[string]*url.URL) (vocab.ActivityStreamsCollectionPage, error) {
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

// StatusesToASOutboxPage returns an ordered collection page using the given statuses and parameters as contents.
//
// The maxID and minID should be the parameters that were passed to the database to obtain the given statuses.
// These will be used to create the 'id' field of the collection.
//
// OutboxID is used to create the 'partOf' field in the collection.
//
// Appropriate 'next' and 'prev' fields will be created based on the highest and lowest IDs present in the statuses slice.
// the goal is to end up with something like this:
//
//	{
//		"id": "https://example.org/users/whatever/outbox?page=true",
//		"type": "OrderedCollectionPage",
//		"next": "https://example.org/users/whatever/outbox?max_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
//		"prev": "https://example.org/users/whatever/outbox?min_id=01FJC1Q0E3SSQR59TD2M1KP4V8&page=true",
//		"partOf": "https://example.org/users/whatever/outbox",
//		"orderedItems": [
//			"id": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7/activity",
//			"type": "Create",
//			"actor": "https://example.org/users/whatever",
//			"published": "2021-10-18T20:06:18Z",
//			"to": [
//				"https://www.w3.org/ns/activitystreams#Public"
//			],
//			"cc": [
//				"https://example.org/users/whatever/followers"
//			],
//			"object": "https://example.org/users/whatever/statuses/01FJC1MKPVX2VMWP2ST93Q90K7"
//		]
//	}
func (c *Converter) StatusesToASOutboxPage(ctx context.Context, outboxID string, maxID string, minID string, statuses []*gtsmodel.Status) (vocab.ActivityStreamsOrderedCollectionPage, error) {
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

		activity := WrapStatusableInCreate(note, true)
		itemsProp.AppendActivityStreamsCreate(activity)

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

// StatusesToASFeaturedCollection converts a slice of statuses into an ordered collection
// of URIs, suitable for serializing and serving via the activitypub API.
func (c *Converter) StatusesToASFeaturedCollection(ctx context.Context, featuredCollectionID string, statuses []*gtsmodel.Status) (vocab.ActivityStreamsOrderedCollection, error) {
	collection := streams.NewActivityStreamsOrderedCollection()

	collectionIDProp := streams.NewJSONLDIdProperty()
	featuredCollectionIDURI, err := url.Parse(featuredCollectionID)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s", featuredCollectionID)
	}
	collectionIDProp.SetIRI(featuredCollectionIDURI)
	collection.SetJSONLDId(collectionIDProp)

	itemsProp := streams.NewActivityStreamsOrderedItemsProperty()
	for _, s := range statuses {
		uri, err := url.Parse(s.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing url %s", s.URI)
		}
		itemsProp.AppendIRI(uri)
	}
	collection.SetActivityStreamsOrderedItems(itemsProp)

	totalItemsProp := streams.NewActivityStreamsTotalItemsProperty()
	totalItemsProp.Set(len(statuses))
	collection.SetActivityStreamsTotalItems(totalItemsProp)

	return collection, nil
}

// ReportToASFlag converts a gts model report into an activitystreams FLAG, suitable for federation.
func (c *Converter) ReportToASFlag(ctx context.Context, r *gtsmodel.Report) (vocab.ActivityStreamsFlag, error) {
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
	instanceAccount, err := c.state.DB.GetInstanceAccount(ctx, "")
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

// PollVoteToASCreate converts a vote on a poll into a Create
// activity, suitable for federation, with each choice in the
// vote appended as a Note to the Create's Object field.
//
// TODO: as soon as other AP server implementations support
// the use of multiple objects in a single create, update this
// to return just the one create event again.
func (c *Converter) PollVoteToASCreates(
	ctx context.Context,
	vote *gtsmodel.PollVote,
) ([]vocab.ActivityStreamsCreate, error) {
	if len(vote.Choices) == 0 {
		panic("no vote.Choices")
	}

	// Ensure the vote is fully populated (this fetches author).
	if err := c.state.DB.PopulatePollVote(ctx, vote); err != nil {
		return nil, gtserror.Newf("error populating vote from db: %w", err)
	}

	// Get the vote author.
	author := vote.Account

	// Get the JSONLD ID IRI for vote author.
	authorIRI, err := url.Parse(author.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid author uri: %w", err)
	}

	// Get the vote poll.
	poll := vote.Poll

	// Ensure the poll is fully populated with status.
	if err := c.state.DB.PopulatePoll(ctx, poll); err != nil {
		return nil, gtserror.Newf("error populating poll from db: %w", err)
	}

	// Get the JSONLD ID IRI for poll's source status.
	statusIRI, err := url.Parse(poll.Status.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid status uri: %w", err)
	}

	// Get the JSONLD ID IRI for poll's author account.
	pollAuthorIRI, err := url.Parse(poll.Status.AccountURI)
	if err != nil {
		return nil, gtserror.Newf("invalid account uri: %w", err)
	}

	// Parse each choice to a Note and add it to the list of Creates.
	creates := make([]vocab.ActivityStreamsCreate, len(vote.Choices))
	for i, choice := range vote.Choices {

		// Allocate Create activity and address 'To' poll author.
		create := streams.NewActivityStreamsCreate()
		ap.AppendTo(create, pollAuthorIRI)

		// Create ID formatted as: {$voterIRI}/activity#vote{$index}/{$statusIRI}.
		createID := fmt.Sprintf("%s/activity#vote%d/%s", author.URI, i, poll.Status.URI)
		ap.MustSet(ap.SetJSONLDIdStr, ap.WithJSONLDId(create), createID)

		// Set Create actor appropriately.
		ap.AppendActorIRIs(create, authorIRI)

		// Set publish time for activity.
		ap.SetPublished(create, vote.CreatedAt)

		// Allocate new note to hold the vote.
		note := streams.NewActivityStreamsNote()

		// For AP IRI generate from author URI + poll ID + vote choice.
		id := fmt.Sprintf("%s#%s/votes/%d", author.URI, poll.ID, choice)
		ap.MustSet(ap.SetJSONLDIdStr, ap.WithJSONLDId(note), id)

		// Attach new name property to note with vote choice.
		nameProp := streams.NewActivityStreamsNameProperty()
		nameProp.AppendXMLSchemaString(poll.Options[choice])
		note.SetActivityStreamsName(nameProp)

		// Set 'to', 'attribTo', 'inReplyTo' fields.
		ap.AppendAttributedTo(note, authorIRI)
		ap.AppendInReplyTo(note, statusIRI)
		ap.AppendTo(note, pollAuthorIRI)

		// Append this note to the Create Object.
		appendStatusableToActivity(create, note, false)

		// Set create in slice.
		creates[i] = create
	}

	return creates, nil
}

// populateValuesForProp appends the given PolicyValues
// to the given property, for the given status.
func populateValuesForProp[T ap.WithIRI](
	prop ap.Property[T],
	status *gtsmodel.Status,
	urns gtsmodel.PolicyValues,
) error {
	iriStrs := make([]string, 0)

	for _, urn := range urns {
		switch urn {

		case gtsmodel.PolicyValueAuthor:
			iriStrs = append(iriStrs, status.Account.URI)

		case gtsmodel.PolicyValueMentioned:
			for _, m := range status.Mentions {
				iriStrs = append(iriStrs, m.TargetAccount.URI)
			}

		case gtsmodel.PolicyValueFollowing:
			iriStrs = append(iriStrs, status.Account.FollowingURI)

		case gtsmodel.PolicyValueFollowers:
			iriStrs = append(iriStrs, status.Account.FollowersURI)

		case gtsmodel.PolicyValuePublic:
			iriStrs = append(iriStrs, pub.PublicActivityPubIRI)

		default:
			iriStrs = append(iriStrs, string(urn))
		}
	}

	// Deduplicate the iri strings to
	// make sure we're not parsing + adding
	// the same string multiple times.
	iriStrs = xslices.Deduplicate(iriStrs)

	// Append them to the property.
	for _, iriStr := range iriStrs {
		iri, err := url.Parse(iriStr)
		if err != nil {
			return err
		}

		prop.AppendIRI(iri)
	}

	return nil
}

// InteractionPolicyToASInteractionPolicy returns a
// GoToSocial interaction policy suitable for federation.
//
// Note: This currently includes deprecated properties `always`
// and `approvalRequired`. These will be removed in v0.21.0.
func (c *Converter) InteractionPolicyToASInteractionPolicy(
	ctx context.Context,
	interactionPolicy *gtsmodel.InteractionPolicy,
	status *gtsmodel.Status,
) (vocab.GoToSocialInteractionPolicy, error) {
	policy := streams.NewGoToSocialInteractionPolicy()

	/*
		CAN LIKE
	*/

	// Build canLike
	canLike := streams.NewGoToSocialCanLike()

	// Build canLike.automaticApproval
	canLikeAutomaticApprovalProp := streams.NewGoToSocialAutomaticApprovalProperty()
	if err := populateValuesForProp(
		canLikeAutomaticApprovalProp,
		status,
		interactionPolicy.CanLike.AutomaticApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canLike.automaticApproval: %w", err)
	}

	// Set canLike.manualApproval
	canLike.SetGoToSocialAutomaticApproval(canLikeAutomaticApprovalProp)

	// Build canLike.manualApproval
	canLikeManualApprovalProp := streams.NewGoToSocialManualApprovalProperty()
	if err := populateValuesForProp(
		canLikeManualApprovalProp,
		status,
		interactionPolicy.CanLike.ManualApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canLike.manualApproval: %w", err)
	}

	// Set canLike.manualApproval.
	canLike.SetGoToSocialManualApproval(canLikeManualApprovalProp)

	// deprecated: Build canLike.always
	canLikeAlwaysProp := streams.NewGoToSocialAlwaysProperty()
	if err := populateValuesForProp(
		canLikeAlwaysProp,
		status,
		interactionPolicy.CanLike.AutomaticApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canLike.always: %w", err)
	}

	// deprecated: Set canLike.always
	canLike.SetGoToSocialAlways(canLikeAlwaysProp)

	// deprecated: Build canLike.approvalRequired
	canLikeApprovalRequiredProp := streams.NewGoToSocialApprovalRequiredProperty()
	if err := populateValuesForProp(
		canLikeApprovalRequiredProp,
		status,
		interactionPolicy.CanLike.ManualApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canLike.approvalRequired: %w", err)
	}

	// deprecated: Set canLike.approvalRequired.
	canLike.SetGoToSocialApprovalRequired(canLikeApprovalRequiredProp)

	// Set canLike on the policy.
	canLikeProp := streams.NewGoToSocialCanLikeProperty()
	canLikeProp.AppendGoToSocialCanLike(canLike)
	policy.SetGoToSocialCanLike(canLikeProp)

	/*
		CAN REPLY
	*/

	// Build canReply
	canReply := streams.NewGoToSocialCanReply()

	// Build canReply.automaticApproval
	canReplyAutomaticApprovalProp := streams.NewGoToSocialAutomaticApprovalProperty()
	if err := populateValuesForProp(
		canReplyAutomaticApprovalProp,
		status,
		interactionPolicy.CanReply.AutomaticApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canReply.automaticApproval: %w", err)
	}

	// Set canReply.manualApproval
	canReply.SetGoToSocialAutomaticApproval(canReplyAutomaticApprovalProp)

	// Build canReply.manualApproval
	canReplyManualApprovalProp := streams.NewGoToSocialManualApprovalProperty()
	if err := populateValuesForProp(
		canReplyManualApprovalProp,
		status,
		interactionPolicy.CanReply.ManualApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canReply.manualApproval: %w", err)
	}

	// Set canReply.manualApproval.
	canReply.SetGoToSocialManualApproval(canReplyManualApprovalProp)

	// deprecated: Build canReply.always
	canReplyAlwaysProp := streams.NewGoToSocialAlwaysProperty()
	if err := populateValuesForProp(
		canReplyAlwaysProp,
		status,
		interactionPolicy.CanReply.AutomaticApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canReply.always: %w", err)
	}

	// deprecated: Set canReply.always
	canReply.SetGoToSocialAlways(canReplyAlwaysProp)

	// deprecated: Build canReply.approvalRequired
	canReplyApprovalRequiredProp := streams.NewGoToSocialApprovalRequiredProperty()
	if err := populateValuesForProp(
		canReplyApprovalRequiredProp,
		status,
		interactionPolicy.CanReply.ManualApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canReply.approvalRequired: %w", err)
	}

	// deprecated: Set canReply.approvalRequired.
	canReply.SetGoToSocialApprovalRequired(canReplyApprovalRequiredProp)

	// Set canReply on the policy.
	canReplyProp := streams.NewGoToSocialCanReplyProperty()
	canReplyProp.AppendGoToSocialCanReply(canReply)
	policy.SetGoToSocialCanReply(canReplyProp)

	/*
		CAN ANNOUNCE
	*/

	// Build canAnnounce
	canAnnounce := streams.NewGoToSocialCanAnnounce()

	// Build canAnnounce.automaticApproval
	canAnnounceAutomaticApprovalProp := streams.NewGoToSocialAutomaticApprovalProperty()
	if err := populateValuesForProp(
		canAnnounceAutomaticApprovalProp,
		status,
		interactionPolicy.CanAnnounce.AutomaticApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canAnnounce.automaticApproval: %w", err)
	}

	// Set canAnnounce.manualApproval
	canAnnounce.SetGoToSocialAutomaticApproval(canAnnounceAutomaticApprovalProp)

	// Build canAnnounce.manualApproval
	canAnnounceManualApprovalProp := streams.NewGoToSocialManualApprovalProperty()
	if err := populateValuesForProp(
		canAnnounceManualApprovalProp,
		status,
		interactionPolicy.CanAnnounce.ManualApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canAnnounce.manualApproval: %w", err)
	}

	// Set canAnnounce.manualApproval.
	canAnnounce.SetGoToSocialManualApproval(canAnnounceManualApprovalProp)

	// deprecated: Build canAnnounce.always
	canAnnounceAlwaysProp := streams.NewGoToSocialAlwaysProperty()
	if err := populateValuesForProp(
		canAnnounceAlwaysProp,
		status,
		interactionPolicy.CanAnnounce.AutomaticApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canAnnounce.always: %w", err)
	}

	// deprecated: Set canAnnounce.always
	canAnnounce.SetGoToSocialAlways(canAnnounceAlwaysProp)

	// deprecated: Build canAnnounce.approvalRequired
	canAnnounceApprovalRequiredProp := streams.NewGoToSocialApprovalRequiredProperty()
	if err := populateValuesForProp(
		canAnnounceApprovalRequiredProp,
		status,
		interactionPolicy.CanAnnounce.ManualApproval,
	); err != nil {
		return nil, gtserror.Newf("error setting canAnnounce.approvalRequired: %w", err)
	}

	// deprecated: Set canAnnounce.approvalRequired.
	canAnnounce.SetGoToSocialApprovalRequired(canAnnounceApprovalRequiredProp)

	// Set canAnnounce on the policy.
	canAnnounceProp := streams.NewGoToSocialCanAnnounceProperty()
	canAnnounceProp.AppendGoToSocialCanAnnounce(canAnnounce)
	policy.SetGoToSocialCanAnnounce(canAnnounceProp)

	return policy, nil
}

// InteractionReqToASAccept converts a *gtsmodel.InteractionRequest
// to an ActivityStreams Accept, addressed to the interacting account.
func (c *Converter) InteractionReqToASAccept(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (vocab.ActivityStreamsAccept, error) {
	accept := streams.NewActivityStreamsAccept()

	acceptID, err := url.Parse(req.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid accept uri: %w", err)
	}

	actorIRI, err := url.Parse(req.TargetAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid account uri: %w", err)
	}

	objectIRI, err := url.Parse(req.InteractionURI)
	if err != nil {
		return nil, gtserror.Newf("invalid object uri: %w", err)
	}

	if req.Status == nil {
		req.Status, err = c.state.DB.GetStatusByID(ctx, req.StatusID)
		if err != nil {
			return nil, gtserror.Newf("db error getting interaction req target status: %w", err)
		}
	}

	targetIRI, err := url.Parse(req.Status.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid interaction req target status uri: %w", err)
	}

	toIRI, err := url.Parse(req.InteractingAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid interacting account uri: %w", err)
	}

	// Set id to the URI of
	// interaction request.
	ap.SetJSONLDId(accept, acceptID)

	// Actor is the account that
	// owns the approval / accept.
	ap.AppendActorIRIs(accept, actorIRI)

	// Object is the interaction URI.
	ap.AppendObjectIRIs(accept, objectIRI)

	// Target is the URI of the
	// status being interacted with.
	ap.AppendTargetIRIs(accept, targetIRI)

	// Address to the owner
	// of interaction URI.
	ap.AppendTo(accept, toIRI)

	// Whether or not we cc this Accept to
	// followers and public depends on the
	// type of interaction it Accepts.

	var cc bool
	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		// Accept of Like doesn't get cc'd
		// because it's not that important.

	case gtsmodel.InteractionReply:
		// Accept of reply gets cc'd.
		cc = true

	case gtsmodel.InteractionAnnounce:
		// Accept of announce gets cc'd.
		cc = true
	}

	if cc {
		publicIRI, err := url.Parse(pub.PublicActivityPubIRI)
		if err != nil {
			return nil, gtserror.Newf("invalid public uri: %w", err)
		}

		followersIRI, err := url.Parse(req.TargetAccount.FollowersURI)
		if err != nil {
			return nil, gtserror.Newf("invalid followers uri: %w", err)
		}

		ap.AppendCc(accept, publicIRI, followersIRI)
	}

	return accept, nil
}

// InteractionReqToASReject converts a *gtsmodel.InteractionRequest
// to an ActivityStreams Reject, addressed to the interacting account.
func (c *Converter) InteractionReqToASReject(
	ctx context.Context,
	req *gtsmodel.InteractionRequest,
) (vocab.ActivityStreamsReject, error) {
	reject := streams.NewActivityStreamsReject()

	rejectID, err := url.Parse(req.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid reject uri: %w", err)
	}

	actorIRI, err := url.Parse(req.TargetAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid account uri: %w", err)
	}

	objectIRI, err := url.Parse(req.InteractionURI)
	if err != nil {
		return nil, gtserror.Newf("invalid object uri: %w", err)
	}

	if req.Status == nil {
		req.Status, err = c.state.DB.GetStatusByID(ctx, req.StatusID)
		if err != nil {
			return nil, gtserror.Newf("db error getting interaction req target status: %w", err)
		}
	}

	targetIRI, err := url.Parse(req.Status.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid interaction req target status uri: %w", err)
	}

	toIRI, err := url.Parse(req.InteractingAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("invalid interacting account uri: %w", err)
	}

	// Set id to the URI of
	// interaction request.
	ap.SetJSONLDId(reject, rejectID)

	// Actor is the account that
	// owns the approval / reject.
	ap.AppendActorIRIs(reject, actorIRI)

	// Object is the interaction URI.
	ap.AppendObjectIRIs(reject, objectIRI)

	// Target is the URI of the
	// status being interacted with.
	ap.AppendTargetIRIs(reject, targetIRI)

	// Address to the owner
	// of interaction URI.
	ap.AppendTo(reject, toIRI)

	// Whether or not we cc this Reject to
	// followers and public depends on the
	// type of interaction it Rejects.

	var cc bool
	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		// Reject of Like doesn't get cc'd
		// because it's not that important.

	case gtsmodel.InteractionReply:
		// Reject of reply gets cc'd.
		cc = true

	case gtsmodel.InteractionAnnounce:
		// Reject of announce gets cc'd.
		cc = true
	}

	if cc {
		publicIRI, err := url.Parse(pub.PublicActivityPubIRI)
		if err != nil {
			return nil, gtserror.Newf("invalid public uri: %w", err)
		}

		followersIRI, err := url.Parse(req.TargetAccount.FollowersURI)
		if err != nil {
			return nil, gtserror.Newf("invalid followers uri: %w", err)
		}

		ap.AppendCc(reject, publicIRI, followersIRI)
	}

	return reject, nil
}
