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
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (c *converter) ASRepresentationToAccount(accountable Accountable) (*gtsmodel.Account, error) {
	// first check if we actually already know this account
	uriProp := accountable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found on person, or id was not an iri")
	}
	uri := uriProp.GetIRI()

	acct := &gtsmodel.Account{}
	err := c.db.GetWhere("uri", uri.String(), acct)
	if err == nil {
		// we already know this account so we can skip generating it
		return acct, nil
	}
	if _, ok := err.(db.ErrNoEntries); !ok {
		// we don't know the account and there's been a real error
		return nil, fmt.Errorf("error getting account with uri %s from the database: %s", uri.String(), err)
	}

	// we don't know the account so we need to generate it from the person -- at least we already have the URI!
	acct = &gtsmodel.Account{}
	acct.URI = uri.String()

	// Username aka preferredUsername
	// We need this one so bail if it's not set.
	username, err := extractPreferredUsername(accountable)
	if err != nil {
		return nil, fmt.Errorf("couldn't extract username: %s", err)
	}
	acct.Username = username

	// Domain
	acct.Domain = uri.Host

	// avatar aka icon
	// if this one isn't extractable in a format we recognise we'll just skip it
	if avatarURL, err := extractIconURL(accountable); err == nil {
		acct.AvatarRemoteURL = avatarURL.String()
	}

	// header aka image
	// if this one isn't extractable in a format we recognise we'll just skip it
	if headerURL, err := extractImageURL(accountable); err == nil {
		acct.HeaderRemoteURL = headerURL.String()
	}

	// display name aka name
	// we default to the username, but take the more nuanced name property if it exists
	acct.DisplayName = username
	if displayName, err := extractName(accountable); err == nil {
		acct.DisplayName = displayName
	}

	// TODO: fields aka attachment array

	// note aka summary
	note, err := extractSummary(accountable)
	if err == nil && note != "" {
		acct.Note = note
	}

	// check for bot and actor type
	switch gtsmodel.ActivityStreamsActor(accountable.GetTypeName()) {
	case gtsmodel.ActivityStreamsPerson, gtsmodel.ActivityStreamsGroup, gtsmodel.ActivityStreamsOrganization:
		// people, groups, and organizations aren't bots
		acct.Bot = false
		// apps and services are
	case gtsmodel.ActivityStreamsApplication, gtsmodel.ActivityStreamsService:
		acct.Bot = true
	default:
		// we don't know what this is!
		return nil, fmt.Errorf("type name %s not recognised or not convertible to gtsmodel.ActivityStreamsActor", accountable.GetTypeName())
	}
	acct.ActorType = gtsmodel.ActivityStreamsActor(accountable.GetTypeName())

	// TODO: locked aka manuallyApprovesFollowers

	// discoverable
	// default to false -- take custom value if it's set though
	acct.Discoverable = false
	discoverable, err := extractDiscoverable(accountable)
	if err == nil {
		acct.Discoverable = discoverable
	}

	// url property
	url, err := extractURL(accountable)
	if err != nil {
		return nil, fmt.Errorf("could not extract url for person with id %s: %s", uri.String(), err)
	}
	acct.URL = url.String()

	// InboxURI
	if accountable.GetActivityStreamsInbox() != nil || accountable.GetActivityStreamsInbox().GetIRI() != nil {
		acct.InboxURI = accountable.GetActivityStreamsInbox().GetIRI().String()
	}

	// OutboxURI
	if accountable.GetActivityStreamsOutbox() != nil && accountable.GetActivityStreamsOutbox().GetIRI() != nil {
		acct.OutboxURI = accountable.GetActivityStreamsOutbox().GetIRI().String()
	}

	// FollowingURI
	if accountable.GetActivityStreamsFollowing() != nil && accountable.GetActivityStreamsFollowing().GetIRI() != nil {
		acct.FollowingURI = accountable.GetActivityStreamsFollowing().GetIRI().String()
	}

	// FollowersURI
	if accountable.GetActivityStreamsFollowers() != nil && accountable.GetActivityStreamsFollowers().GetIRI() != nil {
		acct.FollowersURI = accountable.GetActivityStreamsFollowers().GetIRI().String()
	}

	// FeaturedURI
	if accountable.GetTootFeatured() != nil && accountable.GetTootFeatured().GetIRI() != nil {
		acct.FeaturedCollectionURI = accountable.GetTootFeatured().GetIRI().String()
	}

	// TODO: FeaturedTagsURI

	// TODO: alsoKnownAs

	// publicKey
	pkey, pkeyURL, err := extractPublicKeyForOwner(accountable, uri)
	if err != nil {
		return nil, fmt.Errorf("couldn't get public key for person %s: %s", uri.String(), err)
	}
	acct.PublicKey = pkey
	acct.PublicKeyURI = pkeyURL.String()

	return acct, nil
}

