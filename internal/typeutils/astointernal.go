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
	"net/url"
	"strings"

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

	// uri at which this status is reachable
	uriProp := statusable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found, or id was not an iri")
	}
	status.URI = uriProp.GetIRI().String()

	// web url for viewing this status
	if statusURL, err := extractURL(statusable); err == nil {
		status.URL = statusURL.String()
	}

	// the html-formatted content of this status
	if content, err := extractContent(statusable); err == nil {
		status.Content = content
	}

	// attachments to dereference and fetch later on (we don't do that here)
	if attachments, err := extractAttachments(statusable); err == nil {
		status.GTSMediaAttachments = attachments
	}

	// hashtags to dereference later on
	if hashtags, err := extractHashtags(statusable); err == nil {
		status.GTSTags = hashtags
	}

	// emojis to dereference and fetch later on
	if emojis, err := extractEmojis(statusable); err == nil {
		status.GTSEmojis = emojis
	}

	// mentions to dereference later on
	if mentions, err := extractMentions(statusable); err == nil {
		status.GTSMentions = mentions
	}

	// cw string for this status
	if cw, err := extractSummary(statusable); err == nil {
		status.ContentWarning = cw
	}

	// when was this status created?
	published, err := extractPublished(statusable)
	if err == nil {
		status.CreatedAt = published
	}

	// which account posted this status?
	// if we don't know the account yet we can dereference it later
	attributedTo, err := extractAttributedTo(statusable)
	if err != nil {
		return nil, errors.New("attributedTo was empty")
	}
	status.APStatusOwnerURI = attributedTo.String()

	statusOwner := &gtsmodel.Account{}
	if err := c.db.GetWhere("uri", attributedTo.String(), statusOwner); err != nil {
		return nil, fmt.Errorf("couldn't get status owner from db: %s", err)
	}
	status.AccountID = statusOwner.ID
	status.GTSAccount = statusOwner

	// check if there's a post that this is a reply to
	inReplyToURI, err := extractInReplyToURI(statusable)
	if err == nil {
		// something is set so we can at least set this field on the
		// status and dereference using this later if we need to
		status.APReplyToStatusURI = inReplyToURI.String()

		// now we can check if we have the replied-to status in our db already
		inReplyToStatus := &gtsmodel.Status{}
		if err := c.db.GetWhere("uri", inReplyToURI.String(), inReplyToStatus); err == nil {
			// we have the status in our database already
			// so we can set these fields here and then...
			status.InReplyToID = inReplyToStatus.ID
			status.InReplyToAccountID = inReplyToStatus.AccountID
			status.GTSReplyToStatus = inReplyToStatus

			// ... check if we've seen the account already
			inReplyToAccount := &gtsmodel.Account{}
			if err := c.db.GetByID(inReplyToStatus.AccountID, inReplyToAccount); err == nil {
				status.GTSReplyToAccount = inReplyToAccount
			}
		}
	}

	// visibility entry for this status
	var visibility gtsmodel.Visibility

	to, err := extractTos(statusable)
	if err != nil {
		return nil, fmt.Errorf("error extracting TO values: %s", err)
	}

	cc, err := extractCCs(statusable)
	if err != nil {
		return nil, fmt.Errorf("error extracting CC values: %s", err)
	}

	if len(to) == 0 && len(cc) == 0 {
		return nil, errors.New("message wasn't TO or CC anyone")
	}

	// for visibility derivation, we start by assuming most restrictive, and work our way to least restrictive

	// if it's a DM then it's addressed to SPECIFIC ACCOUNTS and not followers or public
	if len(to) != 0 && len(cc) == 0 {
		visibility = gtsmodel.VisibilityDirect
	}

	// if it's just got followers in TO and it's not also CC'ed to public, it's followers only
	if isFollowers(to, statusOwner.FollowersURI) {
		visibility = gtsmodel.VisibilityFollowersOnly
	}

	// if it's CC'ed to public, it's public or unlocked
	// mentioned SPECIFIC ACCOUNTS also get added to CC'es if it's not a direct message
	if isPublic(cc) || isPublic(to) {
		visibility = gtsmodel.VisibilityPublic
	}

	// we should have a visibility by now
	if visibility == "" {
		return nil, errors.New("couldn't derive visibility")
	}
	status.Visibility = visibility

	// advanced visibility for this status
	// TODO: a lot of work to be done here -- a new type needs to be created for this in go-fed/activity using ASTOOL

	// sensitive
	// TODO: this is a bool

	// language
	// we might be able to extract this from the contentMap field

	// ActivityStreamsType
	status.ActivityStreamsType = gtsmodel.ActivityStreamsObject(statusable.GetTypeName())

	return status, nil
}

func (c *converter) ASFollowToFollowRequest(followable Followable) (*gtsmodel.FollowRequest, error) {

	idProp := followable.GetJSONLDId()
	if idProp == nil || !idProp.IsIRI() {
		return nil, errors.New("no id property set on follow, or was not an iri")
	}
	uri := idProp.GetIRI().String()

	origin, err := extractActor(followable)
	if err != nil {
		return nil, errors.New("error extracting actor property from follow")
	}
	originAccount := &gtsmodel.Account{}
	if err := c.db.GetWhere("uri", origin.String(), originAccount); err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	target, err := extractObject(followable)
	if err != nil {
		return nil, errors.New("error extracting object property from follow")
	}
	targetAccount := &gtsmodel.Account{}
	if err := c.db.GetWhere("uri", target.String(), targetAccount); err != nil {
		return nil, fmt.Errorf("error extracting account with uri %s from the database: %s", origin.String(), err)
	}

	followRequest := &gtsmodel.FollowRequest{
		URI:             uri,
		AccountID:       originAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	return followRequest, nil
}

func isPublic(tos []*url.URL) bool {
	for _, entry := range tos {
		if strings.EqualFold(entry.String(), "https://www.w3.org/ns/activitystreams#Public") {
			return true
		}
	}
	return false
}

func isFollowers(ccs []*url.URL, followersURI string) bool {
	for _, entry := range ccs {
		if strings.EqualFold(entry.String(), followersURI) {
			return true
		}
	}
	return false
}