func (c *converter) ASStatusToStatus(statusable Statusable) (*gtsmodel.Status, error) {
	status := &gtsmodel.Status{}

	uriProp := statusable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found, or id was not an iri")
	}
	status.URI = uriProp.GetIRI().String()

	statusURL, err := extractURL(statusable)
	if err == nil {
		status.URL = statusURL.String()
	}

	if content, err := extractContent(statusable); err == nil {
		status.Content = content
	}

	attachments, err := extractAttachments(statusable); if err == nil {
		status.GTSMediaAttachments = attachments
	}

	hashtags, err := extractHashtags(statusable)
	if err == nil {
		status.GTSTags = hashtags
	}

	emojis, err := extractEmojis(statusable)
	if err == nil {
		status.GTSEmojis = emojis
	}

	mentions, err := extractMentions(statusable)
	if err == nil {
		status.GTSMentions = mentions
	}

	cw, err := extractSummary(statusable)
	if err == nil && cw != "" {
		status.ContentWarning = cw
	}

	inReplyToURI, err := extractInReplyToURI(statusable)
	if err == nil {
		inReplyToStatus := &gtsmodel.Status{}
		if err := c.db.GetWhere("uri", inReplyToURI.String(), inReplyToStatus); err == nil {
			status.InReplyToID = inReplyToStatus.ID
		}
	}

	published, err := extractPublished(statusable)
	if err == nil {
		status.CreatedAt = published
	}

	attributedTo, err := extractAttributedTo(statusable)
	if err != nil {
		return nil, errors.New("attributedTo was empty")
	}

	// if we don't know the account yet we can dereference it later
	statusOwner := &gtsmodel.Status{}
	if err := c.db.GetWhere("uri", attributedTo.String(), statusOwner); err == nil {
		status.AccountID = statusOwner.ID
	}


	return status, nil
}

// // id of the status in the database
// ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
// // uri at which this status is reachable
// URI string `pg:",unique"`
// // web url for viewing this status
// URL string `pg:",unique"`
// // the html-formatted content of this status
// Content string
// // Database IDs of any media attachments associated with this status
// Attachments []string `pg:",array"`
// // Database IDs of any tags used in this status
// Tags []string `pg:",array"`
// // Database IDs of any accounts mentioned in this status
// Mentions []string `pg:",array"`
// // Database IDs of any emojis used in this status
// Emojis []string `pg:",array"`
// // when was this status created?
// CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
// // when was this status updated?
// UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
// // is this status from a local account?
// Local bool
// // which account posted this status?
// AccountID string
// // id of the status this status is a reply to
// InReplyToID string
// // id of the account that this status replies to
// InReplyToAccountID string
// // id of the status this status is a boost of
// BoostOfID string
// // cw string for this status
// ContentWarning string
// // visibility entry for this status
// Visibility Visibility `pg:",notnull"`
// // mark the status as sensitive?
// Sensitive bool
// // what language is this status written in?
// Language string
// // Which application was used to create this status?
// CreatedWithApplicationID string
// // advanced visibility for this status
// VisibilityAdvanced *VisibilityAdvanced
// // What is the activitystreams type of this status? See: https://www.w3.org/TR/activitystreams-vocabulary/#object-types
// // Will probably almost always be Note but who knows!.
// ActivityStreamsType ActivityStreamsObject
// // Original text of the status without formatting
// Text string

// // Mentions created in this status
// GTSMentions []*Mention `pg:"-"`
// // Hashtags used in this status
// GTSTags []*Tag `pg:"-"`
// // Emojis used in this status
// GTSEmojis []*Emoji `pg:"-"`
// // MediaAttachments used in this status
// GTSMediaAttachments []*MediaAttachment `pg:"-"`
// // Status being replied to
// GTSReplyToStatus *Status `pg:"-"`
// // Account being replied to
// GTSReplyToAccount *Account `pg:"-"`
// // Status being boosted
// GTSBoostedStatus *Status `pg:"-"`
// // Account of the boosted status
// GTSBoostedAccount *Account `pg:"-"`
